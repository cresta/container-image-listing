package containerimagelisting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerHub_ListTags_NoAuth(t *testing.T) {
	t.Parallel()

	auth := &Auth{}
	dockerClient := auth.NewDockerHubClient()

	tags, err := dockerClient.ListTags("library/redis")
	assert.NoError(t, err)

	t.Logf("Tags: %+v", tags)

	// Checking if we can handle a lot of tags without paging
	// Docker Registry doesn't seem to require paging or we haven't found a repository
	// with enough tags yet. It also seems that docker registry doesn't honor the paging
	// parameters set forth in https://docs.docker.com/registry/spec/api/#listing-image-tags
	assert.Greater(t, len(tags), 500)

	assert.True(t, containsTag("latest", tags))
}

func TestDockerHub_ListTags_WithAuth(t *testing.T) {
	t.Parallel()

	auth := &Auth{}
	auth.FromEnv()

	if !assert.NotEmpty(t, auth.DockerHubUsername, "Make sure DOCKERHUB_USERNAME env variable is set for testing") {
		t.FailNow()
	}
	if !assert.NotEmpty(t, auth.DockerHubPassword, "Make sure DOCKERHUB_PASSWORD env variable is set for testing") {
		t.FailNow()
	}

	dockerClient := auth.NewDockerHubClient()

	// Test private repo with access token
	tags, err := dockerClient.ListTags("crestaai/build-cache")
	assert.NoError(t, err)

	// Picked this tag because it is funny ;)
	assert.True(t, containsTag("jacktest-f", tags))
}
