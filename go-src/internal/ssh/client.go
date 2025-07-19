package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client connection
type Client struct {
	config *ssh.ClientConfig
	client *ssh.Client
	host   string
	port   string
}

// Config represents SSH connection configuration
type Config struct {
	Host       string
	Port       string
	Username   string
	Password   string
	PrivateKey string
	KeyPath    string
	Timeout    time.Duration
}

// NewClient creates a new SSH client
func NewClient(config *Config) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Create SSH client config
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use ssh.FixedHostKey()
		Timeout:         config.Timeout,
	}

	// Add authentication method
	if config.Password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(config.Password))
	}

	if config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	if config.KeyPath != "" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	// Establish connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", config.Host, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s:%s: %w", config.Host, config.Port, err)
	}

	logrus.Infof("✅ SSH connection established to %s:%s", config.Host, config.Port)

	return &Client{
		config: sshConfig,
		client: client,
		host:   config.Host,
		port:   config.Port,
	}, nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ExecuteCommand executes a command on the remote server
func (c *Client) ExecuteCommand(command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// ExecuteCommandWithOutput executes a command and returns stdout and stderr separately
func (c *Client) ExecuteCommandWithOutput(command string) (string, string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	if err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("command failed: %w", err)
	}

	return stdout.String(), stderr.String(), nil
}

// ExecuteCommandStream executes a command and streams the output
func (c *Client) ExecuteCommandStream(command string, stdout, stderr io.Writer) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(command)
}

// UploadFile uploads a file to the remote server
func (c *Client) UploadFile(localPath, remotePath string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Create remote file
	remoteFile, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start scp command
	go func() {
		defer remoteFile.Close()
		fmt.Fprintf(remoteFile, "C0644 %s %s\n", filepath.Base(remotePath), filepath.Base(remotePath))
		io.Copy(remoteFile, file)
		fmt.Fprint(remoteFile, "\x00")
	}()

	// Execute scp command
	err = session.Run(fmt.Sprintf("scp -t %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	logrus.Infof("✅ File uploaded: %s -> %s", localPath, remotePath)
	return nil
}

// DownloadFile downloads a file from the remote server
func (c *Client) DownloadFile(remotePath, localPath string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Get remote file
	remoteFile, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start scp command
	go func() {
		scanner := bufio.NewScanner(remoteFile)
		scanner.Scan() // Skip header line
		io.Copy(file, remoteFile)
	}()

	// Execute scp command
	err = session.Run(fmt.Sprintf("scp -f %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	logrus.Infof("✅ File downloaded: %s -> %s", remotePath, localPath)
	return nil
}

// TestConnection tests the SSH connection
func (c *Client) TestConnection() error {
	_, err := c.ExecuteCommand("echo 'SSH connection test successful'")
	return err
}

// GetHostInfo returns basic host information
func (c *Client) GetHostInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Get OS info
	if output, err := c.ExecuteCommand("uname -a"); err == nil {
		info["os"] = output
	}

	// Get hostname
	if output, err := c.ExecuteCommand("hostname"); err == nil {
		info["hostname"] = output
	}

	// Get uptime
	if output, err := c.ExecuteCommand("uptime"); err == nil {
		info["uptime"] = output
	}

	// Get memory info
	if output, err := c.ExecuteCommand("free -h"); err == nil {
		info["memory"] = output
	}

	// Get disk info
	if output, err := c.ExecuteCommand("df -h"); err == nil {
		info["disk"] = output
	}

	return info, nil
}
