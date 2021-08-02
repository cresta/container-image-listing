package containerimagelisting

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// TODO move this to a common file?
type ContainerClient interface {
	ListTags()
}

type DockerClient struct {
	ContainerClient interface{}
}

// parseBearerResponse - Parses bearer token from auth response
func (d *DockerClient) parseBearerResponse(body io.ReadCloser) (string, error) {
	type authResponse struct {
		Token       string    `json:"token"`
		AccessToken string    `json:"access_token"`
		ExpiresIn   int       `json:"expires_in"`
		IssuedAt    time.Time `json:"issued_at"`
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

func (d *DockerClient) getBearerForRepo(name string) (string, error) {
	req, err := http.NewRequest("GET", "https://auth.docker.io/token", nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("service", "registry.docker.io")
	q.Add("scope", fmt.Sprintf("repository:%s:pull", name))
	req.URL.RawQuery = q.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	token, err := d.parseBearerResponse(resp.Body)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ListTags - Return tags for name in no particular order
// IE, name="library/redis"
func (d *DockerClient) ListTags(name string) ([]string, error) {
	// Get auth token
	token, err := d.getBearerForRepo(name)
	if err != nil {
		return nil, err
	}

	// Create URL
	url := fmt.Sprintf("https://registry-1.docker.io/v2/%s/tags/list", name)

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

	err = json.Unmarshal(bodyBytes, &tlr)
	if err != nil {
		return nil, err
	}


	return tlr.Tags, nil
}
