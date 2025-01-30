#!/bin/bash

PROJECT_ID="single-patrol-449300-j9"
REGISTRY="gcr.io"
IMAGE_NAME="video_intelligence"
TAG="latest"
IMAGE="${REGISTRY}/${PROJECT_ID}/${IMAGE_NAME}:${TAG}"
SERVICE_NAME="video-intelligence"
REGION="us-central1"

if ! docker info > /dev/null 2>&1; then
  echo "Docker is not running! Please start Docker."
  exit 1
fi

echo "Logging into Google Cloud..."
gcloud auth login

echo "Logging into Google Cloud Docker Registry..."
gcloud auth configure-docker

echo "Building Docker image for amd64 platform..."
docker build --platform linux/amd64 -t ${IMAGE} .

if [ $? -ne 0 ]; then
  echo "Error building Docker image!"
  exit 1
fi

echo "Pushing Docker image to GCR..."
docker push ${IMAGE}

if [ $? -ne 0 ]; then
  echo "Error pushing Docker image to GCR!"
  exit 1
fi

echo "Docker image successfully pushed to GCR: ${IMAGE}"

echo "Deploying container to Cloud Run..."
gcloud run deploy ${SERVICE_NAME} \
  --image ${IMAGE} \
  --platform managed \
  --region ${REGION} \
  --allow-unauthenticated

if [ $? -ne 0 ]; then
  echo "Error deploying container to Cloud Run!"
  exit 1
fi

echo "Container successfully deployed to Google Cloud Run!"
