package containerimagelisting

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiURLHostMatcher_RepositoryForURL(t *testing.T) {
	testFunc := func(given MultiURLHostMatcher, repo string, expected string) func(t *testing.T) {
		return func(t *testing.T) {
			require.Equal(t, expected, given.RepositoryForURL(repo))
		}
	}
	t.Run("empty", testFunc(MultiURLHostMatcher{}, "test", ""))
	t.Run("full_match", testFunc(MultiURLHostMatcher{ValidDomains: []string{"hello"}}, "hello/world", "world"))
	t.Run("ecr_match", testFunc(MultiURLHostMatcher{ValidDomainSubstrings: []string{".amazonaws.com"}}, "123123123.dkr.ecr.us-west-2.amazonaws.com/cresta/example-service", "cresta/example-service"))
	t.Run("ecr_match_regex", testFunc(MultiURLHostMatcher{ValidRegex: []*regexp.Regexp{regexp.MustCompile(`dkr\.ecr\..*\.amazonaws\.com`)}}, "123123123.dkr.ecr.us-west-2.amazonaws.com/cresta/example-service", "cresta/example-service"))
}

func TestDockerHubLocator(t *testing.T) {
	testFunc := func(given DockerHubLocator, repo string, expected string) func(t *testing.T) {
		return func(t *testing.T) {
			require.Equal(t, expected, given.RepositoryForURL(repo))
		}
	}
	t.Run("empty", testFunc(DockerHubLocator{}, "test", "test"))
	t.Run("simple_match", testFunc(DockerHubLocator{}, "ubuntu", "ubuntu"))
	t.Run("non_simple_match", testFunc(DockerHubLocator{}, "ghcr.io/bob", ""))
}
