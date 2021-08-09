package containerimagelisting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuayClient_ListTags(t *testing.T) {
	t.Parallel()

	quayClient := QuayClient{}
	tags, err := quayClient.ListTags("bedrock/ubuntu")
	assert.NoError(t, err)
	assert.True(t, containsTag("saucy", tags))
}

func TestQuayClient_ListTagsWithAuth(t *testing.T) {
	t.Parallel()

	auth := &Auth{}
	auth.FromEnv()
	if !assert.NotEmpty(t, auth.QuayBearerToken, "Make sure QUAY_TOKEN env variable is set for testing") {
		t.FailNow()
	}
	quayClient := QuayClient{Token: auth.QuayBearerToken}
	tags, err := quayClient.ListTags("cresta/chatmon")
	assert.NoError(t, err)
	t.Logf("Tags: %+v", tags)
	assert.True(t, containsTag("190522-204917-master", tags))
}
