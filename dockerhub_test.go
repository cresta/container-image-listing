package containerimagelisting_test

import (
	"testing"

	"github.com/cresta/container-image-listing"
	"github.com/stretchr/testify/assert"
)

func TestDockerClient_ListTags(t *testing.T) {
	t.Parallel()

	dockerClient := containerimagelisting.DockerClient{}

	tags, err := dockerClient.ListTags("library/redis")
	assert.NoError(t, err)

	t.Logf("Tags: %s", tags)

	assert.Contains(t, tags, "latest", "Did not find tag 'latest'")
}
