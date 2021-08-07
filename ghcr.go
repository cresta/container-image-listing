package containerimagelisting

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

const GHCRBaseUrl = "ghcr.io"

type GHCRClient struct {
	ContainerClient interface{}
}

// parseBearerResponse - Parses bearer token from auth response
func (c *GHCRClient) parseBearerResponse(body io.ReadCloser) (string, error) {
	type authResponse struct {
		Token string `json:"token"`
	}

	ar := authResponse{}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(bodyBytes, &ar)
	if err != nil {
		return "", err
	}

	return ar.Token, nil
}

func (c *GHCRClient) getBearerForRepo(name string) (string, error) {
	url := fmt.Sprintf("https://%s/token", GHCRBaseUrl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("scope", fmt.Sprintf("repository:%s:pull", name))
	req.URL.RawQuery = q.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	token, err := c.parseBearerResponse(resp.Body)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ListTags - Return tags for name in no particular order
// IE, name="library/redis"
func (c *GHCRClient) ListTags(name string) ([]string, error) {
	// Get auth token
	token, err := c.getBearerForRepo(name)
	if err != nil {
		return nil, err
	}

	// Create URL
	url := fmt.Sprintf("https://%s/v2/%s/tags/list", GHCRBaseUrl, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add auth
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Perform request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("status code was %d instead of 200 with status %s", resp.StatusCode, resp.Status)
		log.Print(msg)
		return nil, errors.New(msg)
	}

	// Parse tags
	type tagListResp struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	var tlr tagListResp

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TODO remove
	//log.Printf("BodyBytes: %s", bodyBytes)

	err = json.Unmarshal(bodyBytes, &tlr)
	if err != nil {
		return nil, err
	}

	return tlr.Tags, nil
}
