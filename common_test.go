package containerimagelisting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func containsTag(name string, tags []Tag) bool {
	for _, tag := range tags {
		if tag.Name == name {
			return true
		}
	}
	return false
}

func TestNewContainerClient(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantedClient ContainerClient
	}{
		// Test Cases
		{
			name:         "quay",
			url:          "https://quay.io/cresta/chatmon",
			wantedClient: &QuayClient{},
		},
		{
			name:         "dockerhub",
			url:          "https://docker.io",
			wantedClient: &DockerRegistryClient{},
		},
	}

	for _, tt := range tests {
		client, err := NewClientFromEnv(tt.url)
		assert.NoError(t, err)
		assert.IsType(t, tt.wantedClient, client)
	}
}

func Test_ListTags(t *testing.T) {
	t.Parallel()

	urls := []string{
		"242659714806.dkr.ecr.us-west-2.amazonaws.com/cresta/auth-service",
		"ghcr.io/homebrew/core/docker",
		"quay.io/cresta/chatmon",
		"https://hub.docker.com/repository/docker/crestaai/build-cache",
	}
	for _, url := range urls {
		tags, err := ListTags(url)
		assert.NoError(t, err, "encountered an error while requesting %s", url)
		assert.Greater(t, len(tags), 1)
	}
}

func Test_urlToName(t *testing.T) {
	type args struct {
		imageURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "aws",
			args: args{imageURL: "242659714806.dkr.ecr.us-west-2.amazonaws.com/cresta/auth-service"},
			want: "cresta/auth-service",
		},
		{
			name: "ghcr",
			args: args{imageURL: "ghcr.io/homebrew/core/docker"},
			want: "homebrew/core/docker",
		},
		{
			name: "quay",
			args: args{imageURL: "quay.io/cresta/chatmon"},
			want: "cresta/chatmon",
		},
		{
			name: "dockerhub",
			args: args{imageURL: "https://hub.docker.com/repository/docker/crestaai/build-cache"},
			want: "crestaai/build-cache",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlToName(tt.args.imageURL); got != tt.want {
				t.Errorf("urlToName() = %v, want %v", got, tt.want)
			}
		})
	}
}
