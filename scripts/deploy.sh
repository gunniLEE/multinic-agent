#!/bin/bash

# Kubernetes 배포 스크립트

set -e

echo "🚀 Deploying MultiNic Agent to Kubernetes..."

# kubectl이 설치되어 있는지 확인
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Kubernetes 연결 확인
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Cannot connect to Kubernetes cluster"
    exit 1
fi

echo "✅ Kubernetes connection verified"

# 매니페스트 적용
echo "📁 Applying manifests..."

echo "  📍 Creating namespace..."
kubectl apply -f deployments/01-namespace.yaml

echo "  🗂️  Creating configmap..."
kubectl apply -f deployments/02-configmap.yaml

echo "  🔐 Creating secret..."
kubectl apply -f deployments/03-secret.yaml

echo "  👤 Creating RBAC..."
kubectl apply -f deployments/04-rbac.yaml

echo "  🔄 Creating DaemonSet..."
kubectl apply -f deployments/05-daemonset.yaml

echo ""
echo "✅ Deployment completed!"
echo ""
echo "📊 Checking deployment status:"
kubectl get all -n multinic-system

echo ""
echo "🔍 To view logs:"
echo "   kubectl logs -f daemonset/multinic-agent -n multinic-system"
echo ""
echo "🗑️  To cleanup:"
echo "   ./scripts/cleanup.sh" 