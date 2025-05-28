package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"github.com/ibyeong-geon/multinic-agent/internal/config"
)

// Client는 데이터베이스 클라이언트입니다
type Client struct {
	db     *sql.DB
	logger *zap.Logger
}

// NetworkInterface는 네트워크 인터페이스 정보입니다
type NetworkInterface struct {
	NodeName      string    `db:"node_name"`
	InterfaceName string    `db:"interface_name"`
	MacAddress    string    `db:"mac_address"`
	IpAddress     string    `db:"ip_address"`
	SubnetMask    string    `db:"subnet_mask"`
	NetworkID     string    `db:"network_id"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// NewClient는 새로운 데이터베이스 클라이언트를 생성합니다
func NewClient(cfg *config.DatabaseConfig, logger *zap.Logger) (*Client, error) {
	// MySQL DSN 생성
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)

	// 데이터베이스 연결
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 연결 테스트
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 연결 풀 설정
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return &Client{
		db:     db,
		logger: logger,
	}, nil
}

// Close는 데이터베이스 연결을 닫습니다
func (c *Client) Close() error {
	return c.db.Close()
}

// GetNodeInterfaces는 특정 노드의 네트워크 인터페이스 정보를 조회합니다
func (c *Client) GetNodeInterfaces(nodeName string) ([]NetworkInterface, error) {
	query := `
		SELECT 
			node_name,
			interface_name,
			mac_address,
			ip_address,
			subnet_mask,
			network_id,
			updated_at
		FROM network_interfaces
		WHERE node_name = ?
		ORDER BY interface_name
	`

	rows, err := c.db.Query(query, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to query interfaces: %w", err)
	}
	defer rows.Close()

	var interfaces []NetworkInterface
	for rows.Next() {
		var iface NetworkInterface
		err := rows.Scan(
			&iface.NodeName,
			&iface.InterfaceName,
			&iface.MacAddress,
			&iface.IpAddress,
			&iface.SubnetMask,
			&iface.NetworkID,
			&iface.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		interfaces = append(interfaces, iface)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	c.logger.Debug("Retrieved node interfaces",
		zap.String("node_name", nodeName),
		zap.Int("count", len(interfaces)),
	)

	return interfaces, nil
}
