# Build stage
FROM golang:1.23-alpine AS builder

# 작업 디렉토리 설정
WORKDIR /app

# 의존성 복사 및 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 바이너리 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o multinic-agent cmd/agent/main.go

# Runtime stage - Ubuntu로 변경
FROM ubuntu:22.04

# 필요한 패키지 설치
RUN apt-get update && \
    apt-get install -y ca-certificates netplan.io && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# 작업 디렉토리 생성
WORKDIR /root/

# 바이너리 복사
COPY --from=builder /app/multinic-agent .

# netplan 디렉토리 생성
RUN mkdir -p /etc/netplan /var/backups/netplan

# 권한 설정
RUN chmod +x ./multinic-agent

# 실행
CMD ["./multinic-agent"] 