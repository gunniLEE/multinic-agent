# MultiNic Agent 배포 가이드

OpenStack VM 환경에서 네트워크 인터페이스를 자동으로 구성하는 Kubernetes DaemonSet 에이전트의 배포 가이드입니다.

## 배포 환경 선택

### 🏭 프로덕션 환경
- **용도**: 실제 운영 환경
- **데이터베이스**: 외부 MariaDB/MySQL 필요
- **배포 범위**: Agent DaemonSet + 기본 리소스만

### 🧪 테스트 환경
- **용도**: 개발, 테스트, 데모
- **데이터베이스**: 내장 MariaDB 자동 배포
- **배포 범위**: Agent + MariaDB + 테스트 데이터

## 빠른 배포

### 테스트 환경 (권장 시작점)

```bash
# 1단계: 테스트 환경 배포
./scripts/deploy.sh test

# 2단계: 배포 상태 확인
kubectl get pods -n multinic-system

# 3단계: 로그 확인
kubectl logs -f daemonset/multinic-agent -n multinic-system
```

### 프로덕션 환경

```bash
# 1단계: 외부 DB 설정 (사전 준비 필요)
# deployments/production/02-configmap.yaml 수정
# deployments/production/03-secret.yaml 수정

# 2단계: 프로덕션 배포
./scripts/deploy.sh production

# 3단계: 배포 상태 확인
kubectl get pods -n multinic-system
```

## 상세 배포 과정

### 🧪 테스트 환경 배포

#### 1. 사전 준비
```bash
# kubectl 연결 확인
kubectl cluster-info

# Docker 이미지 빌드 (선택사항)
./scripts/build-image.sh
```

#### 2. 배포 실행
```bash
# 방법 1: 통합 스크립트 사용
./scripts/deploy.sh test

# 방법 2: 직접 스크립트 실행
./scripts/deploy-test.sh
```

#### 3. 배포 확인
```bash
# 모든 리소스 확인
kubectl get all -n multinic-system

# Pod 상태 확인
kubectl get pods -n multinic-system
NAME                   READY   STATUS    RESTARTS   AGE
mariadb-0              1/1     Running   0          2m
multinic-agent-xxxxx   1/1     Running   0          1m

# 에이전트 로그 확인
kubectl logs -f daemonset/multinic-agent -n multinic-system

# MariaDB 로그 확인
kubectl logs -f statefulset/mariadb -n multinic-system
```

#### 4. 데이터베이스 확인
```bash
# MariaDB 접속
kubectl exec -it mariadb-0 -n multinic-system -- mysql -u root -pqudrjs1245!

# 테스트 데이터 확인
USE multinic;
SELECT n.attached_node_name, COUNT(mi.id) as interface_count
FROM node_table n
LEFT JOIN multi_interface mi ON n.attached_node_id = mi.attached_node_id
GROUP BY n.attached_node_name;
```

### 🏭 프로덕션 환경 배포

#### 1. 외부 데이터베이스 준비
외부 MariaDB/MySQL에 스키마를 설정하세요:
```bash
# 로컬에서 스키마 생성 (외부 DB에 적용)
mysql -h your-db-host -u your-username -p < scripts/create_test_db.sql
```

#### 2. 설정 파일 수정

**ConfigMap 설정** (`deployments/production/02-configmap.yaml`):
```yaml
data:
  DB_HOST: "your-mysql-host.example.com"
  DB_PORT: "3306"
  DB_NAME: "multinic"
  DB_USERNAME: "multinic_user"
  AGENT_CHECK_INTERVAL: "30s"
  LOG_LEVEL: "info"
```

**Secret 설정** (`deployments/production/03-secret.yaml`):
```bash
# 비밀번호 Base64 인코딩
echo -n "your-password" | base64

# Secret 파일 수정
data:
  DB_PASSWORD: "<base64-encoded-password>"
```

#### 3. 배포 실행
```bash
# 방법 1: 통합 스크립트 사용
./scripts/deploy.sh production

# 방법 2: 직접 스크립트 실행
./scripts/deploy-production.sh
```

#### 4. 배포 확인
```bash
# 에이전트 상태 확인
kubectl get daemonset -n multinic-system

# 로그에서 DB 연결 확인
kubectl logs -f daemonset/multinic-agent -n multinic-system
```

## 트러블슈팅

### 일반적인 문제들

