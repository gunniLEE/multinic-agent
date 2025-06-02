#!/bin/bash

# Docker 이미지 빌드 스크립트

set -e

IMAGE_NAME="multinic-agent"
IMAGE_TAG="latest"

echo "🔨 Building Docker image: ${IMAGE_NAME}:${IMAGE_TAG}"

# Docker 이미지 빌드
docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .

echo "✅ Docker image built successfully!"

# 이미지 확인
echo "📋 Image details:"
docker images | grep ${IMAGE_NAME}

echo ""
echo "🚀 To deploy to Kubernetes:"
echo "   ./scripts/deploy.sh" 