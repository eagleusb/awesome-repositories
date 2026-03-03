package github

import (
	"fmt"
	"os"

	"github.com/google/go-github/v83/github"
)

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

type GitHubClient struct {
	client *github.Client
	config *githubConfig
	StarredRepos []*github.Repository
	Repos *githubRepos
}

type githubConfig struct {
	username     string
	page         int
	itemsPerPage int
	maxItems     int
}

func NewGitHubClient() (*GitHubClient, error) {
	if GITHUB_TOKEN == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	return &GitHubClient{
		client: github.NewClient(nil).WithAuthToken(GITHUB_TOKEN),
		config: &githubConfig{},
		Repos: &githubRepos{
			ByLanguage: make(map[string][]*githubRepo),
			ByCategory: make(map[string][]*githubRepo),
		},
	}, nil
}

func (c *GitHubClient) setConfig(username string, itemsPerPage, maxItems int) {
	c.config = &githubConfig{
		username:     username,
		itemsPerPage: itemsPerPage,
		maxItems:     maxItems,
	}

	if c.config.itemsPerPage <= 0 || c.config.itemsPerPage > 100 {
		c.config.itemsPerPage = 100
	}
	if c.config.maxItems <= 0 {
		c.config.maxItems = 999999
	}
}
