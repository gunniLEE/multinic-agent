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

// NodeInterface는 노드의 네트워크 인터페이스 정보입니다 (조인된 결과)
type NodeInterface struct {
	InterfaceID    int       `db:"interface_id"`
	PortID         string    `db:"port_id"`
	NodeID         string    `db:"node_id"`
	NodeName       string    `db:"node_name"`
	MacAddress     string    `db:"macaddress"`
	SubnetID       string    `db:"subnet_id"`
	SubnetName     string    `db:"subnet_name"`
	CIDR           string    `db:"cidr"`
	NetworkID      string    `db:"network_id"`
	CRNamespace    string    `db:"cr_namespace"`
	CRName         string    `db:"cr_name"`
	NetplanSuccess bool      `db:"netplan_success"`
	Status         string    `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
	ModifiedAt     time.Time `db:"modified_at"`
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
func (c *Client) GetNodeInterfaces(nodeName string) ([]NodeInterface, error) {
	query := `
		SELECT 
			mi.id as interface_id,
			mi.port_id,
			n.attached_node_id as node_id,
			n.attached_node_name as node_name,
			mi.macaddress,
			ms.subnet_id,
			ms.subnet_name,
			ms.cidr,
			ms.network_id,
			mi.cr_namespace,
			mi.cr_name,
			mi.netplan_success,
			mi.status,
			mi.created_at,
			mi.modified_at
		FROM multi_interface mi
		JOIN node_table n ON mi.attached_node_id = n.attached_node_id
		JOIN multi_subnet ms ON mi.subnet_id = ms.subnet_id
		WHERE n.attached_node_name = ? 
		  AND mi.status = 'active'
		  AND n.status = 'active'
		  AND ms.status = 'active'
		ORDER BY ms.subnet_name
	`

	rows, err := c.db.Query(query, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to query interfaces: %w", err)
	}
	defer rows.Close()

	var interfaces []NodeInterface
	for rows.Next() {
		var iface NodeInterface
		err := rows.Scan(
			&iface.InterfaceID,
			&iface.PortID,
			&iface.NodeID,
			&iface.NodeName,
			&iface.MacAddress,
			&iface.SubnetID,
			&iface.SubnetName,
			&iface.CIDR,
			&iface.NetworkID,
			&iface.CRNamespace,
			&iface.CRName,
			&iface.NetplanSuccess,
			&iface.Status,
			&iface.CreatedAt,
			&iface.ModifiedAt,
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

// UpdateNetplanSuccess는 특정 인터페이스의 netplan 적용 성공 여부를 업데이트합니다
func (c *Client) UpdateNetplanSuccess(portID string, success bool) error {
	query := `
		UPDATE multi_interface 
		SET netplan_success = ?, modified_at = NOW()
		WHERE port_id = ?
	`

	_, err := c.db.Exec(query, success, portID)
	if err != nil {
		return fmt.Errorf("failed to update netplan success: %w", err)
	}

	c.logger.Debug("Updated netplan success status",
		zap.String("port_id", portID),
		zap.Bool("success", success),
	)

	return nil
}
