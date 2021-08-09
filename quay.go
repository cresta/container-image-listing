package containerimagelisting

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

type QuayClient struct {
	Token string
}

var _ ContainerClient = &QuayClient{}

const QuayBaseURL = "quay.io"
const QuayMaxPageSize = 100

type Tag struct {
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

func (c *QuayClient) parseListTagResult(body io.ReadCloser) (tags []Tag, additionalPages bool, err error) {
	type listTagsResult struct {
		HasAdditional bool  `json:"has_additional"`
		Page          int   `json:"page"`
		Tags          []Tag `json:"tags"`
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, false, err
	}

	ltr := listTagsResult{}
	err = json.Unmarshal(bodyBytes, &ltr)
	if err != nil {
		return nil, false, err
	}

	return ltr.Tags, ltr.HasAdditional, nil
}

// listTagsPaging - Recursive helper function to confirm we've gotten all pages for the tags from Quay
// Should typically start with page 0
func (c *QuayClient) listTagsPaging(name string, page int) ([]Tag, error) {
	// Create URL
	url := fmt.Sprintf("https://%s/api/v1/repository/%s/tag/", QuayBaseURL, name) // NOTE: Fails without trailing slash

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Added header if it exists
	if c.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	// Add parameters
	q := req.URL.Query()
	q.Add("page", fmt.Sprintf("%d", page))
	q.Add("onlyActiveTags", "true")
	q.Add("limit", fmt.Sprintf("%d", QuayMaxPageSize))
	req.URL.RawQuery = q.Encode()

	// Perform request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("status code was %d instead of 200 with status %s", resp.StatusCode, resp.Status)
		log.Print(msg)
		return nil, errors.New(msg)
	}

	// Parse
	tags, hasAdditional, err := c.parseListTagResult(resp.Body)
	if err != nil {
		return nil, err
	}

	// Closing request before we call next recursive function
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Fetch and join tags if needed
	if hasAdditional {
		recursiveTags, err := c.listTagsPaging(name, page+1)
		if err != nil {
			return nil, err
		}

		tags = append(tags, recursiveTags...)
	}

	return tags, nil
}

// ListTags - Return tags for name in no particular order
// IE, name="bedrock/ubuntu"
// For Quay some repositories that are public return a 401 unauthorized.
// For example bedrock/ubuntu works fine however
// prometheus/node-exporter gives a 401 unauthorized
func (c *QuayClient) ListTags(name string) ([]Tag, error) {
	tags, err := c.listTagsPaging(name, 0)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
