package netplan

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// NetplanConfig represents the netplan configuration structure
type NetplanConfig struct {
	Network NetworkConfig `yaml:"network"`
}

type NetworkConfig struct {
	Version   int                          `yaml:"version"`
	Renderer  string                       `yaml:"renderer,omitempty"`
	Ethernets map[string]EthernetInterface `yaml:"ethernets"`
}

type EthernetInterface struct {
	Match     *MatchConfig `yaml:"match,omitempty"`
	SetName   string       `yaml:"set-name,omitempty"`
	DHCP4     *bool        `yaml:"dhcp4,omitempty"`
	MTU       int          `yaml:"mtu,omitempty"`
	Addresses []string     `yaml:"addresses,omitempty"`
	Routes    []Route      `yaml:"routes,omitempty"`
}

type MatchConfig struct {
	MACAddress string `yaml:"macaddress,omitempty"`
	Driver     string `yaml:"driver,omitempty"`
}

type Route struct {
	To     string `yaml:"to"`
	Via    string `yaml:"via"`
	Metric int    `yaml:"metric,omitempty"`
}

// InterfaceData represents database interface information
type InterfaceData struct {
	PortID         string
	MACAddress     string
	SubnetName     string
	CIDR           string
	NetworkID      string
	NetplanSuccess bool
}

// NetplanManager manages netplan configuration
type NetplanManager struct {
	logger      *zap.Logger
	netplanDir  string
	backupDir   string
	dryRun      bool
	defaultGW   string
	nameservers []string
}

// NewNetplanManager creates a new NetplanManager
func NewNetplanManager(logger *zap.Logger, dryRun bool) *NetplanManager {
	return &NetplanManager{
		logger:      logger,
		netplanDir:  "/etc/netplan",
		backupDir:   "/var/backups/netplan",
		dryRun:      dryRun,
		defaultGW:   "10.0.0.1",                     // Default gateway - should be configurable
		nameservers: []string{"8.8.8.8", "8.8.4.4"}, // Default DNS - should be configurable
	}
}

// GenerateNetplanConfig generates netplan configuration for given interfaces
func (nm *NetplanManager) GenerateNetplanConfig(nodeName string, interfaces []InterfaceData) (*NetplanConfig, error) {
	config := &NetplanConfig{
		Network: NetworkConfig{
			Version:   2,
			Ethernets: make(map[string]EthernetInterface),
		},
	}

	for i, iface := range interfaces {
		interfaceName := fmt.Sprintf("eth%d", i+1) // eth1, eth2, etc.

		dhcp4 := true
		ethernet := EthernetInterface{
			Match: &MatchConfig{
				MACAddress: strings.ToLower(iface.MACAddress),
			},
			SetName: interfaceName,
			DHCP4:   &dhcp4,
			MTU:     1450,
		}

		config.Network.Ethernets[interfaceName] = ethernet

		nm.logger.Info("Configured interface for DHCP",
			zap.String("interface", interfaceName),
			zap.String("mac", iface.MACAddress))
	}

	return config, nil
}

// WriteNetplanFile writes the netplan configuration to a file
func (nm *NetplanManager) WriteNetplanFile(nodeName string, config *NetplanConfig) error {
	filename := fmt.Sprintf("99-multinic-%s.yaml", nodeName)
	filePath := filepath.Join(nm.netplanDir, filename)

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(nm.backupDir, 0755); err != nil {
		nm.logger.Error("Failed to create backup directory",
			zap.String("path", nm.backupDir),
			zap.Error(err))
	}

	// Backup existing file if it exists
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filepath.Join(nm.backupDir, fmt.Sprintf("%s.%d", filename, time.Now().Unix()))
		if err := nm.copyFile(filePath, backupPath); err != nil {
			nm.logger.Warn("Failed to backup existing netplan file",
				zap.String("source", filePath),
				zap.String("backup", backupPath),
				zap.Error(err))
		} else {
			nm.logger.Info("Backed up existing netplan file",
				zap.String("backup", backupPath))
		}
	}

	// Marshal config to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal netplan config: %w", err)
	}

	if nm.dryRun {
		nm.logger.Info("DRY RUN: Would write netplan file",
			zap.String("file", filePath),
			zap.String("content", string(yamlData)))
		return nil
	}

	// Write YAML to file with correct permissions (600)
	if err := os.WriteFile(filePath, yamlData, 0600); err != nil {
		return fmt.Errorf("failed to write netplan file: %w", err)
	}

	nm.logger.Info("Successfully wrote netplan file",
		zap.String("file", filePath))

	return nil
}

