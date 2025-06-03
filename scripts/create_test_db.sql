-- 데이터베이스 생성
CREATE DATABASE IF NOT EXISTS multinic CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE multinic;

-- 기존 테이블 삭제 (스키마 변경으로 인한)
DROP TABLE IF EXISTS cr_state;
DROP TABLE IF EXISTS multi_interface;
DROP TABLE IF EXISTS node_table;
DROP TABLE IF EXISTS multi_subnet;

-- 서브넷 테이블 생성
CREATE TABLE IF NOT EXISTS multi_subnet (
    id INT AUTO_INCREMENT PRIMARY KEY,
    subnet_id VARCHAR(36) NOT NULL UNIQUE,
    subnet_name VARCHAR(255) NOT NULL,
    cidr VARCHAR(255) NOT NULL,
    network_id VARCHAR(36) NOT NULL COMMENT 'OpenStack network ID',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP NULL,
    modified_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);

-- 노드 테이블 생성
CREATE TABLE IF NOT EXISTS node_table (
    id INT AUTO_INCREMENT PRIMARY KEY,
    attached_node_id VARCHAR(36) NOT NULL UNIQUE,
    attached_node_name VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP NULL,
    modified_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);

-- 인터페이스 테이블 생성
CREATE TABLE IF NOT EXISTS multi_interface (
    id INT AUTO_INCREMENT PRIMARY KEY,
    port_id VARCHAR(36) NOT NULL UNIQUE,
    subnet_id VARCHAR(36) NOT NULL,
    macaddress VARCHAR(17) NOT NULL,
    attached_node_id VARCHAR(36),
    attached_node_name VARCHAR(255) NULL,
    cr_namespace VARCHAR(255) NOT NULL COMMENT 'OpenstackConfig CR namespace',
    cr_name VARCHAR(255) NOT NULL COMMENT 'OpenstackConfig CR name',
    status VARCHAR(50) DEFAULT 'active',
    netplan_success TINYINT(1) NOT NULL DEFAULT 0 COMMENT 'Netplan apply success (0: fail/not applied, 1: success)',
    created_at TIMESTAMP NULL,
    modified_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (subnet_id) REFERENCES multi_subnet(subnet_id),
    FOREIGN KEY (attached_node_id) REFERENCES node_table(attached_node_id),
    FOREIGN KEY (attached_node_name) REFERENCES node_table(attached_node_name),
    UNIQUE KEY unique_cr_interface (cr_namespace, cr_name, port_id)
);

-- CR 상태 테이블 생성
CREATE TABLE IF NOT EXISTS cr_state (
    id INT AUTO_INCREMENT PRIMARY KEY,
    cr_namespace VARCHAR(255) NOT NULL,
    cr_name VARCHAR(255) NOT NULL,
    spec_hash VARCHAR(64) NOT NULL COMMENT 'SHA256 hash of CR spec',
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_cr (cr_namespace, cr_name)
);

-- 테스트 데이터 삽입

-- 서브넷 데이터
INSERT INTO multi_subnet (subnet_id, subnet_name, cidr, network_id, created_at, modified_at) VALUES
('mgmt-subnet-uuid', 'Management Network', '10.0.0.0/24', 'mgmt-network-openstack-id', NOW(), NOW()),
('data-subnet-1-uuid', 'Data Network 1', '192.168.1.0/24', 'data-network-1-openstack-id', NOW(), NOW()),
('data-subnet-2-uuid', 'Data Network 2', '192.168.2.0/24', 'data-network-2-openstack-id', NOW(), NOW()),
('data-subnet-3-uuid', 'Data Network 3', '192.168.3.0/24', 'data-network-3-openstack-id', NOW(), NOW());

