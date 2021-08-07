package containerimagelisting_test

import (
	"testing"

	containerimagelisting "github.com/cresta/container-image-listing"
	"github.com/stretchr/testify/assert"
)

func containsTag(name string, tags []containerimagelisting.Tag) bool {
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
		wantedClient containerimagelisting.ContainerClient
	}{
		// Test Cases
		{
			name:         "quay",
			url:          "quay.io/cresta/chatmon",
			wantedClient: &containerimagelisting.QuayClient{},
		},
		{
			name:         "dockerhub",
			url:          "docker.io", // TODO add better url
			wantedClient: &containerimagelisting.DockerHubClient{},
		},
	}

	for _, tt := range tests {
		client := containerimagelisting.NewClientFromEnv(tt.url)
		assert.IsType(t, tt.wantedClient, client)
	}
}
