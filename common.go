package containerimagelisting

import (
	"context"
)

// Tag is the tag of a docker image.  Some repositories, like quay for example, may extend this interface with extra
// information
type Tag interface {
	// Tag returns the tag. For example, for container "ubuntu:latest" it would return "latest"
	Tag() string
}

type staticTag struct {
	tag string
}

func (s *staticTag) Tag() string {
	return s.tag
}

var _ Tag = &staticTag{}

// Registry is anything that stores docker images and can list images for a given repository
type Registry interface {
	// ListTags should return all tags for a repository inside this registry.  If unable to return all tags, it should
	// prioritize the tags most recently created.
	ListTags(ctx context.Context, repository string) ([]Tag, error)
}
