# MultiNic Agent

OpenStack VM 환경에서 네트워크 인터페이스를 자동으로 구성하는 Kubernetes DaemonSet 에이전트입니다.

## 개요

OpenStack에서 VM에 추가된 네트워크 인터페이스가 자동으로 VM 내부에 반영되지 않는 문제를 해결합니다. 이 에이전트는 관리 클러스터의 데이터베이스에서 네트워크 인터페이스 정보를 읽어와 netplan 파일을 자동으로 생성하고 적용합니다.

## 프로젝트 구조

```
multinic-agent/
├── cmd/
│   └── agent/
│       └── main.go                 # 에이전트 메인 애플리케이션
├── pkg/
│   ├── config/
│   │   └── config.go              # 구성 관리
│   ├── database/
│   │   └── database.go            # 데이터베이스 연결 및 쿼리
│   └── logger/
│       └── logger.go              # 로깅 설정
├── config/
│   ├── config.yaml               # 로컬 개발용 설정
│   └── config.example.yaml       # 설정 템플릿
├── deployments/
│   ├── production/               # 프로덕션 환경용 매니페스트
│   │   ├── 01-namespace.yaml
│   │   ├── 02-configmap.yaml
│   │   ├── 03-secret.yaml
│   │   ├── 04-rbac.yaml
│   │   └── 05-daemonset.yaml
│   └── test-db/                  # 테스트 환경용 DB 매니페스트
│       ├── 06-mariadb-configmap.yaml
│       ├── 07-mariadb-secret.yaml
│       ├── 08-mariadb-service.yaml
│       └── 09-mariadb-statefulset.yaml
├── scripts/
│   ├── deploy.sh                 # 통합 배포 스크립트
│   ├── cleanup.sh               # 통합 정리 스크립트
│   ├── deploy-production.sh     # 프로덕션 배포
│   ├── deploy-test.sh          # 테스트 환경 배포
│   ├── cleanup-production.sh   # 프로덕션 정리
│   ├── cleanup-test.sh         # 테스트 환경 정리
│   ├── build-image.sh          # Docker 이미지 빌드
│   └── create_test_db.sql      # 로컬 테스트 DB 설정
├── Dockerfile
├── go.mod
└── README.md
```

## 배포 환경

### 프로덕션 환경
- **용도**: 실제 운영 환경
- **데이터베이스**: 외부 MariaDB/MySQL 사용
- **포함 리소스**: Agent DaemonSet, ConfigMap, Secret, RBAC

### 테스트 환경
- **용도**: 개발 및 테스트
- **데이터베이스**: 내장 MariaDB 사용 (테스트 데이터 포함)
- **포함 리소스**: Agent + MariaDB StatefulSet + 모든 의존성

## 빠른 시작

### 1. 테스트 환경 배포

```bash
# 테스트 환경 배포 (내장 MariaDB 포함)
./scripts/deploy.sh test

# 또는 직접 실행
./scripts/deploy-test.sh
```

### 2. 프로덕션 환경 배포

```bash
# 프로덕션 환경 배포 (외부 DB 필요)
./scripts/deploy.sh production

# 또는 직접 실행
./scripts/deploy-production.sh
```

### 3. 배포 확인

```bash
# Pod 상태 확인
kubectl get pods -n multinic-system

# 에이전트 로그 확인
kubectl logs -f daemonset/multinic-agent -n multinic-system

# 테스트 환경의 경우 MariaDB 로그도 확인 가능
kubectl logs -f statefulset/mariadb -n multinic-system
```

### 4. 정리

```bash
# 테스트 환경 정리
./scripts/cleanup.sh test

# 프로덕션 환경 정리
./scripts/cleanup.sh production
```

## 설정

### 프로덕션 환경 설정

프로덕션 배포 전에 `deployments/production/02-configmap.yaml`과 `deployments/production/03-secret.yaml`을 수정하여 외부 데이터베이스 연결 정보를 설정하세요.

#### ConfigMap 설정
```yaml
# deployments/production/02-configmap.yaml
DB_HOST: "your-mysql-host"
DB_PORT: "3306"
DB_NAME: "multinic"
DB_USERNAME: "your-username"
```

#### Secret 설정
```yaml
# deployments/production/03-secret.yaml
data:
  DB_PASSWORD: "<base64-encoded-password>"
```

### 로컬 개발 환경

```bash
# 로컬 MariaDB 설정 (로컬 개발용)
mysql -u root -p < scripts/create_test_db.sql

# 로컬 실행
go run cmd/agent/main.go
```

## 데이터베이스 스키마

### 테이블 구조

1. **multi_subnet**: 서브넷 정보 (CIDR 포함)
2. **node_table**: 노드 정보
3. **multi_interface**: 인터페이스 정보 (MAC, 포트 ID 등)
4. **cr_state**: CR 변경 추적

### 샘플 데이터

테스트 환경에는 다음 노드들의 샘플 데이터가 포함됩니다:
- `cluster2-control-plane` (실제 클러스터 노드)
- `worker-node-1`, `worker-node-2`, `worker-node-3` (샘플 노드)

## 모니터링

### 로그 확인
```bash
# 에이전트 로그 (실시간)
kubectl logs -f daemonset/multinic-agent -n multinic-system

# MariaDB 로그 (테스트 환경)
kubectl logs -f statefulset/mariadb -n multinic-system
```

### 데이터베이스 접속 (테스트 환경)
```bash
# MariaDB 접속
kubectl exec -it mariadb-0 -n multinic-system -- mysql -u root -p

# 인터페이스 데이터 확인
USE multinic;
SELECT n.attached_node_name, mi.port_id, ms.subnet_name, ms.cidr 
FROM multi_interface mi
JOIN node_table n ON mi.attached_node_id = n.attached_node_id
JOIN multi_subnet ms ON mi.subnet_id = ms.subnet_id
WHERE mi.status = 'active';
```

## 개발

### Docker 이미지 빌드
```bash
./scripts/build-image.sh
```

### 요구사항
- Go 1.23+
- Docker
- Kubernetes 클러스터
- kubectl

## 라이선스

MIT License 