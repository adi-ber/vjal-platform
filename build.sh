set -euo pipefail

APP_NAME=vjal-app
IMAGE_TAG=adi-ber/vjal-platform:latest
BUILD_DIR=build
ZIP_NAME=vjal-platform-offline.zip

# Clean previous
rm -rf "\${BUILD_DIR}"
mkdir -p "\${BUILD_DIR}/offline" "\${BUILD_DIR}/bin"

# Build Docker image
docker build -t "\${IMAGE_TAG}" .

# Create a container and copy out the compressed binary
CID=$(docker create "\${IMAGE_TAG}")
docker cp "\${CID}:/\${APP_NAME}" "\${BUILD_DIR}/bin/\${APP_NAME}"
docker rm "\${CID}"

# Copy runtime assets into the offline folder
cp config.json license.json "\${BUILD_DIR}/offline/"
mkdir -p "\${BUILD_DIR}/offline/forms" "\${BUILD_DIR}/offline/docs"
cp -r forms docs "\${BUILD_DIR}/offline/"

# Move binary into offline folder
cp "\${BUILD_DIR}/bin/\${APP_NAME}" "\${BUILD_DIR}/offline/"

# Zip it up
(cd "\${BUILD_DIR}/offline" && zip -r "../\${ZIP_NAME}" .)

echo "Offline package is ready at \${BUILD_DIR}/\${ZIP_NAME}"
