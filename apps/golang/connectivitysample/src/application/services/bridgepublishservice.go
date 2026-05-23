package services

import (
	bridgeevents "connectivitysample/internal/events"
	bridgemetadata "connectivitysample/internal/metadata"
	"connectivitysample/src/infrastructure/repositories"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	defaultPublisherQueueSize = 256
	defaultPublisherRetryWait = 500 * time.Millisecond
)

// BridgePublisherService exposes structured publishing APIs that external
// analytics producers can call through the application layer.
type BridgePublisherService struct {
	graphqlService     *GraphqlService
	metadataTopics     *bridgemetadata.TopicRegistry
	eventPublishers    sync.Map
	metadataPublishers sync.Map
}

func NewBridgePublisherService(graphqlService *GraphqlService) *BridgePublisherService {
	return &BridgePublisherService{
		graphqlService: graphqlService,
		metadataTopics: bridgemetadata.NewTopicRegistry(),
	}
}

func (service *BridgePublisherService) RegisterMetadataTopic(topic bridgemetadata.MetadataTopic) error {
	return service.metadataTopics.RegisterTopic(topic)
}

func (service *BridgePublisherService) AssociateMetadataStream(topicName, cameraID, streamID string) error {
	return service.metadataTopics.AssociateStream(topicName, cameraID, streamID)
}

func (service *BridgePublisherService) MapSourceStream(cameraID, streamID, sourceStreamID string) {
	service.metadataTopics.MapSourceStream(cameraID, streamID, sourceStreamID)
}

func (service *BridgePublisherService) PublishAnalyticsEvent(topicName string, event bridgeevents.AnalyticsEvent) error {
	publisher, err := service.eventPublisher(topicName)
	if err != nil {
		return err
	}

	return publisher.Publish(event)
}

func (service *BridgePublisherService) PublishMetadataFrame(topicName string, frame bridgemetadata.MetadataFrame) error {
	if _, found := service.metadataTopics.Topic(topicName); !found {
		if err := service.metadataTopics.RegisterTopic(bridgemetadata.NewMetadataTopic(topicName)); err != nil {
			return err
		}
	}

	publisher, err := service.metadataPublisher(topicName)
	if err != nil {
		return err
	}

	return publisher.Publish(frame)
}

func (service *BridgePublisherService) PublishMetadataFrameForStream(topicName, cameraID, streamID string, frame bridgemetadata.MetadataFrame) error {
	if frame.SourceStreamID == "" {
		frame.SourceStreamID = service.metadataTopics.ResolveOrBuildSourceStreamID(cameraID, streamID)
	}

	if cameraID != "" {
		if _, found := service.metadataTopics.Topic(topicName); !found {
			if err := service.metadataTopics.RegisterTopic(bridgemetadata.NewMetadataTopic(topicName)); err != nil {
				return err
			}
		}

		if err := service.metadataTopics.AssociateStream(topicName, cameraID, streamID); err != nil {
			return err
		}
	}

	return service.PublishMetadataFrame(topicName, frame)
}

func (service *BridgePublisherService) eventPublisher(topicName string) (bridgeevents.Publisher, error) {
	if topicName == "" {
		return nil, fmt.Errorf("topicName must be provided")
	}

	if existing, found := service.eventPublishers.Load(topicName); found {
		return existing.(bridgeevents.Publisher), nil
	}

	publisher, err := bridgeevents.NewPublisher(bridgeevents.Config{
		QueueSize:  defaultPublisherQueueSize,
		RetryDelay: defaultPublisherRetryWait,
		ResolveTopic: func(ctx context.Context) (string, error) {
			return service.graphqlService.GetRestEventTopicEndpoint(ctx, topicName)
		},
		Sender: bridgeRestSender{},
		Logger: log.Default(),
	})
	if err != nil {
		return nil, err
	}

	actual, loaded := service.eventPublishers.LoadOrStore(topicName, publisher)
	if loaded {
		return actual.(bridgeevents.Publisher), nil
	}

	return publisher, nil
}

func (service *BridgePublisherService) metadataPublisher(topicName string) (bridgemetadata.MetadataPublisher, error) {
	if topicName == "" {
		return nil, fmt.Errorf("topicName must be provided")
	}

	if existing, found := service.metadataPublishers.Load(topicName); found {
		return existing.(bridgemetadata.MetadataPublisher), nil
	}

	publisher, err := bridgemetadata.NewPublisher(bridgemetadata.Config{
		QueueSize:  defaultPublisherQueueSize,
		RetryDelay: defaultPublisherRetryWait,
		ResolveTopic: func(ctx context.Context) (string, error) {
			return service.graphqlService.GetRestMetadataTopicEndpoint(ctx, topicName)
		},
		Sender:     bridgeRestSender{},
		Serializer: bridgemetadata.NewONVIFAnalyticsFrameSerializer(),
		Logger:     log.Default(),
	})
	if err != nil {
		return nil, err
	}

	actual, loaded := service.metadataPublishers.LoadOrStore(topicName, publisher)
	if loaded {
		return actual.(bridgemetadata.MetadataPublisher), nil
	}

	return publisher, nil
}

type bridgeRestSender struct{}

func (bridgeRestSender) Send(url string, payload []byte, contentType string) error {
	return repositories.SendPostRequest(url, string(payload), contentType)
}
