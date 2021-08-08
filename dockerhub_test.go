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

	// Checking if we can handle a lot of tags without paging
	// Docker Registry doesn't seem to require paging or we haven't found a repository
	// with enough tags yet. It also seems that docker registry doesn't honor the paging
	// parameters set forth in https://docs.docker.com/registry/spec/api/#listing-image-tags
	assert.Greater(t, len(tags), 500)

	assert.True(t, containsTag("latest", tags))
}

func TestDockerClient_ListTagsWithAuth(t *testing.T) {
	t.Parallel()

	auth := &containerimagelisting.Auth{}
	auth.FromEnv()

	if !assert.NotEmpty(t, auth.DockerHubUsername, "Make sure DOCKERHUB_USERNAME env variable is set for testing") {
		t.FailNow()
	}
	if !assert.NotEmpty(t, auth.DockerHubPassword, "Make sure DOCKERHUB_PASSWORD env variable is set for testing") {
		t.FailNow()
	}

	dockerClient := containerimagelisting.DockerHubClient{Username: auth.DockerHubUsername,
		Password: auth.DockerHubPassword}

	// Test private repo with access token
	tags, err := dockerClient.ListTags("crestaai/build-cache")
	assert.NoError(t, err)

	t.Logf("Tags: %+v", tags)

	// Picked this tag because it is funny ;)
	assert.True(t, containsTag("jacktest-f", tags))
}
