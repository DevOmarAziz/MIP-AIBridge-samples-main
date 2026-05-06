# Connectivity Sample App Source Flow

This file explains the code flow for `apps/golang/connectivitysample/src/` from request handling through GraphQL registration to event and metadata submission.

---

## 1. App startup and GraphQL registration

### main.go
- Parses CLI flags and builds runtime configuration:
  - `entities.NewTlsConfiguration(...)`
  - `entities.NewSnapshotConfiguration(...)`
  - `entities.NewCommandLineParameters(...)`
- Instantiates service layer objects:
  - `repositories.NewGraphqlRepository()`
  - `services.NewGraphqlService(queryURL)`
  - `services.NewTokenService(...)`
  - `services.NewQueryStringService()`
  - `services.NewTopicRestService(&sync.Map{}, graphqlService)`
- Creates handler objects:
  - `handlers.NewHomeHandler()`
  - `handlers.NewSnapshotHandler(...)`
  - `handlers.NewEventHandler(...)`
  - `handlers.NewOnvifHandler(...)`
  - `handlers.NewOnvifFrameHandler(...)`
  - `handlers.NewDeepstreamMinimalHandler(...)`
- Registers HTTP routes:
  - `/` -> `homeHandler.Handle`
  - `/<app-url-path>/snapshot/` -> `snapshotHandler.Handle`
  - `/<app-url-path>/event/` -> `eventHandler.Handle`
  - `/<app-url-path>/onvif/` -> `onvifHandler.Handle`
  - `/<app-url-path>/onvifframe/` -> `onvifFrameHandler.Handle`
  - `/<app-url-path>/deepstreamminimal/` -> `deepstreamMinimalHandler.Handle`
  - `/<app-url-path>/event/processing/` -> event processing endpoint
  - `/<app-url-path>/onvif/processing/` -> ONVIF processing endpoint
  - `/<app-url-path>/onvifframe/processing/` -> ONVIF frame processing endpoint
  - `/<app-url-path>/deepstreamminimal/processing/` -> DeepStream minimal processing endpoint

### src/domain/entities/appregistration.go
- Loads `config/register.graphql`.
- Populates placeholders using environment variables and default app constants.
- Returns the populated GraphQL registration payload.

### src/application/services/graphqlservice.go
- Wraps `GraphqlRepository`.
- Exposes methods for GraphQL operations:
  - `GetIdentityProviderRegisteredIssuers(...)`
  - `GetVmsIds(...)`
  - `GetSnapshot(...)`
  - `GetRestEventTopicEndpoint(...)`
  - `GetRestMetadataTopicEndpoint(...)`
  - `RegisterConnectivitySample(...)`

### src/infrastructure/repositories/graphqlrepository.go
- Sends raw GraphQL requests to the AI Bridge endpoint.
- Implements these operations:
  - `GetIdentityProviderRegisteredIssuers(...)`
  - `GetVmsIds(...)`
  - `GetSnapshot(...)`
  - `GetRestEventTopicEndpoint(...)`
  - `GetRestMetadataTopicEndpoint(...)`
  - `RegisterConnectivitySample(...)`
- Uses `sendRequest(...)` to POST GraphQL JSON and validate HTTP 200.

### GraphQL registration sequence
1. `main.go` calls `graphqlService.GetVmsIds(...)`.
2. `main.go` calls `appRegistration.GetPopulatedRegistrationFileContent()`.
3. `main.go` loops through each VMS ID and calls `graphqlService.RegisterConnectivitySample(...)`.
4. `graphqlService.RegisterConnectivitySample(...)` delegates to `graphqlRepository.RegisterConnectivitySample(...)`.
5. The repository sends the GraphQL mutation:
   - `mutation { register(input: { id: "<vmsId>" ... }) { id } }`

---

## 2. Request handling flow

### src/application/handlers/homehandler.go
- Serves the root `home.html` page for `"/"`.

### src/application/handlers/snapshothandler.go
Flow for `"/snapshot/"` requests:
1. Validates OAuth token using `tokenService.ExtractAndVerifyToken(...)`.
2. If the request is exactly `/<app-url-path>/snapshot/`, it renders the configuration page.
3. Otherwise, extracts `cameraID` and `streamID` from the URL.
4. Calls `graphqlService.GetSnapshot(...)` to fetch a base64 snapshot image.
5. Renders `templates/snapshot-camera-page.html` with the image.

