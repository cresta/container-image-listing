package containerimagelisting

import (
	"os"
	"strings"
)

type Auth struct {
	QuayBearerToken string
	DockerHubToken  string // TODO is a placeholder
	ECRToken        string // TODO fix this to make sense, this is a placeholder
	GHCRToken       string // TODO confirm this makes sense, is a placeholder
}

type ContainerClient interface {
	ListTags(name string) ([]Tag, error)
}

func (a *Auth) NewClient(url string) ContainerClient {
	return NewClient(url, a)
}

func (a *Auth) FromEnv() {
	// TODO finish this once everything is coded
	if value, exists := os.LookupEnv("QUAY_TOKEN"); exists {
		a.QuayBearerToken = value
	}
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
	case strings.Contains(url, "docker.io"): // TODO confirm this
		containerClient = &DockerHubClient{} // TODO add token in here
		// TODO uncomment GHCR
		//case strings.Contains(url, "github.com"):
		//return &GHCRClient{}, nil
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