// ApplyNetplan applies the netplan configuration
func (nm *NetplanManager) ApplyNetplan() error {
	if nm.dryRun {
		nm.logger.Info("DRY RUN: Would apply netplan configuration")
		return nil
	}

	// Try validation first
	if err := nm.ValidateNetplan(); err != nil {
		return fmt.Errorf("netplan validation failed: %w", err)
	}

	// Check if running in privileged mode with host network
	if nm.isRunningInContainer() && !nm.isPrivilegedMode() {
		nm.logger.Info("Running in non-privileged container environment - skipping netplan apply")
		return nil
	}

	// Apply netplan configuration
	nm.logger.Info("Applying netplan configuration...")

	var cmd *exec.Cmd

	// If in container with privileged mode, try nsenter to run in host namespace
	if nm.isRunningInContainer() && nm.isPrivilegedMode() {
		nm.logger.Info("Using nsenter to run netplan apply in host namespace")
		// nsenter -t 1 -m -u -n -i netplan apply
		cmd = exec.Command("nsenter", "-t", "1", "-m", "-u", "-n", "-i", "netplan", "apply")
	} else {
		// Direct execution (for non-container environment)
		cmd = exec.Command("timeout", "60", "netplan", "apply")
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If nsenter/netplan apply fails, try alternative approaches
		nm.logger.Error("Failed to apply netplan configuration with primary method",
			zap.Error(err),
			zap.String("stdout", stdout.String()),
			zap.String("stderr", stderr.String()))

		// Try alternative: use systemd-run if available
		if nm.isRunningInContainer() && nm.isPrivilegedMode() {
			nm.logger.Info("Trying alternative method with systemd-run...")
			altCmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-n", "-i", "systemd-run", "--no-block", "netplan", "apply")
			var altStdout, altStderr bytes.Buffer
			altCmd.Stdout = &altStdout
			altCmd.Stderr = &altStderr

			if altErr := altCmd.Run(); altErr == nil {
				nm.logger.Info("Successfully applied netplan with systemd-run",
					zap.String("output", altStdout.String()))
				return nil
			} else {
				nm.logger.Warn("systemd-run method also failed",
					zap.Error(altErr),
					zap.String("stdout", altStdout.String()),
					zap.String("stderr", altStderr.String()))
			}
		}

		// Try fallback: generate only
		nm.logger.Info("Falling back to netplan generate only...")
		fallbackCmd := exec.Command("netplan", "generate")
		var fallbackStdout, fallbackStderr bytes.Buffer
		fallbackCmd.Stdout = &fallbackStdout
		fallbackCmd.Stderr = &fallbackStderr

		if fallbackErr := fallbackCmd.Run(); fallbackErr != nil {
			nm.logger.Error("Fallback netplan generate also failed",
				zap.Error(fallbackErr),
				zap.String("stdout", fallbackStdout.String()),
				zap.String("stderr", fallbackStderr.String()))
			return fmt.Errorf("all netplan methods failed: primary_err=%w, generate_err=%v", err, fallbackErr)
		}

		nm.logger.Info("Fallback netplan generate succeeded",
			zap.String("output", fallbackStdout.String()))

		// Try to manually apply network configuration using ip commands
		if nm.isRunningInContainer() && nm.isPrivilegedMode() {
			nm.logger.Info("Attempting manual network configuration using ip commands...")
			if applyErr := nm.applyNetworkManually(); applyErr != nil {
				nm.logger.Warn("Manual network configuration failed", zap.Error(applyErr))
			} else {
				nm.logger.Info("Successfully applied network configuration manually")
			}
		}

		return nil
	}

	nm.logger.Info("Successfully applied netplan configuration",
		zap.String("output", stdout.String()))

	return nil
}

// ValidateNetplan validates the netplan configuration
func (nm *NetplanManager) ValidateNetplan() error {
	if nm.dryRun {
		nm.logger.Info("DRY RUN: Would validate netplan configuration")
		return nil
	}

	cmd := exec.Command("netplan", "generate")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		nm.logger.Error("Netplan validation failed",
			zap.Error(err),
			zap.String("stdout", stdout.String()),
			zap.String("stderr", stderr.String()))
		return fmt.Errorf("netplan validation failed: %w", err)
	}

	nm.logger.Info("Netplan configuration is valid")
	return nil
}

// copyFile copies a file from src to dst
func (nm *NetplanManager) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// ProcessInterfaces processes interfaces and applies netplan configuration
func (nm *NetplanManager) ProcessInterfaces(nodeName string, interfaces []InterfaceData) error {
	nm.logger.Info("Processing interfaces for netplan configuration",
		zap.String("node", nodeName),
		zap.Int("interface_count", len(interfaces)))

	// Debug: Check actual host interfaces before generating config
	if err := nm.debugHostInterfaces(); err != nil {
		nm.logger.Warn("Failed to debug host interfaces", zap.Error(err))
	}

	// Generate netplan configuration
	config, err := nm.GenerateNetplanConfig(nodeName, interfaces)
	if err != nil {
		return fmt.Errorf("failed to generate netplan config: %w", err)
	}

	// Write configuration to file
	if err := nm.WriteNetplanFile(nodeName, config); err != nil {
		return fmt.Errorf("failed to write netplan file: %w", err)
	}

	// Validate configuration
	if err := nm.ValidateNetplan(); err != nil {
		return fmt.Errorf("netplan validation failed: %w", err)
	}

	// Apply configuration
	if err := nm.ApplyNetplan(); err != nil {
		return fmt.Errorf("failed to apply netplan: %w", err)
	}

	nm.logger.Info("Successfully processed interfaces and applied netplan configuration",
		zap.String("node", nodeName))

	return nil
}

