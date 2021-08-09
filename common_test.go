package containerimagelisting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func containsTag(name string, tags []Tag) bool {
	for _, tag := range tags {
		if tag.Name == name {
			return true
		}
	}
	return false
}

func TestNewContainerClient(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantedClient ContainerClient
	}{
		// Test Cases
		{
			name:         "quay",
			url:          "https://quay.io/cresta/chatmon",
			wantedClient: &QuayClient{},
		},
		{
			name:         "dockerhub",
			url:          "https://docker.io",
			wantedClient: &DockerRegistryClient{},
		},
	}

	for _, tt := range tests {
		client, err := NewClientFromEnv(tt.url)
		assert.NoError(t, err)
		assert.IsType(t, tt.wantedClient, client)
	}
}
