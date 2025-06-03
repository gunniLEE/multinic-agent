#!/bin/bash

# MultiNic Agent 프로덕션 배포 스크립트
# 외부 데이터베이스를 사용하는 프로덕션 환경용

set -e

echo "🚀 Deploying MultiNic Agent (Production) to Kubernetes..."

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

# 프로덕션 매니페스트 적용
echo "📁 Applying production manifests..."

echo "  📍 Creating namespace..."
kubectl apply -f deployments/production/01-namespace.yaml

echo "  🗂️  Creating configmap..."
kubectl apply -f deployments/production/02-configmap.yaml

echo "  🔐 Creating secret..."
kubectl apply -f deployments/production/03-secret.yaml

echo "  👤 Creating RBAC..."
kubectl apply -f deployments/production/04-rbac.yaml

echo "  🔄 Creating DaemonSet..."
kubectl apply -f deployments/production/05-daemonset.yaml

echo ""
echo "✅ Production deployment completed!"
echo ""
echo "📊 Checking deployment status:"
kubectl get all -n multinic-system

echo ""
echo "🔍 To view logs:"
echo "   kubectl logs -f daemonset/multinic-agent -n multinic-system"
echo ""
echo "⚠️  NOTE: This deployment expects an external database."
echo "   Make sure your database connection settings in configmap are correct."
echo ""
echo "🗑️  To cleanup:"
echo "   ./scripts/cleanup-production.sh" 