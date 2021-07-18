#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

docker build -t "$IMAGE_TAG:latest" .
docker tag "$IMAGE_TAG:latest" "$IMAGE_TAG:$(git rev-parse HEAD)"

docker push "$IMAGE_TAG:latest"
docker push "$IMAGE_TAG:$(git rev-parse HEAD)"
