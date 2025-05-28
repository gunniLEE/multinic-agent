package main

import (
	"fmt"
	"log"

	"github.com/ibyeong-geon/multinic-agent/internal/config"
	"github.com/ibyeong-geon/multinic-agent/pkg/database"
	"github.com/ibyeong-geon/multinic-agent/pkg/logger"
)

func main() {
	// 설정 로드
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 로거 초기화
	zapLogger, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// DB 연결 테스트
	fmt.Println("Connecting to database...")
	dbClient, err := database.NewClient(&cfg.Database, zapLogger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbClient.Close()

	fmt.Println("\n✅ Database connected successfully!\n")

	// 테스트용 노드 이름
	testNodes := []string{"worker-node-1", "worker-node-2", "worker-node-3"}

	for _, nodeName := range testNodes {
		fmt.Printf("=== 📍 Interfaces for %s ===\n", nodeName)

		interfaces, err := dbClient.GetNodeInterfaces(nodeName)
		if err != nil {
			log.Printf("Error getting interfaces for %s: %v", nodeName, err)
			continue
		}

		if len(interfaces) == 0 {
			fmt.Printf("❌ No interfaces found for %s\n", nodeName)
			continue
		}

		for _, iface := range interfaces {
			fmt.Printf("  • %s: %s (%s)\n",
				iface.InterfaceName,
				iface.IpAddress,
				iface.MacAddress,
			)
			fmt.Printf("    └─ Subnet: %s, Network ID: %s\n",
				iface.SubnetMask,
				iface.NetworkID,
			)
		}
		fmt.Println()
	}
}
