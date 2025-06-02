#!/bin/bash

# Kubernetes 리소스 정리 스크립트

set -e

echo "🗑️  Cleaning up MultiNic Agent from Kubernetes..."

# kubectl이 설치되어 있는지 확인
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# 리소스 삭제 (역순으로)
echo "🔄 Deleting DaemonSet..."
kubectl delete -f deployments/05-daemonset.yaml --ignore-not-found=true

echo "👤 Deleting RBAC..."
kubectl delete -f deployments/04-rbac.yaml --ignore-not-found=true

echo "🔐 Deleting secret..."
kubectl delete -f deployments/03-secret.yaml --ignore-not-found=true

echo "🗂️  Deleting configmap..."
kubectl delete -f deployments/02-configmap.yaml --ignore-not-found=true

echo "📍 Deleting namespace..."
kubectl delete -f deployments/01-namespace.yaml --ignore-not-found=true

echo ""
echo "✅ Cleanup completed!"

# 확인
echo "📊 Verifying cleanup:"
if kubectl get namespace multinic-system &> /dev/null; then
    echo "⚠️  Namespace still exists (may take a moment to fully delete)"
else
    echo "✅ All resources deleted successfully"
fi 