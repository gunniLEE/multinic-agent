# MultiNic Agent DaemonSet 배포 가이드

## 전체 배포 흐름

```
🔧 준비 → 🔨 빌드 → 🚀 배포 → 🔍 확인 → 🗑️ 정리
```

## 1. 사전 준비사항

### 필수 요구사항
- **Docker**: 이미지 빌드용
- **Kubernetes 클러스터**: 배포 대상
- **kubectl**: 클러스터 관리용
- **MySQL 데이터베이스**: 클러스터 내부 또는 외부

## 2. 빌드 및 배포

### 단계 1: Docker 이미지 빌드
```bash
./scripts/build-image.sh
```

### 단계 2: Kubernetes 배포
```bash
./scripts/deploy.sh
```

### 단계 3: 배포 상태 확인
```bash
kubectl get all -n multinic-system
kubectl logs -f daemonset/multinic-agent -n multinic-system
```

## 3. 정리
```bash
./scripts/cleanup.sh
```

## 4. 주요 설정
- `DB_HOST`: 데이터베이스 호스트 
- `AGENT_CHECK_INTERVAL`: 체크 주기 (기본: 30초)
- `LOG_LEVEL`: 로그 레벨
- `NETPLAN_DRY_RUN`: 테스트 모드 