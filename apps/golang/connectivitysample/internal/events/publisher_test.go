package events

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestAsyncPublisherPublishesEventsInOrder(t *testing.T) {
	sender := &recordingSender{done: make(chan struct{}, 2)}
	publisher, err := NewPublisher(Config{
		QueueSize: 2,
		ResolveTopic: func(ctx context.Context) (string, error) {
			return "http://bridge/topic", nil
		},
		Sender: sender,
	})
	if err != nil {
		t.Fatalf("NewPublisher() error = %v", err)
	}

	events := []AnalyticsEvent{
		{
			EventID:   "evt-1",
			CameraID:  "camera-1",
			StreamID:  "stream-1",
			Timestamp: time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
			Type:      PersonDetected,
		},
		{
			EventID:   "evt-2",
			CameraID:  "camera-1",
			StreamID:  "stream-1",
			Timestamp: time.Date(2026, 5, 18, 10, 0, 1, 0, time.UTC),
			Type:      VehicleDetected,
		},
	}

	for _, event := range events {
		if err := publisher.Publish(event); err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	waitForRecordedPayloads(t, sender.done, len(events))

	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	if len(sender.payloads) != len(events) {
		t.Fatalf("expected %d payloads, got %d", len(events), len(sender.payloads))
	}

	for index, payload := range sender.payloads {
		var body map[string]any
		if err := json.Unmarshal(payload, &body); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if got, want := body["id"], events[index].EventID; got != want {
			t.Fatalf("payload %d id = %v, want %s", index, got, want)
		}
	}
}

func waitForRecordedPayloads(t *testing.T, done <-chan struct{}, count int) {
	t.Helper()

	timeout := time.After(2 * time.Second)
	for range count {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("timed out waiting for published payloads")
		}
	}
}

type recordingSender struct {
	mutex    sync.Mutex
	payloads [][]byte
	done     chan struct{}
}

func (sender *recordingSender) Send(url string, payload []byte, contentType string) error {
	sender.mutex.Lock()
	sender.payloads = append(sender.payloads, append([]byte(nil), payload...))
	sender.mutex.Unlock()

	sender.done <- struct{}{}
	return nil
}
