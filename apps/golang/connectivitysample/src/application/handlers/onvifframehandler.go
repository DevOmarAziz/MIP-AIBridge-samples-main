package handlers

import (
	services "connectivitysample/src/application/services"
	entities "connectivitysample/src/domain/entities"
	"connectivitysample/src/domain/enums"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

// Handles all requests coming to the '/onvifframe' endpoint.
type OnvifFrameHandler struct {
	queryStringService    *services.QueryStringService
	TopicRestService      *services.TopicRestService
	commandLineParameters *entities.CommandLineParameters
	fileReader            *services.FileReader
}

func NewOnvifFrameHandler(queryStringService *services.QueryStringService,
	TopicRestService *services.TopicRestService, commandLineParameters *entities.CommandLineParameters) *OnvifFrameHandler {
	return &OnvifFrameHandler{
		queryStringService:    queryStringService,
		TopicRestService:      TopicRestService,
		commandLineParameters: commandLineParameters,
		fileReader:            services.NewFileReader("onvif-frame", "xml"),
	}
}

// Renders 'onvif-frame-metadata.html' when '/onvifframe' endpoint gets requested. (commonly from MC)
// This page:
// 1 - Allows a MC user to start or stop sending ONVIF Frame metadata to a certain camera.
// 2 - Can be loaded in MC> Recording Servers (node) -> Select a camera -> select 'properties' tab -> select 'Processing server' tab -> select 'sendframe' topic from the treeview.

func (oh *OnvifFrameHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// The path can include a device id and stream id. If so, we extract it.
	// URL to load on the processing-server tab from a certain cameraID: https://${EXTERNAL_HOSTNAME}:7443/onvifframe/sendframe/camID/streamIDcamID
	queryStringContext, err := oh.queryStringService.ExtractTopicNameCameraAndStreamIDsFromPath("/onvifframe/(.+)/(.+?)/.*/(.+)", r, w)
	if err != nil {
		// The path didn't include device id and stream id.
		// Since we didn't implement a 'configuration' page for this topic, we just return.
		return
	}

	topicName := queryStringContext.TopicName
	cameraID := queryStringContext.CameraID
	streamID := queryStringContext.StreamID

	// return the onvif-metadata-page.html
	path := "templates/onvif-frame-metadata-page.html"
	template, err := template.ParseFS(templateFS, path)
	if err != nil {
		log.Printf("Error parsing template file %s: %v", path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// data to be passed to the template
	pageData := struct {
		CameraID   string
		StreamID   string
		TopicName  string
		Status     string
		AppUrlPath string
	}{
		CameraID:   cameraID,
		StreamID:   streamID,
		TopicName:  topicName,
		Status:     sendStatus,
		AppUrlPath: oh.commandLineParameters.AppUrlPath(),
	}
	// write to the response
	template.Execute(w, pageData)
}

// Handles the requests coming to the '/metadata/processing/sendframe' endpoint
// This happens when the user clicks the 'send' button in the MC.
// As a result, the app sends the ONVIF Frame metadata payloads on demand.
func (oh *OnvifFrameHandler) ProcessingHandle(w http.ResponseWriter, r *http.Request) {
	// Request body coming in
	var eventData struct {
		CameraID  string `json:"cameraId"`
		StreamID  string `json:"streamId"`
		TopicName string `json:"topicName"`
	}
	err := json.NewDecoder(r.Body).Decode(&eventData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Response body coming out
	var eventDataResponse struct {
		EventStatus string `json:"EventStatus"`
	}

	log.Printf("Sending metadata")
	xmls, err := oh.fileReader.ReadMultipleFiles()
	if err != nil {
		log.Printf("Error getting the ONVIF Frame metadata content: %s", err)
		http.Error(w, "Failed to load ONVIF Frame metadata", http.StatusInternalServerError)
		return
	}
	err = oh.TopicRestService.SendData(eventData.CameraID, eventData.StreamID, eventData.TopicName, enums.Metadata, "xml", xmls)
	if err != nil {
		log.Printf("Error sending the ONVIF Frame metadata: %s", err)
		http.Error(w, "Failed to send ONVIF Frame metadata", http.StatusBadGateway)
		return
	}
	eventDataResponse.EventStatus = sendStatus

	// Respond to the client (Button Start or Stop text in MC)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventDataResponse)
}
