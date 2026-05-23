package metadata

import "fmt"

// BoundingBox uses pixel coordinates from the source frame.
type BoundingBox struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func (box BoundingBox) Validate() error {
	if box.Right < box.Left {
		return fmt.Errorf("bounding box right must be greater than or equal to left")
	}
	if box.Bottom < box.Top {
		return fmt.Errorf("bounding box bottom must be greater than or equal to top")
	}

	return nil
}

// MetadataObject represents a single object emitted by the external analytics
// layer for a frame.
type MetadataObject struct {
	TrackingID int
	ObjectType string
	Confidence float32
	Box        BoundingBox
	Attributes map[string]string
}

func (object MetadataObject) Validate() error {
	if object.TrackingID < 0 {
		return fmt.Errorf("trackingID must be zero or greater")
	}
	if object.ObjectType == "" {
		return fmt.Errorf("objectType must be provided")
	}

	return object.Box.Validate()
}
