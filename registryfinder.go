package containerimagelisting

import (
	"context"
	"net/http"
	"regexp"
)

// RegistryWithFinder is used by RegistryFinder to match docker images with the registry that should fetch it
type RegistryWithFinder struct {
	Registry          Registry
	RepositoryLocator RepositoryLocator
}

// RegistryFinder helps aggregate different registries with a way to match images to the registry
type RegistryFinder struct {
	Registries []RegistryWithFinder
}

var _ Registry = &RegistryFinder{}

// ListTags for a repository using many backends.
// Should take a repository like what we would see on "docker pull X"
func (r *RegistryFinder) ListTags(ctx context.Context, repository string) ([]Tag, error) {
	for _, registry := range r.Registries {
		scrubbedURL := registry.RepositoryLocator.RepositoryForURL(repository)
		if scrubbedURL != "" {
			return registry.Registry.ListTags(ctx, scrubbedURL)
		}
	}
	// TODO: Do we want a "not found" error code of some kind?
	return nil, nil
}

// RegistryFinderOptionalConfig configures the helper functions for registries
type RegistryFinderOptionalConfig struct {
	Client *http.Client
}

func (r *RegistryFinderOptionalConfig) getClient() *http.Client {
	if r.Client == nil {
		return http.DefaultClient
	}
	return r.Client
}

// ForGHCR factory helps create a GHCR registry with its finder
func ForGHCR(ghcrUsername string, ghcrPassword string, cfg RegistryFinderOptionalConfig) RegistryWithFinder {
	return RegistryWithFinder{
		Registry: &DockerV2{
			BaseURL: "https://ghcr.io",
			Client:  cfg.getClient(),
			ReAuth: &ScopeReauther{
				Username: ghcrUsername,
				Password: ghcrPassword,
			},
		},
		RepositoryLocator: &MultiURLHostMatcher{
			ValidDomains: []string{"ghcr.io"},
		},
	}
}

// ForDockerhub factory helps create a docker hub registry with its finder
func ForDockerhub(dockerhubUsername string, dockerhubPassword string, cfg RegistryFinderOptionalConfig) RegistryWithFinder {
	return RegistryWithFinder{
		Registry: &DockerV2{
			BaseURL: "https://registry-1.docker.io/",
			Client:  cfg.getClient(),
			ReAuth: &ScopeReauther{
				Username: dockerhubUsername,
				Password: dockerhubPassword,
			},
		},
		RepositoryLocator: &DockerHubLocator{
			MultiURLHostMatcher: MultiURLHostMatcher{
				ValidDomains: []string{"docker.io"},
			},
		},
	}
}

// ForQuay factory helps create a quay registry with its finder
func ForQuay(quayToken string, cfg RegistryFinderOptionalConfig) RegistryWithFinder {
	return RegistryWithFinder{
		Registry: &Quay{
			Token:  quayToken,
			Client: cfg.getClient(),
		},
		RepositoryLocator: &MultiURLHostMatcher{
			ValidDomains: []string{"quay.io"},
		},
	}
}

// ForECR factory helps create a ECR registry with its finder
func ForECR(ecrClient ECRClient, ecrBaseURL string, cfg RegistryFinderOptionalConfig) RegistryWithFinder {
	return RegistryWithFinder{
		Registry: &DockerV2{
			BaseURL: ecrBaseURL,
			Client:  cfg.getClient(),
			RequestWrapper: &ECRAuthWrapper{
				ECR:            ecrClient,
				AuthBufferTime: 0,
			},
		},
		RepositoryLocator: &MultiURLHostMatcher{
			ValidRegex: []*regexp.Regexp{regexp.MustCompile(`dkr\.ecr\..*\.amazonaws\.com`)},
		},
	}
}
