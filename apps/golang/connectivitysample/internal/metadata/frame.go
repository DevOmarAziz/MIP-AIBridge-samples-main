package metadata

import (
	"fmt"
	"time"
)

// MetadataFrame is the transport model passed from an external analytics
// producer into the bridge.
type MetadataFrame struct {
	Timestamp      time.Time
	SourceStreamID string
	Objects        []MetadataObject
}

func (frame MetadataFrame) Validate() error {
	if frame.Timestamp.IsZero() {
		return fmt.Errorf("timestamp must be provided")
	}
	if frame.SourceStreamID == "" {
		return fmt.Errorf("sourceStreamID must be provided")
	}

	for index, object := range frame.Objects {
		if err := object.Validate(); err != nil {
			return fmt.Errorf("object %d is invalid: %w", index, err)
		}
	}

	return nil
}
