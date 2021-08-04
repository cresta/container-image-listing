package containerimagelisting

// stringNamesToTags - Converts a slice of strings to a slice of Tags
func stringNamesToTags(names []string) []Tag {
	var tags []Tag
	for _, name := range names {
		tags = append(tags, Tag{Name: name})
	}
	return tags
}