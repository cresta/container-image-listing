package containerimagelisting_test

import (
	"testing"

	"github.com/cresta/container-image-listing"
	"github.com/stretchr/testify/assert"
)

func TestGHCRClient_ListTags(t *testing.T) {
	t.Parallel()

	ghcrClient := containerimagelisting.GHCRClient{}

	tags, err := ghcrClient.ListTags("homebrew/core/docker")
	assert.NoError(t, err)

	t.Logf("Tags: %s", tags)

	assert.Contains(t, tags, "20.10.7")
}

func TestGHCRWithDockerHub_WithAuth(t *testing.T) {
	t.Parallel()

	auth := containerimagelisting.Auth{}
	auth.FromEnv()

	client := &containerimagelisting.DockerHubClient{
		Username: auth.GHCRUsername,
		Password: auth.GHCRPassword,
		BaseURL:  containerimagelisting.GHCRBaseUrl,
	}

	tags, err := client.ListTags("homebrew/core/docker")
	assert.NoError(t, err)

	t.Logf("Tags: %+v", tags)

	assert.True(t, containsTag("20.10.7", tags))
}
