package containerimagelisting

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseWwwAuthenticate(t *testing.T) {
	authChecker := func(given string, expectedBearer string, expectedValues map[string]string) func(t *testing.T) {
		return func(t *testing.T) {
			x := parseWwwAuthenticate(given)
			require.Equal(t, expectedBearer, x.Type)
			require.Equal(t, expectedValues, x.Values)
		}
	}

	t.Run("doc_example", authChecker(
		`Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:samalba/my-app:pull,push"`,
		"Bearer",
		map[string]string{
			"realm":   "https://auth.docker.io/token",
			"service": "registry.docker.io",
			"scope":   "repository:samalba/my-app:pull,push",
		},
	))
	t.Run("strange_case", authChecker(
		`Auth realm="https://auth.docker.io/token"`,
		"Auth",
		map[string]string{
			"realm": "https://auth.docker.io/token",
		},
	))
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func TestScopeReauther(t *testing.T) {
	x := ScopeReauther{
		Username: "john",
		Password: "doe",
	}
	dummyResp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Header:     make(http.Header),
	}
	dummyResp.Header.Set("Www-Authenticate", `Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:samalba/my-app:pull,push"`)
	ctx := context.Background()
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			u, p, ok := r.BasicAuth()
			require.True(t, ok)
			require.Equal(t, x.Username, u)
			require.Equal(t, x.Password, p)
			require.Equal(t, "repository:samalba/my-app:pull,push", r.URL.Query().Get("scope"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"token":"abc"}`)),
			}, nil
		}),
	}
	recvFunc, err := x.CheckForReauth(ctx, dummyResp, client)
	require.NoError(t, err)
	require.NotNil(t, recvFunc)
	testReq, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	require.NoError(t, recvFunc.Wrap(testReq))
	require.Equal(t, "Bearer abc", testReq.Header.Get("Authorization"))
}
