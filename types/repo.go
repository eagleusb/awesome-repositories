package types

// Repo represents a GitHub repository with relevant metadata
type Repo struct {
	Name        string
	Description string
	Language    string
	Category    []string
	URL         string
	Stars       int
}
