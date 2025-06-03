#!/bin/bash

# MultiNic Agent 통합 배포 스크립트
# 프로덕션 또는 테스트 환경 선택 가능

set -e

# 사용법 출력
usage() {
    echo "Usage: $0 [production|test]"
    echo ""
    echo "Environments:"
    echo "  production  - Deploy agent only (requires external database)"
    echo "  test        - Deploy agent with MariaDB (for testing)"
    echo ""
    echo "Examples:"
    echo "  $0 production"
    echo "  $0 test"
    exit 1
}

# 인수 확인
if [ $# -ne 1 ]; then
    usage
fi

ENVIRONMENT=$1

case $ENVIRONMENT in
    production)
        echo "🏭 Starting production deployment..."
        exec ./scripts/deploy-production.sh
        ;;
    test)
        echo "🧪 Starting test environment deployment..."
        exec ./scripts/deploy-test.sh
        ;;
    *)
        echo "❌ Invalid environment: $ENVIRONMENT"
        usage
        ;;
esac 