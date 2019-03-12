#!/usr/bin/env bash

set -e

if [[ -z "$1" ]]; then
  echo "RUNNER_URL is mandatory"
  exit 1
fi

RUNNER_URL=$1
IMAGE_TAG="${2:-latest}"

helm tiller run \
    helm upgrade --install \
    survey-launcher \
    k8s/helm \
    --set surveyRunnerUrl=https://${RUNNER_URL} \
    --set image.tag=${IMAGE_TAG}