### src/application/handlers/eventhandler.go
Flow for `"/event/"` requests:
- `Handle(...)`:
  1. Extracts `topicName`, `cameraID`, and `streamID` from the URL.
  2. Checks whether event sending is active via `TopicRestService.IsDataBeingSent(cameraID)`.
  3. Renders `templates/event-camera-page.html` with status data.
- `ProcessingHandle(...)`:
  1. Reads JSON body containing `cameraId` and `topicName`.
  2. If already sending, stops with `TopicRestService.StopSendingData(cameraID)`.
  3. Otherwise, reads a single analytic event file and calls `TopicRestService.SendDataAsync(...)`.
  4. Responds with JSON event status.

### src/application/handlers/onvifhandler.go
Flow for `"/onvif/"` requests:
- `Handle(...)`:
  1. Extracts `topicName`, `cameraID`, `streamID`.
  2. Checks sending state.
  3. Renders `templates/onvif-metadata-page.html`.
- `ProcessingHandle(...)`:
  1. Reads JSON body with `cameraId`, `streamId`, and `topicName`.
  2. Stops if already sending.
  3. Otherwise, loads XML metadata files and calls `TopicRestService.SendDataAsync(...)`.

### src/application/handlers/onvifframehandler.go
Flow for `"/onvifframe/"` requests:
- Same pattern as `OnvifHandler`, but uses `templates/onvif-frame-metadata-page.html` and `NewFileReader("onvif-frame", "xml")`.

### src/application/handlers/deepstreamminimalhandler.go
Flow for `"/deepstreamminimal/"` requests:
- Same pattern as `OnvifHandler`, but uses `templates/deepstreamminimal-metadata-page.html` and `NewFileReader("dsmin", "json")`.

---

## 3. Event and metadata submission flow

### src/application/services/topicrestservice.go
Core submission engine:
- `SendDataAsync(cameraID, streamID, topicName, topicFormat, fileFormat, files)`:
  1. Marks the `cameraID` active in a `sync.Map`.
  2. Launches a goroutine with a 1-second ticker.
  3. Resolves topic REST endpoint using `graphqlService`:
     - `GetRestEventTopicEndpoint(...)` for events.
     - `GetRestMetadataTopicEndpoint(...)` for metadata.
  4. Loops every second while the cameraID is active.
  5. Selects and processes the payload file:
     - events: `TreatEventFile(...)`
     - metadata: `TreatMetadataFile(...)`
  6. Sends the payload via `repositories.SendPostRequest(...)`.
- `StopSendingData(cameraID)` removes the camera from the running set.

### src/application/services/filereader.go
Payload loader and template injector:
- `ReadSingleFile()` loads `templates/<fileName>.<fileFormat>`.
- `ReadMultipleFiles()` loads all embedded template files matching `<fileName>\d+.<fileFormat>`.
- `TreatEventFile(content, cameraID)` replaces `{{ .CameraID }}`.
- `TreatMetadataFile(content, sourceStreamID)` substitutes:
  - `##CurrentTimestamp##`
  - `##CameraStreamId##`

### src/infrastructure/repositories/restrepository.go
- `SendPostRequest(url, body, contentType)` performs the actual HTTP POST.
- Requires HTTP 200 response to succeed.

---

## 4. Utility support in the flow

### src/application/services/querystringservice.go
- Extracts `cameraID`, `streamID`, and `topicName` from request URLs using regex.
- Used by all handlers for routing and parameter parsing.

### src/application/services/tokenservice.go
- Extracts `Authorization: Bearer ...` header when OAuth is enforced.
- Calls `graphqlRepository.GetIdentityProviderRegisteredIssuers(...)`.
- Verifies token claims with `entities.NewTokenValidator(...)`.

### src/domain/entities/tokenclaims.go
- Defines the expected client ID and registered issuers.
- Validates the token issuer matches a registered AI Bridge issuer.

### src/domain/entities/tokenvalidator.go
- Parses token claims without verification.
- Uses OIDC discovery from the token issuer URL.
- Verifies the JWT signature and issuer.

---

## 5. End-to-end summary

1. App starts in `main.go`.
2. The sample app registers itself with AI Bridge using `register.graphql`.
3. Incoming HTTP requests are routed to handlers.
4. Handlers parse the request path and optionally validate OAuth tokens.
5. Snapshot requests call GraphQL snapshot query.
6. Event and metadata actions launch asynchronous REST publishing.
7. `TopicRestService` resolves the topic REST endpoint and repeatedly posts templated payload data.
8. Payload templates are loaded from embedded `templates/` files and substituted at runtime.
