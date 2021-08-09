package containerimagelisting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGHCR_ListTagsWithAuth(t *testing.T) {
	t.Parallel()

	auth := Auth{}
	auth.FromEnv()

	client := &DockerRegistryClient{
		Username: auth.GHCRUsername,
		Password: auth.GHCRPassword,
		BaseURL:  GHCRBaseURL,
	}

	tags, err := client.ListTags("homebrew/core/docker")
	assert.NoError(t, err)

	t.Logf("Tags: %+v", tags)

	assert.True(t, containsTag("20.10.7", tags))
}
