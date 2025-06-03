#!/bin/bash

# MultiNic Agent 테스트 배포 스크립트
# 내장 MariaDB를 포함하는 테스트 환경용

set -e

echo "🚀 Deploying MultiNic Agent (Test Environment) with MariaDB to Kubernetes..."

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
echo "📁 Applying test environment manifests..."

echo "  📍 Creating namespace..."
kubectl apply -f deployments/production/01-namespace.yaml

echo "  🗂️  Creating configmap..."
kubectl apply -f deployments/production/02-configmap.yaml

echo "  🔐 Creating secret..."
kubectl apply -f deployments/production/03-secret.yaml

echo "  👤 Creating RBAC..."
kubectl apply -f deployments/production/04-rbac.yaml

echo "  🗃️  Creating MariaDB ConfigMap..."
kubectl apply -f deployments/test-db/06-mariadb-configmap.yaml

echo "  🔐 Creating MariaDB Secret..."
kubectl apply -f deployments/test-db/07-mariadb-secret.yaml

echo "  📊 Creating MariaDB Service..."
kubectl apply -f deployments/test-db/08-mariadb-service.yaml

echo "  💾 Creating MariaDB StatefulSet..."
kubectl apply -f deployments/test-db/09-mariadb-statefulset.yaml

echo "  ⏳ Waiting for MariaDB to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mariadb -n multinic-system --timeout=300s

echo "  🔄 Creating DaemonSet..."
kubectl apply -f deployments/production/05-daemonset.yaml

echo ""
echo "✅ Test environment deployment completed!"
echo ""
echo "📊 Checking deployment status:"
kubectl get all -n multinic-system

echo ""
echo "🔍 To view logs:"
echo "   # Agent logs"
echo "   kubectl logs -f daemonset/multinic-agent -n multinic-system"
echo "   # MariaDB logs"
echo "   kubectl logs -f statefulset/mariadb -n multinic-system"
echo ""
echo "🔗 To connect to MariaDB:"
echo "   kubectl exec -it mariadb-0 -n multinic-system -- mysql -u root -p"
echo ""
echo "📄 Database includes test data for nodes:"
echo "   - cluster2-control-plane (actual cluster node)"
echo "   - worker-node-1, worker-node-2, worker-node-3 (sample nodes)"
echo ""
echo "🗑️  To cleanup:"
echo "   ./scripts/cleanup-test.sh" 