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
			netplanIcon := "❌"
			if iface.NetplanSuccess {
				netplanIcon = "✅"
			}

			fmt.Printf("  • %s: %s %s\n",
				iface.SubnetName,
				iface.MacAddress,
				netplanIcon,
			)
			fmt.Printf("    ├─ CIDR: %s\n", iface.CIDR)
			fmt.Printf("    ├─ Port ID: %s\n", iface.PortID)
			fmt.Printf("    ├─ Network ID: %s\n", iface.NetworkID)
			fmt.Printf("    ├─ CR: %s/%s\n", iface.CRNamespace, iface.CRName)
			fmt.Printf("    ├─ Netplan Applied: %t\n", iface.NetplanSuccess)
			fmt.Printf("    └─ Status: %s\n", iface.Status)
		}
		fmt.Println()
	}

	// UpdateNetplanSuccess 기능 테스트
	fmt.Println("=== 🔧 Testing UpdateNetplanSuccess ===")
	if len(testNodes) > 0 {
		interfaces, err := dbClient.GetNodeInterfaces(testNodes[0])
		if err == nil && len(interfaces) > 0 {
			testInterface := interfaces[0]
			fmt.Printf("Updating netplan success for port %s...\n", testInterface.PortID)

			// true로 업데이트
			err := dbClient.UpdateNetplanSuccess(testInterface.PortID, true)
			if err != nil {
				log.Printf("Failed to update netplan success: %v", err)
			} else {
				fmt.Printf("✅ Successfully updated netplan success to true\n")
			}
		}
	}
}
