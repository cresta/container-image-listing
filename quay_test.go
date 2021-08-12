package containerimagelisting

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuay_ListTags(t *testing.T) {
	q := Quay{
		Token: "test_token",
		Client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
				require.Equal(t, "0", r.URL.Query().Get("page"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
"tags": [{"name": "test_name"}]
}`)),
				}, nil
			}),
		},
	}
	ctx := context.Background()
	tags, err := q.ListTags(ctx, "testing")
	require.NoError(t, err)
	require.Equal(t, []Tag{&QuayTag{Name: "test_name"}}, tags)
}
