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
