package containerimagelisting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caarlos0/env/v6"
)

type TagTest struct {
	Repository   string
	ExpectedTags []string
}

type IntegrationTestConfig struct {
	GhcrUsername      string    `env:"GHCR_USERNAME"`
	GhcrPassword      string    `env:"GHCR_PASSWORD"`
	GhcrTests         []TagTest `env:"GHCR_TESTS"`
	DockerhubUsername string    `env:"DOCKERHUB_USERNAME"`
	DockerhubPassword string    `env:"DOCKERHUB_PASSWORD"`
	DockerhubTests    []TagTest `env:"DOCKERHUB_TESTS"`
	ECRTests          []TagTest `env:"ECR_TESTS"`
	ECRBaseURL        string    `env:"ECR_BASE_URL"`
	QuayToken         string    `env:"QUAY_TOKEN"`
	QuayTests         []TagTest `env:"QUAY_TESTS"`
}

func LoadIntegrationTestConfig() (*IntegrationTestConfig, error) {
	cfgFile := os.Getenv("CONTAINER_TEST_CONFIG_FILE")
	if cfgFile == "" {
		cfgFile = "testing_config.json"
	}
	bytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load config file %s: %w", cfgFile, err)
	}
	var ret IntegrationTestConfig
	if err := json.Unmarshal(bytes, &ret); err != nil {
		return nil, fmt.Errorf("file does not appear to be JSON: %w", err)
	}
	if err := env.Parse(&ret); err != nil {
		return nil, fmt.Errorf("unable to parse config from env: %w", err)
	}
	return &ret, nil
}
