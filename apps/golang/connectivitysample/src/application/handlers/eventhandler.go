package handlers

import (
	services "connectivitysample/src/application/services"
	entities "connectivitysample/src/domain/entities"
	"connectivitysample/src/domain/enums"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

var sendStatus = "Send"

// Handles all requests coming to the '/event' endpoint.
type EventHandler struct {
	queryStringService    *services.QueryStringService
	TopicRestService      *services.TopicRestService
	commandLineParameters *entities.CommandLineParameters
	fileReader            *services.FileReader
}

func NewEventHandler(queryStringService *services.QueryStringService,
	TopicRestService *services.TopicRestService, commandLineParameters *entities.CommandLineParameters) *EventHandler {
	return &EventHandler{
		queryStringService:    queryStringService,
		TopicRestService:      TopicRestService,
		commandLineParameters: commandLineParameters,
		fileReader:            services.NewFileReader("analyticEvent", "json"),
	}
}

// Renders 'event-camera-page.html' when '/event' endpoint gets requested. (commonly from MC)
// This page:
// 1 - Allows a MC user to start or stop sending analytic events from a certain camera.
// 2 - Can be loaded in MC> Recording Servers (node) -> Select a camera -> select 'properties' tab -> select 'Processing server' tab -> select 'sendanalyticevents' topic from the treeview.

func (eh *EventHandler) Handle(w http.ResponseWriter, r *http.Request) {

	// The path can include a device id and stream id. If so, we extract it.
	// URL to load on the processing-server tab from a certain cameraID: https://${EXTERNAL_HOSTNAME}:7443/event/sendanalyticevents/camID/streamIDcamID
	queryStringContext, err := eh.queryStringService.ExtractTopicNameCameraAndStreamIDsFromPath("/event/(.+)/(.+?)/.*/(.+)", r, w)
	if err != nil {
		// The path didn't include device id and stream id.
		// Since we didn't implement a 'configuration' page for this topic, we just return.
		return
	}

	topicName := queryStringContext.TopicName
	cameraID := queryStringContext.CameraID
	streamID := queryStringContext.StreamID

	// return the event-camera-page.html
	path := "templates/event-camera-page.html"
	template, err := template.ParseFS(templateFS, path)
	if err != nil {
		log.Printf("Error parsing template file %s: %v", path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// data to be passed to the template
	pageData := struct {
		CameraID   string
		StreamID   string
		TopicName  string
		Status     string
		AppUrlPath string
	}{
		CameraID:   cameraID,
		StreamID:   streamID,
		TopicName:  topicName,
		Status:     sendStatus,
		AppUrlPath: eh.commandLineParameters.AppUrlPath(),
	}
	// write to the response
	template.Execute(w, pageData)
}

// Handles the requests coming to the '/event/processing/sendanalyticevents' endpoint
// This happens when the user clicks the 'send' button in the MC.
// As a result, the app sends one analytic event on demand.
func (eh *EventHandler) ProcessingHandle(w http.ResponseWriter, r *http.Request) {
	// Request body coming in
	var eventData struct {
		CameraID  string `json:"cameraId"`
		TopicName string `json:"topicName"`
	}
	err := json.NewDecoder(r.Body).Decode(&eventData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Response body coming out
	var eventDataResponse struct {
		EventStatus string `json:"EventStatus"`
	}

	eventPayload, err := eh.fileReader.ReadSingleFile()
	if err != nil {
		log.Printf("Error getting the analytic events: %s", err)
		http.Error(w, "Failed to load analytic event", http.StatusInternalServerError)
		return
	}
	err = eh.TopicRestService.SendData(eventData.CameraID, eventData.TopicName, eventData.TopicName, enums.Event, "json", []string{eventPayload})
	if err != nil {
		log.Printf("Error sending the analytic event: %s", err)
		http.Error(w, "Failed to send analytic event", http.StatusBadGateway)
		return
	}
	eventDataResponse.EventStatus = sendStatus

	// Respond to the client (Button Start or Stop text in MC)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventDataResponse)
}
