package containerimagelisting

import (
	"context"
	"net/http"
	"regexp"
)

type RegistryWithFinder struct {
	Registry          Registry
	RepositoryLocator RepositoryLocator
}

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

type RegistryFinderOptionalConfig struct {
	Client *http.Client
}

func (r *RegistryFinderOptionalConfig) getClient() *http.Client {
	if r.Client == nil {
		return http.DefaultClient
	}
	return r.Client
}

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
