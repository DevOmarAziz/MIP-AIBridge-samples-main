package metadata

import "testing"

func TestTopicRegistryAssociatesTopicsAndSourceStreams(t *testing.T) {
	registry := NewTopicRegistry()
	topic := NewMetadataTopic("sendframe")

	if err := registry.RegisterTopic(topic); err != nil {
		t.Fatalf("RegisterTopic() error = %v", err)
	}

	registry.MapSourceStream("camera-1", "stream-1", "source/camera-1/stream-1")

	if err := registry.AssociateStream("sendframe", "camera-1", "stream-1"); err != nil {
		t.Fatalf("AssociateStream() error = %v", err)
	}

	if got, want := registry.ResolveOrBuildSourceStreamID("camera-1", "stream-1"), "source/camera-1/stream-1"; got != want {
		t.Fatalf("ResolveOrBuildSourceStreamID() = %s, want %s", got, want)
	}

	resolvedTopic, found := registry.TopicForStream("camera-1", "stream-1")
	if !found {
		t.Fatal("TopicForStream() did not find registered topic")
	}
	if resolvedTopic.TopicName != topic.TopicName {
		t.Fatalf("TopicForStream() topic = %s, want %s", resolvedTopic.TopicName, topic.TopicName)
	}
}

func TestBuildSourceStreamID(t *testing.T) {
	if got, want := BuildSourceStreamID("camera-1", "stream-1"), "camera-1/stream-1"; got != want {
		t.Fatalf("BuildSourceStreamID() = %s, want %s", got, want)
	}

	if got, want := BuildSourceStreamID("camera-1", ""), "camera-1"; got != want {
		t.Fatalf("BuildSourceStreamID() = %s, want %s", got, want)
	}
}
