# container-image-listing
Library for listing difference attributes of images.

## Quickstart
There are several ways to interact with this library. The easiest is to use the `ListTags` function which accepts image urls. If the `ENV` variables from the `Local Testing` section are present this will work out of the box.
```go
imageURLS := []string{
	"242659714806.dkr.ecr.us-west-2.amazonaws.com/cresta/auth-service",
	"ghcr.io/homebrew/core/docker",
	"quay.io/cresta/chatmon",
	"https://hub.docker.com/repository/docker/crestaai/build-cache",
}
for _, imageURL := range imaegURLS {
	tags, err := containerimagelisting.ListTags(imageURL)
	// Handle errors
}
```

Another alternative is create a client for a specific endpoint.

```go
auth = containerimagelisting.Auth{}
auth.FromEnv()
```

Alternatively one can also use several `NewClient` functions.

## Local Testing
Populate credentials into your environment something like this. In my case I've got a file called `env.sh` that I use.
```bash
export QUAY_TOKEN="tPTt1..................................."
export DOCKERHUB_USERNAME="crestarobot"
export DOCKERHUB_PASSWORD="a1a5....-....-....-....-............"  # This is an access token, it should be used with a username like a password  https://hub.docker.com/settings/security
export GHCR_USERNAME="scall....."
export GHCR_PASSWORD="ghp_V8N................................."  # This is a personal access token
eval $(aws-export-credentials  --env-export --profile infra-prod)
go test
```