// isRunningInContainer detects if we're running in a container
func (nm *NetplanManager) isRunningInContainer() bool {
	// Check for container environment indicators
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for Kubernetes environment
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	return false
}

// isPrivilegedMode checks if running in privileged mode
func (nm *NetplanManager) isPrivilegedMode() bool {
	// Check if we can access host /proc/1 (init process)
	if _, err := os.Stat("/proc/1/root"); err == nil {
		return true
	}

	// Check if we have NET_ADMIN capability
	if _, err := os.Stat("/proc/self/status"); err == nil {
		if data, err := os.ReadFile("/proc/self/status"); err == nil {
			content := string(data)
			// Look for CapEff (effective capabilities)
			if strings.Contains(content, "CapEff:") {
				// If we have significant capabilities, likely privileged
				return strings.Contains(content, "CapEff:\t0000003fffffffff") ||
					strings.Contains(content, "CapEff:\t000001ffffffffff")
			}
		}
	}

	// Check environment variable that indicates privileged mode
	return os.Getenv("PRIVILEGED_MODE") == "true"
}

// applyNetworkManually applies network configuration using ip commands directly
func (nm *NetplanManager) applyNetworkManually() error {
	nm.logger.Info("Attempting to read and parse generated netplan files...")

	// Try to use networkctl reload if available
	cmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-n", "-i", "networkctl", "reload")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		nm.logger.Warn("networkctl reload failed",
			zap.Error(err),
			zap.String("stdout", stdout.String()),
			zap.String("stderr", stderr.String()))

		// Try alternative: restart systemd-networkd directly
		restartCmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-n", "-i", "systemctl", "restart", "systemd-networkd")
		var restartStdout, restartStderr bytes.Buffer
		restartCmd.Stdout = &restartStdout
		restartCmd.Stderr = &restartStderr

		if restartErr := restartCmd.Run(); restartErr != nil {
			nm.logger.Warn("systemctl restart systemd-networkd also failed",
				zap.Error(restartErr),
				zap.String("stdout", restartStdout.String()),
				zap.String("stderr", restartStderr.String()))
			return fmt.Errorf("both networkctl reload and systemctl restart failed: %w", err)
		} else {
			nm.logger.Info("Successfully restarted systemd-networkd",
				zap.String("output", restartStdout.String()))
			return nil
		}
	} else {
		nm.logger.Info("Successfully reloaded network configuration",
			zap.String("output", stdout.String()))
		return nil
	}
}

// debugHostInterfaces checks and logs actual host network interfaces
func (nm *NetplanManager) debugHostInterfaces() error {
	nm.logger.Info("=== Debugging Host Network Interfaces ===")

	// Method 1: Check /sys/class/net
	if interfaces, err := nm.getSystemInterfaces(); err == nil {
		for _, iface := range interfaces {
			nm.logger.Info("Found system interface",
				zap.String("name", iface.Name),
				zap.String("mac", iface.MAC),
				zap.String("state", iface.State))
		}
	} else {
		nm.logger.Warn("Failed to get system interfaces", zap.Error(err))
	}

	// Method 2: Use ip command if available
	if nm.isRunningInContainer() && nm.isPrivilegedMode() {
		nm.logger.Info("Checking interfaces using ip command in host namespace...")
		cmd := exec.Command("nsenter", "-t", "1", "-n", "ip", "link", "show")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			nm.logger.Warn("Failed to run ip link show",
				zap.Error(err),
				zap.String("stderr", stderr.String()))
		} else {
			nm.logger.Info("Host network interfaces (ip link show)",
				zap.String("output", stdout.String()))
		}
	}

	return nil
}

// SystemInterface represents a system network interface
type SystemInterface struct {
	Name  string
	MAC   string
	State string
}

// getSystemInterfaces reads network interfaces from /sys/class/net
func (nm *NetplanManager) getSystemInterfaces() ([]SystemInterface, error) {
	var interfaces []SystemInterface

	netDir := "/sys/class/net"
	entries, err := os.ReadDir(netDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", netDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip loopback and virtual interfaces
		if name == "lo" || strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "docker") {
			continue
		}

		mac, _ := os.ReadFile(filepath.Join(netDir, name, "address"))
		state, _ := os.ReadFile(filepath.Join(netDir, name, "operstate"))

		iface := SystemInterface{
			Name:  name,
			MAC:   strings.TrimSpace(string(mac)),
			State: strings.TrimSpace(string(state)),
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}
