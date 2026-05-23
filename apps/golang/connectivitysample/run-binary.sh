#!/bin/bash
set -e -o pipefail

# Locate files in the system
APP_ROOT_DIR=$(realpath `dirname "$0"`)
EXECUTABLE_DIR="${APP_ROOT_DIR}/bin"
EXECUTABLE_FILE="${EXECUTABLE_DIR}/connectivitysample"

# Check the binary was built
if [ ! -f "${EXECUTABLE_FILE}" ]; then
	echo 'Make sure to build the binary first.'
	echo 'Navigate to the build folder and run the `build-binary.sh` script.'
	exit 1
fi

# Read the .env file to obtain the App configuration
source "${APP_ROOT_DIR}/.env"

# Declare needed environment variables
export TAG="1.1.0"
export EXTERNAL_HOSTNAME=${EXTERNAL_HOSTNAME}
export APP_WEBSERVER_PORT=${APP_WEBSERVER_PORT}
export APP_URL_PATH=${APP_URL_PATH}
export APP_ID=${APP_ID}
export APP_NAME=${APP_NAME}
export APP_DESCRIPTION=${APP_DESCRIPTION}
export MANUFACTURER_NAME=${MANUFACTURER_NAME}
export SNAPSHOT_TOPIC_NAME=${SNAPSHOT_TOPIC_NAME}
export SNAPSHOT_TOPIC_DESCRIPTION=${SNAPSHOT_TOPIC_DESCRIPTION}
export ANALYTIC_EVENT_TOPIC_NAME=${ANALYTIC_EVENT_TOPIC_NAME}
export ANALYTIC_EVENT_TOPIC_PATH=${ANALYTIC_EVENT_TOPIC_PATH}
export ANALYTIC_EVENT_TOPIC_DESCRIPTION=${ANALYTIC_EVENT_TOPIC_DESCRIPTION}
export ONVIF_TOPIC_NAME=${ONVIF_TOPIC_NAME}
export ONVIF_TOPIC_PATH=${ONVIF_TOPIC_PATH}
export ONVIF_TOPIC_DESCRIPTION=${ONVIF_TOPIC_DESCRIPTION}
export ONVIF_FRAME_TOPIC_NAME=${ONVIF_FRAME_TOPIC_NAME}
export ONVIF_FRAME_TOPIC_PATH=${ONVIF_FRAME_TOPIC_PATH}
export ONVIF_FRAME_TOPIC_DESCRIPTION=${ONVIF_FRAME_TOPIC_DESCRIPTION}
export DEEPSTREAM_MINIMAL_TOPIC_NAME=${DEEPSTREAM_MINIMAL_TOPIC_NAME}
export DEEPSTREAM_MINIMAL_TOPIC_PATH=${DEEPSTREAM_MINIMAL_TOPIC_PATH}
export DEEPSTREAM_MINIMAL_TOPIC_DESCRIPTION=${DEEPSTREAM_MINIMAL_TOPIC_DESCRIPTION}
export TLS_SCHEME=${TLS_SCHEME:-http}

# Run binary providing relevant parameters
${EXECUTABLE_FILE} \
	-aib-webservice-location localhost:4000 \
	-app-registration-file-path "${APP_ROOT_DIR}/config/register.graphql" \
	-app-url-path ${APP_URL_PATH} \
	-app-webserver-port ${APP_WEBSERVER_PORT} \
	-enforce-oauth=true \
	-snapshot-max-height 600 \
	-snapshot-max-width 600 \
	-tls-certificate-file "${APP_ROOT_DIR}/certs/tls-server/server.crt" \
	-tls-key-file "${APP_ROOT_DIR}/certs/tls-server/server.key" \
	-tls-enabled=${TLS_ENABLED:-false}
