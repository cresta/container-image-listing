package containerimagelisting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DockerV2 struct {
	BaseURL string
	Client            *http.Client
	ReAuth            *ScopeReauther
	RequestWrapper    RequestWrapper
	MaxReAuthAttempts int
}

func (c *DockerV2) maxReAuthAttempts() int {
	if c.MaxReAuthAttempts == 0 {
		return 1
	}
	return c.MaxReAuthAttempts
}

var _ Registry = &DockerV2{}

// ListTags - Return tags for name in no particular order.
// IE, name="library/redis"
func (c *DockerV2) ListTags(ctx context.Context, repository string) ([]Tag, error) {
	return c.listTagsWithAuthWrapper(ctx, repository, nil, 1)
}

func (c *DockerV2) listTagsWithAuthWrapper(ctx context.Context, repository string, authWrapper RequestWrapper, attemptNumber int) ([]Tag, error) {
	// Documented at https://docs.docker.com/registry/spec/api/#listing-image-tags

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v2/%s/tags/list", c.BaseURL, repository), nil)
	if err != nil {
		return nil, fmt.Errorf("uanble to build http request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	if authWrapper != nil {
		if err := authWrapper.Wrap(req); err != nil {
			return nil, fmt.Errorf("unable to wrap auth with request wrapper: %w", err)
		}
	}
	if c.RequestWrapper != nil {
		if err := c.RequestWrapper.Wrap(req); err != nil {
			return nil, fmt.Errorf("unable to wrap auth with default wrapper: %w", err)
		}
	}

	// Perform request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("uanble to issue HTTP request to list tags: %w", err)
	}

	var body bytes.Buffer
	if _, err := io.Copy(&body, resp.Body); err != nil {
		return nil, fmt.Errorf("unable to copy from response body: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("unable to close response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to reauth if we have one
		if c.ReAuth != nil {
			if attemptNumber > c.maxReAuthAttempts() {
				return nil, fmt.Errorf("past maximum reauth attempts of %d", c.maxReAuthAttempts())
			}
			reauthFunc, err := c.ReAuth.CheckForReauth(ctx, resp, c.Client)
			if err != nil {
				return nil, fmt.Errorf("unable to check for reauth: %w", err)
			}
			if reauthFunc != nil {
				// TODO: Cache this function for this repository
				return c.listTagsWithAuthWrapper(ctx, repository, reauthFunc, attemptNumber+1)
			}
		}
		return nil, fmt.Errorf("invalid status code %d with response %s", resp.StatusCode, resp.Status)
	}

	// Defined at https://docs.docker.com/registry/spec/api/#listing-image-tags
	type tagListResp struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	var tlr tagListResp

	if err := json.NewDecoder(&body).Decode(&tlr); err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}
	var ret []Tag
	for _, t := range tlr.Tags {
		ret = append(ret, &staticTag{tag: t})
	}

	return ret, nil
}
