#!/bin/bash

# MultiNic Agent 프로덕션 환경 정리 스크립트

set -e

echo "🗑️  Cleaning up MultiNic Agent (Production) from Kubernetes..."

# kubectl이 설치되어 있는지 확인
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# 리소스 삭제 (역순으로)
echo "🔄 Deleting DaemonSet..."
kubectl delete -f deployments/production/05-daemonset.yaml --ignore-not-found=true

echo "👤 Deleting RBAC..."
kubectl delete -f deployments/production/04-rbac.yaml --ignore-not-found=true

echo "🔐 Deleting secret..."
kubectl delete -f deployments/production/03-secret.yaml --ignore-not-found=true

echo "🗂️  Deleting configmap..."
kubectl delete -f deployments/production/02-configmap.yaml --ignore-not-found=true

echo "📍 Deleting namespace..."
kubectl delete -f deployments/production/01-namespace.yaml --ignore-not-found=true

echo ""
echo "✅ Production cleanup completed!"

# 확인
echo "📊 Verifying cleanup:"
if kubectl get namespace multinic-system &> /dev/null; then
    echo "⚠️  Namespace still exists (may take a moment to fully delete)"
    echo "   Remaining resources:"
    kubectl get all -n multinic-system 2>/dev/null || echo "   No resources found"
else
    echo "✅ All production resources deleted successfully"
fi 