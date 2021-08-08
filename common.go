package containerimagelisting

import (
	"os"
	"strings"
)

type Auth struct {
	QuayBearerToken   string
	DockerHubUsername string
	DockerHubPassword string
	ECRToken          string // TODO fix this to make sense, this is a placeholder
	GHCRUsername      string
	GHCRPassword      string
}

type ContainerClient interface {
	ListTags(name string) ([]Tag, error)
}

func (a *Auth) NewClient(url string) ContainerClient {
	return NewClient(url, a)
}

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
	// TODO finish this once everything is coded
}

// NewClientFromEnv - Creates a new ContainerClient checking
// ENV variables for authorization
// QUAY_TOKEN
// TODO add the other env variables once those are coded
func NewClientFromEnv(url string) ContainerClient {
	auth := &Auth{}
	auth.FromEnv()

	return NewClient(url, auth)
}

func NewClient(url string, auth *Auth) ContainerClient {
	var containerClient ContainerClient
	switch {
	case strings.Contains(url, "quay.io"):
		containerClient = &QuayClient{Token: auth.QuayBearerToken}
	default:
		containerClient = &DockerRegistryClient{
			Username: auth.DockerHubUsername,
			Password: auth.DockerHubPassword,
			BaseURL:  DockerHubBaseUrl,
		}
		//case strings.Contains(url, "amazon.com"):
		//	return nil, nil // TODO fill this out
	}

	return containerClient
}

// stringNamesToTags - Converts a slice of strings to a slice of Tags
func stringNamesToTags(names []string) []Tag {
	var tags []Tag
	for _, name := range names {
		tags = append(tags, Tag{Name: name})
	}

	return tags
}
