package events

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AnalyticsEvent is the transport model that external analytics systems pass
// to the bridge for publication into the VMS.
type AnalyticsEvent struct {
	EventID    string
	CameraID   string
	StreamID   string
	Timestamp  time.Time
	Type       AnalyticsEventType
	Confidence float32
	Attributes map[string]string
}

func (event AnalyticsEvent) Validate() error {
	switch {
	case strings.TrimSpace(event.EventID) == "":
		return fmt.Errorf("eventID must be provided")
	case strings.TrimSpace(event.CameraID) == "":
		return fmt.Errorf("cameraID must be provided")
	case event.Timestamp.IsZero():
		return fmt.Errorf("timestamp must be provided")
	case strings.TrimSpace(string(event.Type)) == "":
		return fmt.Errorf("event type must be provided")
	}

	return nil
}

func (event AnalyticsEvent) toPayload() ([]byte, error) {
	if err := event.Validate(); err != nil {
		return nil, err
	}

	attributes := cloneAttributes(event.Attributes)
	if event.StreamID != "" {
		attributes["streamId"] = event.StreamID
	}
	attributes["eventId"] = event.EventID
	attributes["timestamp"] = event.Timestamp.Format(time.RFC3339Nano)

	payload := analyticsEventPayload{
		ID:          event.EventID,
		Name:        string(event.Type),
		Description: fmt.Sprintf("%s reported by external analytics", event.Type),
		Class:       string(event.Type),
		Subclass:    string(event.Type),
		Count:       1,
		Tag:         string(event.Type),
		FromSource: payloadReference{
			Type: "Reference",
			UUID: event.CameraID,
		},
		RelatedTo: []payloadReference{
			{
				Type: "Reference",
				UUID: event.CameraID,
			},
		},
		Attributes: attributes,
	}

	if event.Type.includesObject() {
		payload.InvolvedObject = []payloadObject{
			{
				Type:        "Object",
				Name:        string(event.Type),
				Description: fmt.Sprintf("%s detected", event.Type.defaultObjectType()),
				Class:       event.Type.defaultObjectType(),
				Confidence:  event.Confidence,
				Readout:     event.StreamID,
			},
		}
	}

	return json.Marshal(payload)
}

func cloneAttributes(attributes map[string]string) map[string]string {
	if len(attributes) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(attributes))
	for key, value := range attributes {
		cloned[key] = value
	}

	return cloned
}

type analyticsEventPayload struct {
	ID             string             `json:"id,omitempty"`
	Name           string             `json:"name"`
	Description    string             `json:"description,omitempty"`
	Class          string             `json:"class"`
	Subclass       string             `json:"subclass,omitempty"`
	Count          int                `json:"count"`
	Tag            string             `json:"tag,omitempty"`
	FromSource     payloadReference   `json:"fromSource"`
	RelatedTo      []payloadReference `json:"relatedTo,omitempty"`
	InvolvedObject []payloadObject    `json:"involvedObject,omitempty"`
	Attributes     map[string]string  `json:"attributes,omitempty"`
}

type payloadReference struct {
	Type string `json:"type"`
	UUID string `json:"uuid"`
}

type payloadObject struct {
	Type        string  `json:"type"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Class       string  `json:"class,omitempty"`
	Confidence  float32 `json:"confidence,omitempty"`
	Readout     string  `json:"readout,omitempty"`
}
