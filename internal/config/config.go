package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config는 에이전트의 전체 설정을 담는 구조체입니다
type Config struct {
	Database   DatabaseConfig   `yaml:"database"`
	Agent      AgentConfig      `yaml:"agent"`
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
	Netplan    NetplanConfig    `yaml:"netplan"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// DatabaseConfig는 데이터베이스 연결 설정입니다
type DatabaseConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Database  string `yaml:"database"`
	Charset   string `yaml:"charset"`
	ParseTime bool   `yaml:"parse_time"`
	Loc       string `yaml:"loc"`
}

// AgentConfig는 에이전트 동작 설정입니다
type AgentConfig struct {
	CheckInterval int    `yaml:"check_interval"`
	RetryCount    int    `yaml:"retry_count"`
	RetryInterval int    `yaml:"retry_interval"`
	NodeName      string `yaml:"node_name"`
}

// KubernetesConfig는 Kubernetes 관련 설정입니다
type KubernetesConfig struct {
	Kubeconfig       string `yaml:"kubeconfig"`
	LabelPrefix      string `yaml:"label_prefix"`
	AnnotationPrefix string `yaml:"annotation_prefix"`
}

// NetplanConfig는 Netplan 관련 설정입니다
type NetplanConfig struct {
	ConfigPath string `yaml:"config_path"`
	BackupPath string `yaml:"backup_path"`
	DryRun     bool   `yaml:"dry_run"`
}

// LoggingConfig는 로깅 관련 설정입니다
type LoggingConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

// Load는 설정을 로드합니다. 환경변수가 우선순위가 높습니다.
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// 1. 파일에서 로드 (있는 경우)
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("설정 파일 파싱 실패: %w", err)
		}
	}

	// 2. 환경변수로 오버라이드
	loadFromEnv(config)

	// 3. 기본값 설정
	setDefaults(config)

	return config, nil
}

// loadFromEnv는 환경변수에서 설정을 로드합니다
func loadFromEnv(config *Config) {
	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		config.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		config.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		config.Database.Password = v
	}
	if v := os.Getenv("DB_DATABASE"); v != "" {
		config.Database.Database = v
	}
	if v := os.Getenv("DB_CHARSET"); v != "" {
		config.Database.Charset = v
	}
	if v := os.Getenv("DB_PARSE_TIME"); v != "" {
		config.Database.ParseTime = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("DB_LOC"); v != "" {
		config.Database.Loc = v
	}

	// Agent
	if v := os.Getenv("AGENT_CHECK_INTERVAL"); v != "" {
		if interval, err := strconv.Atoi(v); err == nil {
			config.Agent.CheckInterval = interval
		}
	}
	if v := os.Getenv("AGENT_RETRY_COUNT"); v != "" {
		if count, err := strconv.Atoi(v); err == nil {
			config.Agent.RetryCount = count
		}
	}
	if v := os.Getenv("AGENT_RETRY_INTERVAL"); v != "" {
		if interval, err := strconv.Atoi(v); err == nil {
			config.Agent.RetryInterval = interval
		}
	}
	if v := os.Getenv("NODE_NAME"); v != "" {
		config.Agent.NodeName = v
	}

	// Kubernetes
	if v := os.Getenv("KUBECONFIG"); v != "" {
		config.Kubernetes.Kubeconfig = v
	}
	if v := os.Getenv("K8S_LABEL_PREFIX"); v != "" {
		config.Kubernetes.LabelPrefix = v
	}
	if v := os.Getenv("K8S_ANNOTATION_PREFIX"); v != "" {
		config.Kubernetes.AnnotationPrefix = v
	}

	// Netplan
	if v := os.Getenv("NETPLAN_CONFIG_PATH"); v != "" {
		config.Netplan.ConfigPath = v
	}
	if v := os.Getenv("NETPLAN_BACKUP_PATH"); v != "" {
		config.Netplan.BackupPath = v
	}
	if v := os.Getenv("NETPLAN_DRY_RUN"); v != "" {
		config.Netplan.DryRun = strings.ToLower(v) == "true"
	}

	// Logging
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		config.Logging.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		config.Logging.Format = v
	}
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		config.Logging.Output = v
	}
	if v := os.Getenv("LOG_FILE_PATH"); v != "" {
		config.Logging.FilePath = v
	}
}

// setDefaults는 기본값을 설정합니다
func setDefaults(config *Config) {
	// Database defaults
	if config.Database.Port == 0 {
		config.Database.Port = 3306
	}
	if config.Database.Charset == "" {
		config.Database.Charset = "utf8mb4"
	}
	if config.Database.ParseTime == false {
		config.Database.ParseTime = true
	}
	if config.Database.Loc == "" {
		config.Database.Loc = "UTC"
	}

	// Agent defaults
	if config.Agent.CheckInterval == 0 {
		config.Agent.CheckInterval = 30
	}
	if config.Agent.RetryCount == 0 {
		config.Agent.RetryCount = 3
	}
	if config.Agent.RetryInterval == 0 {
		config.Agent.RetryInterval = 5
	}

	// Kubernetes defaults
	if config.Kubernetes.LabelPrefix == "" {
		config.Kubernetes.LabelPrefix = "multinic.io"
	}
	if config.Kubernetes.AnnotationPrefix == "" {
		config.Kubernetes.AnnotationPrefix = "multinic.io"
	}

	// Netplan defaults
	if config.Netplan.ConfigPath == "" {
		config.Netplan.ConfigPath = "/etc/netplan"
	}
	if config.Netplan.BackupPath == "" {
		config.Netplan.BackupPath = "/var/backups/netplan"
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
}
