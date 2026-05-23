package events

// AnalyticsEventType identifies the category of event received from an
// external analytics producer.
type AnalyticsEventType string

const (
	PersonDetected  AnalyticsEventType = "PersonDetected"
	VehicleDetected AnalyticsEventType = "VehicleDetected"
	FaceDetected    AnalyticsEventType = "FaceDetected"
	LineCrossing    AnalyticsEventType = "LineCrossing"
	AnalyticsOffline AnalyticsEventType = "AnalyticsOffline"
)

func (t AnalyticsEventType) defaultObjectType() string {
	switch t {
	case PersonDetected:
		return "Person"
	case VehicleDetected:
		return "Vehicle"
	case FaceDetected:
		return "Face"
	default:
		return string(t)
	}
}

func (t AnalyticsEventType) includesObject() bool {
	return t != AnalyticsOffline
}
