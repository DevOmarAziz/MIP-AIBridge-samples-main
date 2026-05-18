package services

import (
	entities "connectivitysample/src/domain/entities"
	"connectivitysample/src/domain/enums"
	"connectivitysample/src/infrastructure/repositories"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type TopicRestService struct {
	dataBeingSent  *sync.Map
	graphqlService *GraphqlService
}

func NewTopicRestService(dataBeingSent *sync.Map, graphqlService *GraphqlService) *TopicRestService {
	return &TopicRestService{
		dataBeingSent:  dataBeingSent,
		graphqlService: graphqlService,
	}
}

// Public methods

// GetCameras returns cameras with ID and name from the GraphQL API.
func (ts *TopicRestService) GetCameras(ctx context.Context) ([]entities.Camera, error) {
	return ts.graphqlService.GetCameras(ctx)
}

// SendAnalyticsEvent sends an analytics event payload for a specific camera.
func (ts *TopicRestService) SendAnalyticsEvent(cameraID string, topicName string, payload []byte) error {
	if topicName == "" {
		return fmt.Errorf("topicName must be provided")
	}

	topicRestUrl, err := ts.graphqlService.GetRestEventTopicEndpoint(context.Background(), topicName)
	if err != nil {
		return err
	}

	currentPayload := TreatEventFile(string(payload), cameraID)
	return repositories.SendPostRequest(topicRestUrl, currentPayload, "application/json")
}

// SendMetadata sends a metadata payload for a specific camera.
func (ts *TopicRestService) SendMetadata(cameraID string, streamID string, topicName string, payload []byte, contentType string) error {
	if topicName == "" {
		return fmt.Errorf("topicName must be provided")
	}

	topicRestUrl, err := ts.graphqlService.GetRestMetadataTopicEndpoint(context.Background(), topicName)
	if err != nil {
		return err
	}

	if contentType == "" {
		contentType = "application/json"
	}

	sourceStreamID := cameraID
	if streamID != "" {
		sourceStreamID = cameraID + "/" + streamID
	}

	currentPayload := TreatMetadataFile(string(payload), sourceStreamID)
	return repositories.SendPostRequest(topicRestUrl, currentPayload, contentType)
}

// Check if the given cameraID is being used for sending data
func (ts *TopicRestService) IsDataBeingSent(cameraID string) bool {

	_, found := ts.dataBeingSent.Load(cameraID)

	return found
}

// Sends the provided event or metadata payloads once, in order.
func (ts *TopicRestService) SendData(cameraID string, streamID string, topicName string, topicFormat int, fileFormat string, files []string) error {
	sourceStreamID := cameraID + "/" + streamID
	log.Printf("Sending data manually for the SourceStreamID %s", sourceStreamID)

	topicRestUrl := ""
	var err error

	switch topicFormat {
	case enums.Event:
		topicRestUrl, err = ts.graphqlService.GetRestEventTopicEndpoint(context.Background(), topicName)
	case enums.Metadata:
		topicRestUrl, err = ts.graphqlService.GetRestMetadataTopicEndpoint(context.Background(), topicName)
	default:
		return fmt.Errorf("unsupported topic format: %d", topicFormat)
	}

	if err != nil {
		return err
	}

	for _, file := range files {
		currentFile := ""
		switch topicFormat {
		case enums.Event:
			currentFile = TreatEventFile(file, cameraID)
		case enums.Metadata:
			currentFile = TreatMetadataFile(file, sourceStreamID)
		}

		err = repositories.SendPostRequest(topicRestUrl, currentFile, getTopicContentType(fileFormat))
		if err != nil {
			return err
		}
	}

	log.Printf("Successfully sent %d payload(s) to topic %s via %s", len(files), topicName, topicRestUrl)
	return nil
}

// Start sending data to a given cameraID every 1 second (fire & forget)
func (ts *TopicRestService) SendDataAsync(cameraID string, streamID string, topicName string, topicFormat int, fileFormat string, files []string) {
	// Add or Update the cameraID in the map
	ts.dataBeingSent.Store(cameraID, "running")
	sourceStreamID := cameraID + "/" + streamID
	log.Printf("Sending data for the SourceStreamID %s", sourceStreamID)

	go func() {

		// data will be sent every 1 seconds.
		ticker := time.NewTicker(1 * time.Second)

		topicRestUrl := ""
		var err error

		switch topicFormat {
		case enums.Event:
			topicRestUrl, err = ts.graphqlService.GetRestEventTopicEndpoint(context.Background(), topicName)
		case enums.Metadata:
			topicRestUrl, err = ts.graphqlService.GetRestMetadataTopicEndpoint(context.Background(), topicName)
		}

		if err != nil {
			log.Printf("Error getting the topic rest url: %s", err)
			return
		}

		index := 0

		for stopSendingData := false; !stopSendingData; {
			// wait for the next tick
			<-ticker.C

			// Check if the data is still being sent
			found := ts.IsDataBeingSent(cameraID)
			if !found {
				stopSendingData = true
				return
			}

			// Alternate between all the loaded files
			currentFile := ""
			switch topicFormat {
			case enums.Event:
				currentFile = TreatEventFile(files[index%len(files)], cameraID)
			case enums.Metadata:
				currentFile = TreatMetadataFile(files[index%len(files)], sourceStreamID)
			}
			index++

			// Publish data into Milestone AI Bridge.
			err := repositories.SendPostRequest(topicRestUrl, currentFile, getTopicContentType(fileFormat))
			if err != nil {
				log.Printf("Couldn't publish the data: %v", err)
			} else {
				log.Printf("Successfully sent payload to topic %s via %s", topicName, topicRestUrl)
			}
		}
	}()
}

// Stop sending data related to a certain cameraID
func (ts *TopicRestService) StopSendingData(cameraID string) {

	log.Printf("Stopping sending data related to the cameraID %s", cameraID)
	ts.dataBeingSent.Delete(cameraID)
}

func getTopicContentType(fileFormat string) string {
	switch fileFormat {
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	default:
		return "text/" + fileFormat
	}
}
