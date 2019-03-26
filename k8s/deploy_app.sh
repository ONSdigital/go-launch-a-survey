#!/usr/bin/env bash

set -e

if [[ -z "$1" ]]; then
  echo "RUNNER_URL is mandatory"
  exit 1
fi

RUNNER_URL=$1
DOCKER_REGISTRY="${2:-eu.gcr.io/census-eq-ci}"
IMAGE_TAG="${3:-latest}"

helm tiller run \
    helm upgrade --install \
    survey-launcher \
    k8s/helm \
    --set surveyRunnerUrl=https://${RUNNER_URL} \
    --set image.repository=${DOCKER_REGISTRY}/go-launch-a-survey \
    --set image.tag=${IMAGE_TAG}
