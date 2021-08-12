package containerimagelisting

import (
	"regexp"
	"strings"
)

type RepositoryLocator interface {
	RepositoryForURL(url string) string
}

type URLMatchFunc func(url string) string

func (U URLMatchFunc) RepositoryForURL(url string) string {
	return U(url)
}

type MultiURLHostMatcher struct {
	ValidDomains          []string
	ValidDomainSubstrings []string
	ValidRegex            []*regexp.Regexp
	ReturnFullRepo        bool
}

func (m *MultiURLHostMatcher) matches(firstDomain string) bool {
	for _, s := range m.ValidDomains {
		if s == firstDomain {
			return true
		}
	}
	for _, s := range m.ValidDomainSubstrings {
		if strings.Contains(firstDomain, s) {
			return true
		}
	}
	for _, r := range m.ValidRegex {
		if r.MatchString(firstDomain) {
			return true
		}
	}
	return false
}

func (m *MultiURLHostMatcher) RepositoryForURL(repo string) string {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) == 1 {
		return ""
	}
	if m.matches(parts[0]) {
		if m.ReturnFullRepo {
			return repo
		} else {
			return parts[1]
		}
	}
	return ""
}

var _ RepositoryLocator = URLMatchFunc(nil)

type DockerHubLocator struct {
	MultiURLHostMatcher MultiURLHostMatcher
}

func (m *DockerHubLocator) RepositoryForURL(repo string) string {
	// docker pull cresta/blarg    <--- dockerhub
	// docker pull ghcr.io/a/b     <--- no docker hub

	// If you get ghcr.io/a/b  ->>>> You want to use the repo "a/b" not the repo "ghcr.io/a/b"

	parts := strings.Split(repo, "/")
	if !strings.Contains(parts[0], ".") {
		return repo
	}
	return m.MultiURLHostMatcher.RepositoryForURL(repo)
}
