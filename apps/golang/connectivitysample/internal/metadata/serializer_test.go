package metadata

import (
	"strings"
	"testing"
	"time"
)

func TestONVIFAnalyticsFrameSerializerUsesFrameTimestampAndPixelBoundingBoxes(t *testing.T) {
	frame := MetadataFrame{
		Timestamp:      time.Date(2026, 5, 18, 10, 11, 12, 345678900, time.UTC),
		SourceStreamID: "camera-1/stream-1",
		Objects: []MetadataObject{
			{
				TrackingID: 7,
				ObjectType: "Person",
				Confidence: 0.93,
				Box: BoundingBox{
					Left:   100,
					Top:    20,
					Right:  240,
					Bottom: 300,
				},
				Attributes: map[string]string{
					"zone": "entry",
				},
			},
		},
	}

	payload, err := NewONVIFAnalyticsFrameSerializer().Serialize(frame)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	xmlBody := string(payload)
	for _, part := range []string{
		`<Frame xmlns="http://www.onvif.org/ver10/schema" UtcTime="2026-05-18T10:11:12.3456789Z" SourceStreamID="camera-1/stream-1">`,
		`<Object ObjectId="7">`,
		`<Type>Person</Type>`,
		`<BoundingBox left="100" top="20" right="240" bottom="300"></BoundingBox>`,
		`<Property name="zone">entry</Property>`,
	} {
		if !strings.Contains(xmlBody, part) {
			t.Fatalf("serialized XML missing %q in %s", part, xmlBody)
		}
	}
}
