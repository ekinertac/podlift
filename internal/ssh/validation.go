package ssh

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ServiceInfo represents information about a deployed service
type ServiceInfo struct {
	Name       string
	Containers []string
	DeployedAt string
	Version    string
}

// CheckExistingService checks if a service is already deployed on the server
func (c *Client) CheckExistingService(serviceName string) (bool, *ServiceInfo, error) {
	// Query Docker for containers with podlift.service label
	cmd := fmt.Sprintf(
		`docker ps -a --filter "label=podlift.service=%s" --format "{{.Names}},{{.Label \"podlift.version\"}},{{.Label \"podlift.deployed_at\"}}"`,
		serviceName,
	)

	output, err := c.Execute(cmd)
	if err != nil {
		// Docker might not be installed or no containers found
		return false, nil, nil
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return false, nil, nil
	}

	// Parse output
	lines := strings.Split(output, "\n")
	containers := make([]string, 0)
	var version, deployedAt string

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			containers = append(containers, parts[0])
		}
		if len(parts) >= 2 && version == "" {
			version = parts[1]
		}
		if len(parts) >= 3 && deployedAt == "" {
			deployedAt = parts[2]
		}
	}

	if len(containers) == 0 {
		return false, nil, nil
	}

	info := &ServiceInfo{
		Name:       serviceName,
		Containers: containers,
		Version:    version,
		DeployedAt: deployedAt,
	}

	return true, info, nil
}

// ListPodliftServices lists all podlift-managed services on the server
func (c *Client) ListPodliftServices() ([]string, error) {
	cmd := `docker ps -a --filter "label=podlift.service" --format "{{.Label \"podlift.service\"}}" | sort -u`

	output, err := c.Execute(cmd)
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}

	services := strings.Split(output, "\n")
	
	// Filter out empty strings
	result := make([]string, 0)
	for _, svc := range services {
		if svc != "" {
			result = append(result, svc)
		}
	}

	return result, nil
}

// CheckServiceNameConflict checks if service name conflicts with existing deployments
func (c *Client) CheckServiceNameConflict(serviceName string) error {
	exists, info, err := c.CheckExistingService(serviceName)
	if err != nil {
		return fmt.Errorf("failed to check for existing service: %w", err)
	}

	if !exists {
		return nil // No conflict
	}

	// Service exists - this is fine for redeployment of the same app
	// But we should warn if it looks suspicious (e.g., different repo)
	return &ServiceConflictError{
		ServiceName: serviceName,
		Info:        info,
	}
}

// ServiceConflictError represents a service name conflict
type ServiceConflictError struct {
	ServiceName string
	Info        *ServiceInfo
}

func (e *ServiceConflictError) Error() string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("service '%s' is already deployed on this server\n\n", e.ServiceName))
	msg.WriteString("Existing deployment:\n")
	msg.WriteString(fmt.Sprintf("  Version: %s\n", e.Info.Version))
	if e.Info.DeployedAt != "" {
		msg.WriteString(fmt.Sprintf("  Deployed: %s\n", e.Info.DeployedAt))
	}
	msg.WriteString(fmt.Sprintf("  Containers: %d\n", len(e.Info.Containers)))
	msg.WriteString("\nIf this is the same application:\n")
	msg.WriteString("  This is normal - proceeding will update the existing deployment\n")
	msg.WriteString("\nIf this is a DIFFERENT application:\n")
	msg.WriteString("  Change the service name in podlift.yml to avoid conflicts\n")
	msg.WriteString("\nService names must be unique per server.")
	
	return msg.String()
}

// IsConflictError checks if error is a service conflict
func IsConflictError(err error) bool {
	_, ok := err.(*ServiceConflictError)
	return ok
}

// GetStateFile reads the podlift state file from the server
func (c *Client) GetStateFile(serviceName string) (map[string]interface{}, error) {
	statePath := fmt.Sprintf("/opt/%s/.podlift/state.json", serviceName)
	
	cmd := fmt.Sprintf("cat %s 2>/dev/null || echo '{}'", statePath)
	output, err := c.Execute(cmd)
	if err != nil {
		return nil, err
	}

	var state map[string]interface{}
	if err := json.Unmarshal([]byte(output), &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return state, nil
}

// CompareGitRepo compares current git repo with deployed state
// Returns true if repos are the same, false if different
func (c *Client) CompareGitRepo(serviceName, currentRepo string) (bool, error) {
	state, err := c.GetStateFile(serviceName)
	if err != nil {
		return false, err
	}

	// If no current deployment, can't compare
	if len(state) == 0 {
		return true, nil
	}

	// Get git repo from state
	current, ok := state["current"].(map[string]interface{})
	if !ok {
		return true, nil
	}

	deployedRepo, ok := current["git_repo"].(string)
	if !ok {
		return true, nil
	}

	// Normalize repo URLs for comparison
	currentNorm := normalizeGitURL(currentRepo)
	deployedNorm := normalizeGitURL(deployedRepo)

	return currentNorm == deployedNorm, nil
}

// normalizeGitURL normalizes git URLs for comparison
func normalizeGitURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "git@github.com:")
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.ToLower(url)
	return url
}

