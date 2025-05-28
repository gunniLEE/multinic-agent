package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ibyeong-geon/multinic-agent/internal/config"
	"github.com/ibyeong-geon/multinic-agent/pkg/database"
	"github.com/ibyeong-geon/multinic-agent/pkg/logger"
)

func main() {
	// 커맨드라인 플래그 파싱
	var configPath string
	flag.StringVar(&configPath, "config", "", "configuration file path")
	flag.Parse()

	// 설정 로드
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 로거 초기화
	zapLogger, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// 설정 로그 출력
	zapLogger.Info("Configuration loaded",
		zap.String("db_host", cfg.Database.Host),
		zap.Int("db_port", cfg.Database.Port),
		zap.String("node_name", cfg.Agent.NodeName),
		zap.Int("check_interval", cfg.Agent.CheckInterval),
		zap.String("log_level", cfg.Logging.Level),
	)

	zapLogger.Info("Starting agent...")

	// DB 연결
	dbClient, err := database.NewClient(&cfg.Database, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbClient.Close()

	// TODO: Kubernetes 클라이언트 초기화

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 시그널 핸들링 (graceful shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 메인 루프를 고루틴으로 시작
	go runMainLoop(ctx, cfg, dbClient, zapLogger)

	// 시그널 대기
	sig := <-sigChan
	zapLogger.Info("Received signal", zap.String("signal", sig.String()))
	zapLogger.Info("Shutting down agent...")

	// Context 취소로 모든 고루틴 종료
	cancel()

	// 정리 작업을 위한 약간의 대기 시간
	time.Sleep(2 * time.Second)

	zapLogger.Info("Agent shutdown complete")
}

// runMainLoop는 주기적으로 DB를 체크하고 필요한 작업을 수행합니다
func runMainLoop(ctx context.Context, cfg *config.Config, dbClient *database.Client, logger *zap.Logger) {
	ticker := time.NewTicker(time.Duration(cfg.Agent.CheckInterval) * time.Second)
	defer ticker.Stop()

	// 시작하자마자 한 번 실행
	if err := processNetworkInterfaces(cfg, dbClient, logger); err != nil {
		logger.Error("Failed to process network interfaces", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Main loop stopped")
			return
		case <-ticker.C:
			if err := processNetworkInterfaces(cfg, dbClient, logger); err != nil {
				logger.Error("Failed to process network interfaces", zap.Error(err))
			}
		}
	}
}

// processNetworkInterfaces는 네트워크 인터페이스를 처리합니다
func processNetworkInterfaces(cfg *config.Config, dbClient *database.Client, logger *zap.Logger) error {
	nodeName := cfg.Agent.NodeName
	if nodeName == "" {
		// 노드 이름이 없으면 호스트명 사용
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		nodeName = hostname
	}

	logger.Info("Processing network interfaces", zap.String("node_name", nodeName))

	// DB에서 네트워크 인터페이스 정보 조회
	interfaces, err := dbClient.GetNodeInterfaces(nodeName)
	if err != nil {
		return err
	}

	if len(interfaces) == 0 {
		logger.Warn("No interfaces found for node", zap.String("node_name", nodeName))
		return nil
	}

	logger.Info("Found interfaces",
		zap.String("node_name", nodeName),
		zap.Int("count", len(interfaces)),
	)

	// TODO: Netplan 파일 생성
	// TODO: Netplan 적용
	// TODO: K8s 노드 레이블/어노테이션 업데이트

	return nil
}
