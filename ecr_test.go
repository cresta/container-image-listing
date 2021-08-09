package containerimagelisting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestECR_ListTags(t *testing.T) {
	t.Parallel()
	auth := &Auth{}
	auth.FromEnv()

	imageURL := "242659714806.dkr.ecr.us-west-2.amazonaws.com/cresta/auth-service"
	client, err := auth.NewECRClient(imageURL)
	assert.NoError(t, err)

	tags, err := client.ListTags("cresta/auth-service")
	t.Logf("Tags: %+v", tags)
	assert.Greater(t, len(tags), 1)

	assert.True(t, containsTag("main-gh.86-d851aeb", tags))

}
