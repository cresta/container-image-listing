package containerimagelisting

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/stretchr/testify/require"
)

func runForTests(t *testing.T, registry Registry, tests []TagTest) {
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.Repository, func(t *testing.T) {
			tags, err := registry.ListTags(ctx, test.Repository)
			require.NoError(t, err)
			for _, expectedTag := range test.ExpectedTags {
				seen := false
				for _, seenTag := range tags {
					if seenTag.Tag() == expectedTag {
						seen = true
						break
					}
				}
				if !seen {
					require.Fail(t, fmt.Sprintf("set for %s does not contain %#v", test.Repository, expectedTag))
				}
			}
		})
	}
}

func TestDefaultRegistryGHCR(t *testing.T) {
	cfg, err := LoadIntegrationTestConfig()
	require.NoError(t, err)
	x := DockerV2{
		BaseURL: "https://ghcr.io",
		Client:  http.DefaultClient,
		ReAuth: &ScopeReauther{
			Username: cfg.GhcrUsername,
			Password: cfg.GhcrPassword,
		},
	}
	runForTests(t, &x, cfg.GhcrTests)
}

func TestDefaultRegistryDockerHub(t *testing.T) {
	cfg, err := LoadIntegrationTestConfig()
	require.NoError(t, err)
	if cfg.DockerhubUsername == "-" {
		cfg.DockerhubUsername = ""
		cfg.DockerhubPassword = ""
	}
	x := DockerV2{
		BaseURL: "https://registry-1.docker.io/",
		Client:  http.DefaultClient,
		ReAuth: &ScopeReauther{
			Username: cfg.DockerhubUsername,
			Password: cfg.DockerhubPassword,
		},
	}
	runForTests(t, &x, cfg.DockerhubTests)
}

func TestDefaultRegistryECR(t *testing.T) {
	cfg, err := LoadIntegrationTestConfig()
	require.NoError(t, err)
	ses, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		t.Skipf("Skipping ECR test.  Unable to make AWS session: %v", err)
	}
	stsClient := sts.New(ses)
	t.Log(stsClient.GetCallerIdentity(nil))
	x := DockerV2{
		BaseURL: cfg.ECRBaseURL,
		Client:  http.DefaultClient,
		RequestWrapper: &ECRAuthWrapper{
			ECR: ecr.New(ses),
		},
		ReAuth: &ScopeReauther{},
	}
	runForTests(t, &x, cfg.ECRTests)
}

func TestFullRegistryFinder(t *testing.T) {
	cfg, err := LoadIntegrationTestConfig()
	require.NoError(t, err)
	ses, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		t.Skipf("Skipping TestFullRegistryFinder test.  Unable to make AWS session: %v", err)
	}
	opts := RegistryFinderOptionalConfig{}
	finder := RegistryFinder{
		Registries: []RegistryWithFinder{
			ForGHCR(cfg.GhcrUsername, cfg.GhcrPassword, opts),
			ForDockerhub(cfg.DockerhubUsername, cfg.DockerhubPassword, opts),
			ForQuay(cfg.QuayToken, opts),
			ForECR(ecr.New(ses), cfg.ECRBaseURL, opts),
		},
	}
	runForTests(t, &finder, cfg.DockerhubTests)
	runForTests(t, &finder, addURLToTests("quay.io/", cfg.QuayTests))
	runForTests(t, &finder, addURLToTests("ghcr.io/", cfg.GhcrTests))
	runForTests(t, &finder, addURLToTests(cfg.ECRBaseURL+"/", cfg.ECRTests))
}

func addURLToTests(urlToAdd string, tests []TagTest) []TagTest {
	ret := make([]TagTest, 0, len(tests))
	for _, t := range tests {
		ret = append(ret, TagTest{
			Repository:   strings.TrimPrefix(urlToAdd+t.Repository, "https://"),
			ExpectedTags: t.ExpectedTags,
		})
	}
	return ret
}
