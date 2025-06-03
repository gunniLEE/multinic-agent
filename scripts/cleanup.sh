#!/bin/bash

# MultiNic Agent 통합 정리 스크립트
# 프로덕션 또는 테스트 환경 선택 가능

set -e

# 사용법 출력
usage() {
    echo "Usage: $0 [production|test]"
    echo ""
    echo "Environments:"
    echo "  production  - Cleanup agent only"
    echo "  test        - Cleanup agent and MariaDB"
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
        echo "🏭 Starting production cleanup..."
        exec ./scripts/cleanup-production.sh
        ;;
    test)
        echo "🧪 Starting test environment cleanup..."
        exec ./scripts/cleanup-test.sh
        ;;
    *)
        echo "❌ Invalid environment: $ENVIRONMENT"
        usage
        ;;
esac 