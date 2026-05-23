package common

// XProtect Management Client ClientID
const ManagementClient_Oauth_ClientID string = "VmsAdminClient"

// App constants
const AppID = "c24cf88c-97b4-4917-9fdb-4279396b0dee"
const AppName = "Bridge Connectivity Services"
const AppDescription = "Custom IVA app registered as a separate AI Bridge integration."
const AppDefaultManufacturerName = "Bridge Services"
const DefaultAppURLPath = "bridge-connectivity-services"

// Registered topic defaults
const SnapshotTopicName = "bridge-snapshot-service"
const SnapshotTopicPath = "bridge-snapshot-service"
const SnapshotTopicDescription = "Open the snapshot service for a selected camera."

const AnalyticEventTopicName = "bridge-analytic-events-service"
const AnalyticEventTopicPath = "bridge-analytic-events-service"
const AnalyticEventTopicDescription = "Start or stop analytic event delivery."

const OnvifTopicName = "bridge-onvif-metadata-service"
const OnvifTopicPath = "bridge-onvif-metadata-service"
const OnvifTopicDescription = "Start or stop ONVIF metadata delivery."

const OnvifFrameTopicName = "bridge-onvif-frame-service"
const OnvifFrameTopicPath = "bridge-onvif-frame-service"
const OnvifFrameTopicDescription = "Start or stop ONVIF frame metadata delivery."

const DeepstreamMinimalTopicName = "bridge-deepstream-minimal-service"
const DeepstreamMinimalTopicPath = "bridge-deepstream-minimal-service"
const DeepstreamMinimalTopicDescription = "Start or stop DeepStream minimal metadata delivery."
