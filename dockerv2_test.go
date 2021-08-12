package containerimagelisting

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerV2_ListTags(t *testing.T) {
	d := DockerV2{
		BaseURL: "http://example.com",
		Client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
"name": "test_name",
"tags": ["test_name"]
}`)),
				}, nil
			}),
		},
	}
	ctx := context.Background()
	tags, err := d.ListTags(ctx, "test_repo")
	require.NoError(t, err)
	require.Equal(t, []Tag{&staticTag{tag: "test_name"}}, tags)
}
