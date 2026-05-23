package metadata

import (
	"encoding/xml"
	"sort"
)

const onvifSchemaNamespace = "http://www.onvif.org/ver10/schema"

// Serializer converts metadata frames into a wire format accepted by the VMS.
type Serializer interface {
	Serialize(frame MetadataFrame) ([]byte, error)
}

// ONVIFAnalyticsFrameSerializer serializes metadata frames as
// ONVIF_ANALYTICS_FRAME XML.
type ONVIFAnalyticsFrameSerializer struct{}

func NewONVIFAnalyticsFrameSerializer() *ONVIFAnalyticsFrameSerializer {
	return &ONVIFAnalyticsFrameSerializer{}
}

func (serializer *ONVIFAnalyticsFrameSerializer) Serialize(frame MetadataFrame) ([]byte, error) {
	if err := frame.Validate(); err != nil {
		return nil, err
	}

	framePayload := onvifFrame{
		XMLNS:          onvifSchemaNamespace,
		UtcTime:        frame.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
		SourceStreamID: frame.SourceStreamID,
		Objects:        make([]onvifObject, 0, len(frame.Objects)),
	}

	for _, object := range frame.Objects {
		framePayload.Objects = append(framePayload.Objects, onvifObject{
			ObjectID: object.TrackingID,
			Appearance: onvifAppearance{
				Class: &onvifClass{
					ClassCandidate: onvifClassCandidate{
						Type:       object.ObjectType,
						Likelihood: object.Confidence,
					},
				},
				Shape: onvifShape{
					BoundingBox: onvifBoundingBox{
						Left:   object.Box.Left,
						Top:    object.Box.Top,
						Right:  object.Box.Right,
						Bottom: object.Box.Bottom,
					},
				},
				Extensions: buildAppearanceExtensions(object.Attributes),
			},
		})
	}

	return xml.MarshalIndent(framePayload, "", "    ")
}

func buildAppearanceExtensions(attributes map[string]string) []onvifExtension {
	if len(attributes) == 0 {
		return nil
	}

	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	properties := make([]onvifProperty, 0, len(keys))
	for _, key := range keys {
		properties = append(properties, onvifProperty{
			Name:  key,
			Value: attributes[key],
		})
	}

	return []onvifExtension{
		{
			Properties: &onvifProperties{
				Items: properties,
			},
		},
	}
}

type onvifFrame struct {
	XMLName        xml.Name      `xml:"Frame"`
	XMLNS          string        `xml:"xmlns,attr"`
	UtcTime        string        `xml:"UtcTime,attr"`
	SourceStreamID string        `xml:"SourceStreamID,attr,omitempty"`
	Objects        []onvifObject `xml:"Object"`
}

type onvifObject struct {
	ObjectID   int             `xml:"ObjectId,attr"`
	Appearance onvifAppearance `xml:"Appearance"`
}

type onvifAppearance struct {
	Class      *onvifClass       `xml:"Class,omitempty"`
	Shape      onvifShape        `xml:"Shape"`
	Extensions []onvifExtension  `xml:"Extension,omitempty"`
}

type onvifClass struct {
	ClassCandidate onvifClassCandidate `xml:"ClassCandidate"`
}

type onvifClassCandidate struct {
	Type       string  `xml:"Type"`
	Likelihood float32 `xml:"Likelihood,omitempty"`
}

type onvifShape struct {
	BoundingBox onvifBoundingBox `xml:"BoundingBox"`
}

type onvifBoundingBox struct {
	Left   int `xml:"left,attr"`
	Top    int `xml:"top,attr"`
	Right  int `xml:"right,attr"`
	Bottom int `xml:"bottom,attr"`
}

type onvifExtension struct {
	Properties *onvifProperties `xml:"Properties,omitempty"`
}

type onvifProperties struct {
	Items []onvifProperty `xml:"Property"`
}

type onvifProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}
