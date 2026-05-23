package metadata

import (
	"context"
	"fmt"
	"log"
	"time"
)

const (
	defaultQueueSize = 256
	defaultRetryWait = 500 * time.Millisecond
	contentTypeXML   = "application/xml"
)

// MetadataPublisher publishes metadata frames into the VMS.
type MetadataPublisher interface {
	Publish(frame MetadataFrame) error
}

// EndpointResolver resolves the REST endpoint for a metadata topic.
type EndpointResolver func(ctx context.Context) (string, error)

// Sender is the transport used by the metadata publisher.
type Sender interface {
	Send(url string, payload []byte, contentType string) error
}

// Config holds the dependencies required to publish metadata frames.
type Config struct {
	QueueSize    int
	RetryDelay   time.Duration
	ResolveTopic EndpointResolver
	Sender       Sender
	Serializer   Serializer
	Logger       *log.Logger
}

// AsyncPublisher provides ordered, asynchronous metadata delivery.
type AsyncPublisher struct {
	queue        chan MetadataFrame
	resolveTopic EndpointResolver
	sender       Sender
	serializer   Serializer
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
		retryDelay = defaultRetryWait
	}

	serializer := config.Serializer
	if serializer == nil {
		serializer = NewONVIFAnalyticsFrameSerializer()
	}

	publisher := &AsyncPublisher{
		queue:        make(chan MetadataFrame, queueSize),
		resolveTopic: config.ResolveTopic,
		sender:       config.Sender,
		serializer:   serializer,
		retryDelay:   retryDelay,
		logger:       config.Logger,
	}

	go publisher.run()

	return publisher, nil
}

func (publisher *AsyncPublisher) Publish(frame MetadataFrame) error {
	if err := frame.Validate(); err != nil {
		return err
	}

	select {
	case publisher.queue <- frame:
		return nil
	default:
		return fmt.Errorf("metadata frame queue is full")
	}
}

func (publisher *AsyncPublisher) run() {
	for frame := range publisher.queue {
		for {
			if err := publisher.publish(frame); err != nil {
				publisher.logf("failed publishing metadata frame for source stream %s: %v", frame.SourceStreamID, err)
				time.Sleep(publisher.retryDelay)
				continue
			}

			break
		}
	}
}

func (publisher *AsyncPublisher) publish(frame MetadataFrame) error {
	payload, err := publisher.serializer.Serialize(frame)
	if err != nil {
		return err
	}

	endpoint, err := publisher.resolveTopic(context.Background())
	if err != nil {
		return err
	}

	return publisher.sender.Send(endpoint, payload, contentTypeXML)
}

func (publisher *AsyncPublisher) logf(format string, args ...any) {
	if publisher.logger == nil {
		return
	}

	publisher.logger.Printf(format, args...)
}