#### 1. CrashLoopBackOff
```bash
# 원인: DB 연결 실패
kubectl logs multinic-agent-xxxxx -n multinic-system

# 해결방법:
# - ConfigMap의 DB_HOST 확인
# - Secret의 DB_PASSWORD 확인
# - 외부 DB 접근성 확인
```

#### 2. ImagePullBackOff
```bash
# 원인: Docker 이미지 없음
# 해결방법:
./scripts/build-image.sh
```

#### 3. 권한 오류
```bash
# 원인: RBAC 설정 문제
kubectl get serviceaccount,clusterrole,clusterrolebinding -n multinic-system

# 해결방법:
kubectl apply -f deployments/production/04-rbac.yaml
```

### 로그 분석

#### 정상 동작 로그
```json
{"level":"INFO","timestamp":"2025-06-03T16:11:30.348Z","caller":"agent/main.go:131","msg":"Found interfaces","node_name":"cluster2-control-plane","count":3}
```

#### 오류 로그 예시
```json
{"level":"ERROR","timestamp":"2025-06-03T16:11:30.348Z","caller":"database/database.go:45","msg":"Failed to connect to database","error":"dial tcp: lookup mysql on 127.0.0.11:53: no such host"}
```

## 배포 정리

### 테스트 환경 정리
```bash
# 방법 1: 통합 스크립트
./scripts/cleanup.sh test

# 방법 2: 직접 스크립트
./scripts/cleanup-test.sh
```

### 프로덕션 환경 정리
```bash
# 방법 1: 통합 스크립트
./scripts/cleanup.sh production

# 방법 2: 직접 스크립트
./scripts/cleanup-production.sh
```

## 배포 아키텍처

### 테스트 환경 구성도
```
┌─────────────────────────────────────┐
│ Kubernetes Cluster                  │
│ ┌─────────────────────────────────┐ │
│ │ multinic-system namespace       │ │
│ │                                 │ │
│ │ ┌─────────────┐ ┌─────────────┐ │ │
│ │ │ multinic-   │ │ mariadb-0   │ │ │
│ │ │ agent       │ │ (StatefulSet│ │ │
│ │ │ (DaemonSet) │ │ + PVC)      │ │ │
│ │ └─────────────┘ └─────────────┘ │ │
│ │                                 │ │
│ │ ┌─────────────────────────────┐ │ │
│ │ │ ConfigMaps & Secrets        │ │ │
│ │ │ - Agent Config              │ │ │
│ │ │ - DB Init Scripts           │ │ │
│ │ │ - DB Credentials            │ │ │
│ │ └─────────────────────────────┘ │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### 프로덕션 환경 구성도
```
┌─────────────────────────────────────┐ ┌─────────────────────────┐
│ Kubernetes Cluster                  │ │ External Database       │
│ ┌─────────────────────────────────┐ │ │ ┌─────────────────────┐ │
│ │ multinic-system namespace       │ │ │ │ MariaDB/MySQL       │ │
│ │                                 │ │ │ │                     │ │
│ │ ┌─────────────┐                 │ │ │ │ multinic database   │ │
│ │ │ multinic-   │                 │ │ │ │ - multi_subnet      │ │
│ │ │ agent       │─────────────────┼─┼─┤ │ - node_table        │ │
│ │ │ (DaemonSet) │                 │ │ │ │ - multi_interface   │ │
│ │ └─────────────┘                 │ │ │ │ - cr_state          │ │
│ │                                 │ │ │ └─────────────────────┘ │
│ │ ┌─────────────────────────────┐ │ │ └─────────────────────────┘
│ │ │ ConfigMaps & Secrets        │ │ │
│ │ │ - Agent Config              │ │ │
│ │ │ - DB Credentials            │ │ │
│ │ └─────────────────────────────┘ │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

## 모니터링 및 유지보수

### 헬스체크
```bash
# Pod 상태 주기적 확인
kubectl get pods -n multinic-system

# 메모리/CPU 사용량 확인
kubectl top pods -n multinic-system
```

### 로그 로테이션
```bash
# 최근 100줄만 확인
kubectl logs --tail=100 daemonset/multinic-agent -n multinic-system

# 특정 시간 이후 로그 확인
kubectl logs --since=1h daemonset/multinic-agent -n multinic-system
```

### 업데이트
```bash
# 이미지 업데이트 후 재배포
./scripts/build-image.sh
kubectl rollout restart daemonset/multinic-agent -n multinic-system
```
