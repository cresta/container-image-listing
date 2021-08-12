package containerimagelisting

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Quay implements quay's API in order to fetch docker image tags
type Quay struct {
	Token       string
	BaseURL     string
	MaxPageSize int
	Client      *http.Client
}

func (q *Quay) baseURL() string {
	if q.BaseURL != "" {
		return q.BaseURL
	}
	return "https://quay.io"
}

func (q *Quay) maxPageSize() int {
	if q.MaxPageSize != 0 {
		return q.MaxPageSize
	}
	return 100
}

var _ Registry = &Quay{}

// QuayTag implements the Tag type and also returns extra information quay tags know
type QuayTag struct {
	Name           string `json:"name"`
	Reversion      bool   `json:"reversion"`
	StartTs        int    `json:"start_ts"`
	ImageID        string `json:"image_id"`
	LastModified   string `json:"last_modified"`
	ManifestDigest string `json:"manifest_digest"`
	DockerImageID  string `json:"docker_image_id"`
	IsManifestList bool   `json:"is_manifest_list"`
	Size           int    `json:"size"`
}

func (q *QuayTag) Tag() string {
	return q.Name
}

var _ Tag = &QuayTag{}

func (q *Quay) parseListTagResult(body io.Reader) (tags []QuayTag, additionalPages bool, err error) {
	// Documented on https://access.redhat.com/documentation/en-us/red_hat_quay/3/html-single/red_hat_quay_api_guide/index#get_api_v1_repository_repository_tag
	var ltr struct {
		HasAdditional bool      `json:"has_additional"`
		Page          int       `json:"page"`
		Tags          []QuayTag `json:"tags"`
	}
	if err := json.NewDecoder(body).Decode(&ltr); err != nil {
		return nil, false, fmt.Errorf("unable to read from HTTP body: %w", err)
	}
	return ltr.Tags, ltr.HasAdditional, nil
}

// ListTags returns all quay image tags for a repository
func (q *Quay) ListTags(ctx context.Context, repository string) ([]Tag, error) {
	var ret []Tag
	hasMorePages := true
	for page := 0; hasMorePages; page += 1 {
		// Create URL
		url := fmt.Sprintf("%s/api/v1/repository/%s/tag/", q.baseURL(), repository) // NOTE: Fails without trailing slash

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to create HTTP request URL: %w", err)
		}

		// Added header if it exists
		if q.Token != "" {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", q.Token))
		}

		// Add parameters
		query := req.URL.Query()
		query.Add("page", fmt.Sprintf("%d", page))
		query.Add("onlyActiveTags", "true")
		query.Add("limit", fmt.Sprintf("%d", q.maxPageSize()))
		req.URL.RawQuery = query.Encode()

		// Perform request
		resp, err := q.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("unable to execute HTTP request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			// TODO: Special error code for "repo does not exist"
			return nil, fmt.Errorf("status code was %d instead of 200 with status %s", resp.StatusCode, resp.Status)
		}

		tags, parsedAdditional, err := q.parseListTagResult(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to parse tag results from body: %w", err)
		}
		hasMorePages = parsedAdditional // Note: be careful with shadowing if you move this into the := parseListTagResult line above
		for _, t := range tags {
			t := t
			ret = append(ret, &t)
		}

		// Closing request before we call next recursive function
		err = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("unable to close response body: %w", err)
		}
	}

	return ret, nil
}
