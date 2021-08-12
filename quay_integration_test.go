package containerimagelisting

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuayIntegration(t *testing.T) {
	q := Quay{
		Client: http.DefaultClient,
	}
	ctx := context.Background()
	tags, err := q.ListTags(ctx, "bitnami/sealed-secrets-controller")
	require.NoError(t, err)
	require.NotNil(t, tags)
}

func TestQuayPrivateRepos(t *testing.T) {
	cfg, err := LoadIntegrationTestConfig()
	require.NoError(t, err)
	if cfg.QuayToken == "" {
		t.Skipf("Cannot run Quay integration test due to missing env QUAY_TOKEN")
	}

	x := Quay{
		Token:  cfg.QuayToken,
		Client: http.DefaultClient,
	}
	runForTests(t, &x, cfg.QuayTests)
}
