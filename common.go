package containerimagelisting

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Auth struct {
	QuayBearerToken   string
	DockerHubUsername string
	DockerHubPassword string
	GHCRUsername      string
	GHCRPassword      string
}

type ContainerClient interface {
	ListTags(name string) ([]Tag, error)
}

func (a *Auth) NewQuayClient() ContainerClient {
	return &QuayClient{
		Token: a.QuayBearerToken,
	}
}

func (a *Auth) NewDockerHubClient() ContainerClient {
	return &DockerRegistryClient{
		Username: a.DockerHubUsername,
		Password: a.DockerHubPassword,
		BaseURL:  DockerHubBaseURL,
	}
}

func (a *Auth) NewGHCRClient() ContainerClient {
	return &DockerRegistryClient{
		Username: a.GHCRUsername,
		Password: a.GHCRPassword,
		BaseURL:  GHCRBaseURL,
	}
}

func (a *Auth) NewECRClient(imageURL string) (ContainerClient, error) {
	if !strings.Contains(imageURL, "http") {
		imageURL = "https://" + imageURL
	}
	u, err := url.Parse(imageURL)
	if err != nil {
		return nil, err
	}

	return &DockerRegistryClient{
		BaseURL: u.Hostname(),
	}, nil
}

func (a *Auth) NewClient(url string) (ContainerClient, error) {
	return NewClient(url, a)
}

// FromEnv - Populates Auth with values from the environment.
func (a *Auth) FromEnv() {
	if value, exists := os.LookupEnv("QUAY_TOKEN"); exists {
		a.QuayBearerToken = value
	}
	if value, exists := os.LookupEnv("DOCKERHUB_PASSWORD"); exists {
		a.DockerHubPassword = value
	}
	if value, exists := os.LookupEnv("DOCKERHUB_USERNAME"); exists {
		a.DockerHubUsername = value
	}
	if value, exists := os.LookupEnv("GHCR_USERNAME"); exists {
		a.GHCRUsername = value
	}
	if value, exists := os.LookupEnv("GHCR_PASSWORD"); exists {
		a.GHCRPassword = value
	}
}

// NewClientFromEnv - Creates a new ContainerClient checking
// ENV variables for authorization.
// See Auth.FromEnv() for a complete list.
func NewClientFromEnv(url string) (ContainerClient, error) {
	auth := &Auth{}
	auth.FromEnv()

	return NewClient(url, auth)
}

func NewClient(url string, auth *Auth) (ContainerClient, error) {
	switch {
	case strings.Contains(url, QuayBaseURL):
		return auth.NewQuayClient(), nil
	case strings.Contains(url, ECRBaseURL):
		containerClient, err := auth.NewECRClient(url)
		if err != nil {
			return nil, err
		}
		return containerClient, nil
	case strings.Contains(url, "docker"): // Need to catch hub.docker.com and docker.io
		return auth.NewDockerHubClient(), nil
	case strings.Contains(url, GHCRBaseURL):
		return auth.NewGHCRClient(), nil
	}

	return nil, fmt.Errorf("No clients matched for url %s", url)
}

// stringNamesToTags - Converts a slice of strings to a slice of Tags.
func stringNamesToTags(names []string) []Tag {
	var tags []Tag
	for _, name := range names {
		tags = append(tags, Tag{Name: name})
	}

	return tags
}

func urlToName(imageURL string) string {
	imageURL = strings.TrimPrefix(imageURL, "http://")
	imageURL = strings.TrimPrefix(imageURL, "https://")

	expr := "/.*$"
	r, err := regexp.Compile(expr)
	if err != nil {
		log.Fatalf("Regex %s failed to compile", expr)
	}
	imageURL = r.FindString(imageURL)

	imageURL = strings.TrimPrefix(imageURL, "/")
	imageURL = strings.TrimPrefix(imageURL, "repository/docker/")

	return imageURL

}

// ListTags - Wrapper function to create a new client and list tags in one step.
// TODO make a single function that gets a new client and grabs tags in one step?
func ListTags(imageURL string) ([]Tag, error) {
	client, err := NewClientFromEnv(imageURL)
	if err != nil {
		return nil, err
	}

	name := urlToName(imageURL)

	return client.ListTags(name)
}
