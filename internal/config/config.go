package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the CLI configuration persisted on disk.
type Config struct {
	// APIURL is the base URL of the Repo API (e.g. https://devtools.api.cloud.ru/repo/api/v1).
	APIURL string `json:"api_url"`
	// ProjectID is the default project ID used for API calls.
	ProjectID string `json:"project_id"`
	// APIKey is the X-API-KEY value used for authentication (shared across products).
	APIKey string `json:"api_key"`
	// GitHost is the host name used for git smart HTTP operations
	// (e.g. repo.cloud.ru). It is separate from APIURL.
	GitHost string `json:"git_host"`
	// WorkflowAPIURL is the base URL of the Workflow Studio user API.
	WorkflowAPIURL string `json:"workflow_api_url"`
}

// DefaultAPIURL is the production Repo API URL.
const DefaultAPIURL = "https://devtools.api.cloud.ru/repo/api/v1"

// DefaultGitHost is the production git smart HTTP host.
const DefaultGitHost = "https://repo.cloud.ru/"

// DefaultWorkflowAPIURL is the Workflow Studio user API base URL.
const DefaultWorkflowAPIURL = "https://devtools.api.cloud.ru/workflow/v1"

// Load reads the config from disk, applies defaults and environment overrides.
// Environment variables take precedence over file values.
//
// Evolution DevServices (eds) is the platform; Repo and Workflow Studio are products on it, so
// their credentials/hosts are namespaced per-product. Only truly
// platform-level settings (project id, config path) are unprefixed.
//
//	EDS_PROJECT_ID    - overrides project_id (shared across products)
//	EDS_REPO_API_URL  - overrides api_url (Repo product)
//	EDS_API_KEY     - overrides api_key (shared across products)
//	EDS_REPO_GIT_HOST - overrides git_host (Repo product)
//	EDS_WF_API_URL    - overrides workflow_api_url (Workflow Studio product)
func Load() (*Config, error) {
	cfgPath, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		APIURL:         DefaultAPIURL,
		GitHost:        DefaultGitHost,
		WorkflowAPIURL: DefaultWorkflowAPIURL,
		ProjectID:      os.Getenv("EDS_PROJECT_ID"),
	}

	if data, err := os.ReadFile(cfgPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", cfgPath, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config %s: %w", cfgPath, err)
	}

	// Defaults after loading file so the file can override defaults.
	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}
	if cfg.GitHost == "" {
		cfg.GitHost = DefaultGitHost
	}
	if cfg.WorkflowAPIURL == "" {
		cfg.WorkflowAPIURL = DefaultWorkflowAPIURL
	}

	if v := os.Getenv("EDS_PROJECT_ID"); v != "" {
		cfg.ProjectID = v
	}
	if v := os.Getenv("EDS_REPO_API_URL"); v != "" {
		cfg.APIURL = v
	}
	if v := os.Getenv("EDS_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("EDS_REPO_GIT_HOST"); v != "" {
		cfg.GitHost = v
	}
	if v := os.Getenv("EDS_WF_API_URL"); v != "" {
		cfg.WorkflowAPIURL = v
	}

	return cfg, nil
}

// Save persists the config to disk with 0600 permissions.
func (c *Config) Save() error {
	cfgPath, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ") //nolint:gosec // config file intentionally stores API key
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", cfgPath, err)
	}
	return nil
}

// configPath returns the absolute path to the config file. This is a
// platform-level (not per-product) file, since a single project id
// setup is shared across products. Honours EDS_CONFIG and XDG_CONFIG_HOME
// for non-Linux convenience.
func configPath() (string, error) {
	if v := os.Getenv("EDS_CONFIG"); v != "" {
		return expand(v), nil
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "eds", "config.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", "eds", "config.json"), nil
}

func expand(p string) string {
	if after, ok := strings.CutPrefix(p, "~"); ok {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, after)
		}
	}
	return p
}
