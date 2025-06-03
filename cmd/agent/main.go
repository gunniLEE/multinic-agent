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
	"github.com/ibyeong-geon/multinic-agent/pkg/netplan"
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

	// 인터페이스 정보 로그 출력
	for _, iface := range interfaces {
		logger.Debug("Interface details",
			zap.String("subnet_name", iface.SubnetName),
			zap.String("cidr", iface.CIDR),
			zap.String("mac_address", iface.MacAddress),
			zap.String("port_id", iface.PortID),
			zap.String("network_id", iface.NetworkID),
			zap.String("cr_namespace", iface.CRNamespace),
			zap.String("cr_name", iface.CRName),
			zap.Bool("netplan_success", iface.NetplanSuccess),
			zap.String("status", iface.Status),
		)
	}

	// Netplan 기능 적용
	success := processNetplanConfiguration(nodeName, interfaces, logger)

	// 처리 결과를 DB에 업데이트
	if err := updateNetplanStatus(dbClient, nodeName, interfaces, success, logger); err != nil {
		logger.Error("Failed to update netplan status in database", zap.Error(err))
	}

	return nil
}

// processNetplanConfiguration processes netplan configuration for the given interfaces
func processNetplanConfiguration(nodeName string, interfaces []database.NodeInterface, logger *zap.Logger) bool {
	// NetplanManager 생성 (DRY_RUN 환경변수로 제어)
	dryRun := os.Getenv("DRY_RUN") == "true"
	if dryRun {
		logger.Info("Running in DRY RUN mode - netplan files will not be applied")
	}

	netplanManager := netplan.NewNetplanManager(logger, dryRun)

	// database.NodeInterface를 netplan.InterfaceData로 변환
	netplanInterfaces := make([]netplan.InterfaceData, 0, len(interfaces))
	for _, iface := range interfaces {
		netplanInterfaces = append(netplanInterfaces, netplan.InterfaceData{
			PortID:         iface.PortID,
			MACAddress:     iface.MacAddress,
			SubnetName:     iface.SubnetName,
			CIDR:           iface.CIDR,
			NetworkID:      iface.NetworkID,
			NetplanSuccess: iface.NetplanSuccess,
		})
	}

	// Netplan 구성 처리
	if err := netplanManager.ProcessInterfaces(nodeName, netplanInterfaces); err != nil {
		logger.Error("Failed to process netplan configuration",
			zap.String("node", nodeName),
			zap.Error(err))
		return false
	}

	return true
}

// updateNetplanStatus updates the netplan status in the database
func updateNetplanStatus(dbClient *database.Client, nodeName string, interfaces []database.NodeInterface, success bool, logger *zap.Logger) error {
	for _, iface := range interfaces {
		// 상태가 변경된 경우에만 업데이트
		if iface.NetplanSuccess != success {
			if err := dbClient.UpdateNetplanSuccess(iface.PortID, success); err != nil {
				logger.Error("Failed to update netplan status for interface",
					zap.String("port_id", iface.PortID),
					zap.Bool("success", success),
					zap.Error(err))
				return err
			}

			logger.Info("Updated netplan status",
				zap.String("port_id", iface.PortID),
				zap.Bool("success", success))
		}
	}

	return nil
}
