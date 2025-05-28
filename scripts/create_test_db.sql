-- 데이터베이스 생성
CREATE DATABASE IF NOT EXISTS multinic_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE multinic_db;

-- 네트워크 인터페이스 테이블 생성
CREATE TABLE IF NOT EXISTS network_interfaces (
    id INT AUTO_INCREMENT PRIMARY KEY,
    node_name VARCHAR(255) NOT NULL,
    interface_name VARCHAR(50) NOT NULL,
    mac_address VARCHAR(17) NOT NULL,
    ip_address VARCHAR(15) NOT NULL,
    subnet_mask VARCHAR(15) NOT NULL,
    network_id VARCHAR(36) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_node_name (node_name),
    UNIQUE KEY unique_node_interface (node_name, interface_name)
);

-- 테스트 데이터 삽입
INSERT INTO network_interfaces (node_name, interface_name, mac_address, ip_address, subnet_mask, network_id) VALUES
-- worker-node-1의 인터페이스들
('worker-node-1', 'eth0', 'fa:16:3e:11:11:11', '10.0.0.11', '255.255.255.0', 'mgmt-network-uuid'),
('worker-node-1', 'eth1', 'fa:16:3e:22:22:22', '192.168.1.11', '255.255.255.0', 'data-network-uuid-1'),
('worker-node-1', 'eth2', 'fa:16:3e:33:33:33', '192.168.2.11', '255.255.255.0', 'data-network-uuid-2'),

-- worker-node-2의 인터페이스들
('worker-node-2', 'eth0', 'fa:16:3e:44:44:44', '10.0.0.12', '255.255.255.0', 'mgmt-network-uuid'),
('worker-node-2', 'eth1', 'fa:16:3e:55:55:55', '192.168.1.12', '255.255.255.0', 'data-network-uuid-1'),
('worker-node-2', 'eth2', 'fa:16:3e:66:66:66', '192.168.2.12', '255.255.255.0', 'data-network-uuid-2'),
('worker-node-2', 'eth3', 'fa:16:3e:77:77:77', '192.168.3.12', '255.255.255.0', 'data-network-uuid-3'); 