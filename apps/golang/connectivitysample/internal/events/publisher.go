package events

import (
	"context"
	"fmt"
	"log"
	"time"
)

const (
	defaultQueueSize  = 256
	defaultRetryDelay = 500 * time.Millisecond
	contentTypeJSON   = "application/json"
)

// Publisher publishes analytics events into the VMS.
type Publisher interface {
	Publish(event AnalyticsEvent) error
}

// EndpointResolver resolves the REST endpoint for a specific VMS topic.
type EndpointResolver func(ctx context.Context) (string, error)

// Sender is the transport used by the publisher.
type Sender interface {
	Send(url string, payload []byte, contentType string) error
}

// Config holds the dependencies required to publish analytics events.
type Config struct {
	QueueSize    int
	RetryDelay   time.Duration
	ResolveTopic EndpointResolver
	Sender       Sender
	Logger       *log.Logger
}

// AsyncPublisher provides ordered, asynchronous event delivery.
type AsyncPublisher struct {
	queue        chan AnalyticsEvent
	resolveTopic EndpointResolver
	sender       Sender
	retryDelay   time.Duration
	logger       *log.Logger
}

func NewPublisher(config Config) (*AsyncPublisher, error) {
	if config.ResolveTopic == nil {
		return nil, fmt.Errorf("resolve topic callback must be provided")
	}
	if config.Sender == nil {
		return nil, fmt.Errorf("sender must be provided")
	}

	queueSize := config.QueueSize
	if queueSize <= 0 {
		queueSize = defaultQueueSize
	}

	retryDelay := config.RetryDelay
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}

	publisher := &AsyncPublisher{
		queue:        make(chan AnalyticsEvent, queueSize),
		resolveTopic: config.ResolveTopic,
		sender:       config.Sender,
		retryDelay:   retryDelay,
		logger:       config.Logger,
	}

	go publisher.run()

	return publisher, nil
}

func (publisher *AsyncPublisher) Publish(event AnalyticsEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}

	select {
	case publisher.queue <- event:
		return nil
	default:
		return fmt.Errorf("analytics event queue is full")
	}
}

func (publisher *AsyncPublisher) run() {
	for event := range publisher.queue {
		for {
			if err := publisher.publish(event); err != nil {
				publisher.logf("failed publishing analytics event %s for camera %s: %v", event.EventID, event.CameraID, err)
				time.Sleep(publisher.retryDelay)
				continue
			}

			break
		}
	}
}

func (publisher *AsyncPublisher) publish(event AnalyticsEvent) error {
	payload, err := event.toPayload()
	if err != nil {
		return err
	}

	endpoint, err := publisher.resolveTopic(context.Background())
	if err != nil {
		return err
	}

	return publisher.sender.Send(endpoint, payload, contentTypeJSON)
}

func (publisher *AsyncPublisher) logf(format string, args ...any) {
	if publisher.logger == nil {
		return
	}

	publisher.logger.Printf(format, args...)
}
