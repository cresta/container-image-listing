package containerimagelisting_test

import (
	"testing"

	"github.com/cresta/container-image-listing"
	"github.com/stretchr/testify/assert"
)

func TestGHCR_ListTagsWithAuth(t *testing.T) {
	t.Parallel()

	auth := containerimagelisting.Auth{}
	auth.FromEnv()

	client := &containerimagelisting.DockerRegistryClient{
		Username: auth.GHCRUsername,
		Password: auth.GHCRPassword,
		BaseURL:  containerimagelisting.GHCRBaseUrl,
	}

	tags, err := client.ListTags("homebrew/core/docker")
	assert.NoError(t, err)

	t.Logf("Tags: %+v", tags)

	assert.True(t, containsTag("20.10.7", tags))
}
