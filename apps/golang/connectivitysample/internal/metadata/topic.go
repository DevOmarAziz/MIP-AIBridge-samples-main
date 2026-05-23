package metadata

import (
	"fmt"
	"strings"
	"sync"
)

const ONVIFAnalyticsFrameFormat = "ONVIF_ANALYTICS_FRAME"

// MetadataTopic describes a metadata topic registered in the bridge.
type MetadataTopic struct {
	TopicName      string
	MetadataFormat string
}

func NewMetadataTopic(topicName string) MetadataTopic {
	return MetadataTopic{
		TopicName:      topicName,
		MetadataFormat: ONVIFAnalyticsFrameFormat,
	}
}

// TopicRegistry keeps topic registrations and stream mappings together so an
// external producer can resolve the SourceStreamID consistently.
type TopicRegistry struct {
	mutex              sync.RWMutex
	topics             map[string]MetadataTopic
	streamAssociations map[string]string
	sourceStreamIDs    map[string]string
}

func NewTopicRegistry() *TopicRegistry {
	return &TopicRegistry{
		topics:             make(map[string]MetadataTopic),
		streamAssociations: make(map[string]string),
		sourceStreamIDs:    make(map[string]string),
	}
}

func (registry *TopicRegistry) RegisterTopic(topic MetadataTopic) error {
	if strings.TrimSpace(topic.TopicName) == "" {
		return fmt.Errorf("topic name must be provided")
	}
	if strings.TrimSpace(topic.MetadataFormat) == "" {
		return fmt.Errorf("metadata format must be provided")
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.topics[topic.TopicName] = topic
	return nil
}

func (registry *TopicRegistry) Topic(topicName string) (MetadataTopic, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	topic, found := registry.topics[topicName]
	return topic, found
}

func (registry *TopicRegistry) AssociateStream(topicName, cameraID, streamID string) error {
	if _, found := registry.Topic(topicName); !found {
		return fmt.Errorf("metadata topic %s is not registered", topicName)
	}

	sourceStreamID := registry.ResolveOrBuildSourceStreamID(cameraID, streamID)
	key := streamKey(cameraID, streamID)

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.streamAssociations[key] = topicName
	registry.sourceStreamIDs[key] = sourceStreamID
	return nil
}

func (registry *TopicRegistry) TopicForStream(cameraID, streamID string) (MetadataTopic, bool) {
	registry.mutex.RLock()
	topicName, found := registry.streamAssociations[streamKey(cameraID, streamID)]
	registry.mutex.RUnlock()
	if !found {
		return MetadataTopic{}, false
	}

	return registry.Topic(topicName)
}

func (registry *TopicRegistry) MapSourceStream(cameraID, streamID, sourceStreamID string) {
	key := streamKey(cameraID, streamID)

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.sourceStreamIDs[key] = sourceStreamID
}

func (registry *TopicRegistry) ResolveOrBuildSourceStreamID(cameraID, streamID string) string {
	key := streamKey(cameraID, streamID)

	registry.mutex.RLock()
	sourceStreamID, found := registry.sourceStreamIDs[key]
	registry.mutex.RUnlock()
	if found && sourceStreamID != "" {
		return sourceStreamID
	}

	return BuildSourceStreamID(cameraID, streamID)
}

func BuildSourceStreamID(cameraID, streamID string) string {
	if streamID == "" {
		return cameraID
	}

	return cameraID + "/" + streamID
}

func streamKey(cameraID, streamID string) string {
	return cameraID + "::" + streamID
}
