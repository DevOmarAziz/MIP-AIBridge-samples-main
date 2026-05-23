package entities

import (
	"errors"
	"os"

	"connectivitysample/src/common"
	"connectivitysample/src/shared/utils"
)

const manufacturerNameEnvVar = "MANUFACTURER_NAME" //If set, it will be used as the manufacturer name when registering the app
const appIDEnvVar = "APP_ID"
const appNameEnvVar = "APP_NAME"
const appDescriptionEnvVar = "APP_DESCRIPTION"
const snapshotTopicNameEnvVar = "SNAPSHOT_TOPIC_NAME"
const snapshotTopicPathEnvVar = "SNAPSHOT_TOPIC_PATH"
const snapshotTopicDescriptionEnvVar = "SNAPSHOT_TOPIC_DESCRIPTION"
const analyticEventTopicNameEnvVar = "ANALYTIC_EVENT_TOPIC_NAME"
const analyticEventTopicPathEnvVar = "ANALYTIC_EVENT_TOPIC_PATH"
const analyticEventTopicDescriptionEnvVar = "ANALYTIC_EVENT_TOPIC_DESCRIPTION"
const onvifTopicNameEnvVar = "ONVIF_TOPIC_NAME"
const onvifTopicPathEnvVar = "ONVIF_TOPIC_PATH"
const onvifTopicDescriptionEnvVar = "ONVIF_TOPIC_DESCRIPTION"
const onvifFrameTopicNameEnvVar = "ONVIF_FRAME_TOPIC_NAME"
const onvifFrameTopicPathEnvVar = "ONVIF_FRAME_TOPIC_PATH"
const onvifFrameTopicDescriptionEnvVar = "ONVIF_FRAME_TOPIC_DESCRIPTION"
const deepstreamMinimalTopicNameEnvVar = "DEEPSTREAM_MINIMAL_TOPIC_NAME"
const deepstreamMinimalTopicPathEnvVar = "DEEPSTREAM_MINIMAL_TOPIC_PATH"
const deepstreamMinimalTopicDescriptionEnvVar = "DEEPSTREAM_MINIMAL_TOPIC_DESCRIPTION"

type AppRegistration struct {
	registrationFile string // Path to the app registration file (register.graphql)
}

func NewAppRegistration(registrationFile string) (*AppRegistration, error) {
	if registrationFile == "" {
		return nil, errors.New("parameter 'registrationFile' can't be empty")
	}

	return &AppRegistration{
		registrationFile: registrationFile,
	}, nil
}

// Get the content of 'register.graphql' populated with the environment variables (and custom mappings)
func (ar *AppRegistration) GetPopulatedRegistrationFileContent() (string, error) {

	input, err := os.ReadFile(ar.registrationFile)
	if err != nil {
		return "", err
	}

	populatedRegistrationFileContent := utils.ExpandEnvWithDefault(string(input), appRegistrationMaping)

	return populatedRegistrationFileContent, nil
}

// Callback when applying the environment variables to the 'register.graphql' file
func appRegistrationMaping(key string) string {
	switch key {
	case manufacturerNameEnvVar:
		return getManufacturerName()
	case appIDEnvVar:
		return utils.GetEnv(appIDEnvVar, common.AppID)
	case appNameEnvVar:
		return utils.GetEnv(appNameEnvVar, common.AppName)
	case appDescriptionEnvVar:
		return utils.GetEnv(appDescriptionEnvVar, common.AppDescription)
	case snapshotTopicNameEnvVar:
		return utils.GetEnv(snapshotTopicNameEnvVar, common.SnapshotTopicName)
	case snapshotTopicPathEnvVar:
		return utils.GetEnv(snapshotTopicPathEnvVar, common.SnapshotTopicPath)
	case snapshotTopicDescriptionEnvVar:
		return utils.GetEnv(snapshotTopicDescriptionEnvVar, common.SnapshotTopicDescription)
	case analyticEventTopicNameEnvVar:
		return utils.GetEnv(analyticEventTopicNameEnvVar, common.AnalyticEventTopicName)
	case analyticEventTopicPathEnvVar:
		return utils.GetEnv(analyticEventTopicPathEnvVar, common.AnalyticEventTopicPath)
	case analyticEventTopicDescriptionEnvVar:
		return utils.GetEnv(analyticEventTopicDescriptionEnvVar, common.AnalyticEventTopicDescription)
	case onvifTopicNameEnvVar:
		return utils.GetEnv(onvifTopicNameEnvVar, common.OnvifTopicName)
	case onvifTopicPathEnvVar:
		return utils.GetEnv(onvifTopicPathEnvVar, common.OnvifTopicPath)
	case onvifTopicDescriptionEnvVar:
		return utils.GetEnv(onvifTopicDescriptionEnvVar, common.OnvifTopicDescription)
	case onvifFrameTopicNameEnvVar:
		return utils.GetEnv(onvifFrameTopicNameEnvVar, common.OnvifFrameTopicName)
	case onvifFrameTopicPathEnvVar:
		return utils.GetEnv(onvifFrameTopicPathEnvVar, common.OnvifFrameTopicPath)
	case onvifFrameTopicDescriptionEnvVar:
		return utils.GetEnv(onvifFrameTopicDescriptionEnvVar, common.OnvifFrameTopicDescription)
	case deepstreamMinimalTopicNameEnvVar:
		return utils.GetEnv(deepstreamMinimalTopicNameEnvVar, common.DeepstreamMinimalTopicName)
	case deepstreamMinimalTopicPathEnvVar:
		return utils.GetEnv(deepstreamMinimalTopicPathEnvVar, common.DeepstreamMinimalTopicPath)
	case deepstreamMinimalTopicDescriptionEnvVar:
		return utils.GetEnv(deepstreamMinimalTopicDescriptionEnvVar, common.DeepstreamMinimalTopicDescription)
	default:
		return os.Getenv(key)
	}
}

// The app's manufacturer (it's an OPTIONAL object when registering an app)
// Partners/Developers if you want to register your integration mind adjusting the following values:
// You can register your manufacturer name by sending an email to partner@milestone.dk
func getManufacturerName() string {
	manufacturerName := utils.GetEnv(manufacturerNameEnvVar, common.AppDefaultManufacturerName)

	// Even if the environment variable is set explicitly to 'empty' we will fallback to its default value.
	// AI Bridge will not accept an empty manufacturer name.
	if manufacturerName == "" {
		manufacturerName = common.AppDefaultManufacturerName
	}
	return manufacturerName
}
