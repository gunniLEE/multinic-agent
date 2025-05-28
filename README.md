# MultiNic Agent

OpenStack 환경에서 VM의 네트워크 인터페이스를 자동으로 구성하는 에이전트

## 프로젝트 개요
- OpenStack에서 VM에 attach한 인터페이스를 자동으로 감지하고 설정
- Netplan 파일을 자동으로 생성/적용
- Kubernetes 노드의 label/annotation에 인터페이스 정보 자동 업데이트
- DaemonSet으로 배포 예정

## 현재 진행 상황 (2025-05-29)
- ✅ 프로젝트 구조 설정
- ✅ 설정 관리 모듈 (YAML/환경변수 지원)
- ✅ 로거 구현 (zap 사용, JSON/Text 포맷)
- ✅ MySQL DB 연결 모듈
- ✅ 메인 루프 구조 (30초마다 DB 체크)
- 🔲 Netplan 파일 생성/적용 모듈
- 🔲 Kubernetes 클라이언트 (노드 label/annotation 업데이트)
- 🔲 DaemonSet 배포 매니페스트

## 테스트 환경
- MySQL DB: `multinic_db` (localhost:3306)
- 테스트 데이터: worker-node-1, worker-node-2의 네트워크 인터페이스 정보

## 실행 방법
```bash
# 설정 파일과 함께 실행
./multinic-agent --config config/config.yaml

# DB 테스트
go run cmd/test-db/main.go
```

## 다음 작업
1. Netplan 모듈 개발
2. K8s 클라이언트 통합
3. DaemonSet 배포 준비 