-- 노드 데이터 (실제 클러스터 노드 포함)
INSERT INTO node_table (attached_node_id, attached_node_name, created_at, modified_at) VALUES
('cluster2-control-plane-uuid', 'cluster2-control-plane', NOW(), NOW()),
('node-1-uuid', 'worker-node-1', NOW(), NOW()),
('node-2-uuid', 'worker-node-2', NOW(), NOW()),
('node-3-uuid', 'worker-node-3', NOW(), NOW());

-- 인터페이스 데이터
INSERT INTO multi_interface (port_id, subnet_id, macaddress, attached_node_id, attached_node_name, cr_namespace, cr_name, netplan_success, created_at, modified_at) VALUES
-- cluster2-control-plane의 인터페이스들 (실제 클러스터 노드)
('port-cp-1-uuid', 'mgmt-subnet-uuid', 'fa:16:3e:01:01:01', 'cluster2-control-plane-uuid', 'cluster2-control-plane', 'openstack-system', 'test-config-cp', 0, NOW(), NOW()),
('port-cp-2-uuid', 'data-subnet-1-uuid', 'fa:16:3e:01:01:02', 'cluster2-control-plane-uuid', 'cluster2-control-plane', 'openstack-system', 'test-config-cp', 0, NOW(), NOW()),
('port-cp-3-uuid', 'data-subnet-2-uuid', 'fa:16:3e:01:01:03', 'cluster2-control-plane-uuid', 'cluster2-control-plane', 'openstack-system', 'test-config-cp', 0, NOW(), NOW()),

-- worker-node-1의 인터페이스들
('port-1-1-uuid', 'mgmt-subnet-uuid', 'fa:16:3e:11:11:11', 'node-1-uuid', 'worker-node-1', 'openstack-system', 'test-config-1', 1, NOW(), NOW()),
('port-1-2-uuid', 'data-subnet-1-uuid', 'fa:16:3e:22:22:22', 'node-1-uuid', 'worker-node-1', 'openstack-system', 'test-config-1', 0, NOW(), NOW()),
('port-1-3-uuid', 'data-subnet-2-uuid', 'fa:16:3e:33:33:33', 'node-1-uuid', 'worker-node-1', 'openstack-system', 'test-config-1', 0, NOW(), NOW()),

-- worker-node-2의 인터페이스들
('port-2-1-uuid', 'mgmt-subnet-uuid', 'fa:16:3e:44:44:44', 'node-2-uuid', 'worker-node-2', 'openstack-system', 'test-config-2', 1, NOW(), NOW()),
('port-2-2-uuid', 'data-subnet-1-uuid', 'fa:16:3e:55:55:55', 'node-2-uuid', 'worker-node-2', 'openstack-system', 'test-config-2', 0, NOW(), NOW()),
('port-2-3-uuid', 'data-subnet-2-uuid', 'fa:16:3e:66:66:66', 'node-2-uuid', 'worker-node-2', 'openstack-system', 'test-config-2', 0, NOW(), NOW()),
('port-2-4-uuid', 'data-subnet-3-uuid', 'fa:16:3e:77:77:77', 'node-2-uuid', 'worker-node-2', 'openstack-system', 'test-config-2', 0, NOW(), NOW());

-- CR 상태 데이터
INSERT INTO cr_state (cr_namespace, cr_name, spec_hash) VALUES
('openstack-system', 'test-config-cp', 'cp123abc456def'),
('openstack-system', 'test-config-1', 'abc123def456789'),
('openstack-system', 'test-config-2', 'def456ghi789abc');

-- 데이터 확인
SELECT 
    n.attached_node_name,
    mi.port_id,
    mi.macaddress,
    ms.subnet_name,
    ms.cidr,
    mi.cr_namespace,
    mi.cr_name,
    mi.netplan_success,
    mi.status
FROM multi_interface mi
JOIN node_table n ON mi.attached_node_id = n.attached_node_id
JOIN multi_subnet ms ON mi.subnet_id = ms.subnet_id
WHERE mi.status = 'active'
ORDER BY n.attached_node_name, ms.subnet_name; 