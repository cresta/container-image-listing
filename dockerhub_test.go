package containerimagelisting_test

import (
	"testing"

	"github.com/cresta/container-image-listing"
	"github.com/stretchr/testify/assert"
)

func TestDockerClient_ListTags(t *testing.T) {
	t.Parallel()

	dockerClient := containerimagelisting.DockerHubClient{}

	tags, err := dockerClient.ListTags("library/redis")
	assert.NoError(t, err)

	t.Logf("Tags: %+v", tags)

	assert.True(t, containsTag("latest", tags))
}
