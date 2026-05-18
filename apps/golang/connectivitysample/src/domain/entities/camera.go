package entities

type Camera struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	VideoStreams []VideoStream `json:"videoStreams"`
}

type VideoStream struct {
	ID string `json:"id"`
}
