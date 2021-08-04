package containerimagelisting_test

import (
	"testing"

	containerimagelisting "github.com/cresta/container-image-listing"
	"github.com/stretchr/testify/assert"
)

func TestQuayClient_ListTags(t *testing.T) {
	t.Parallel()

	quayClient := containerimagelisting.QuayClient{}
	tags, err := quayClient.ListTags("bedrock/ubuntu")
	assert.NoError(t, err)
	assert.True(t, containsTag("saucy", tags))
}
