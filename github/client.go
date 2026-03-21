package github

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	ghtransport "github.com/bored-engineer/github-conditional-http-transport"
	bboltstorage "github.com/bored-engineer/github-conditional-http-transport/bbolt"
	"github.com/google/go-github/v83/github"
	"go.etcd.io/bbolt"

	"github.com/eagleusb/awesome-repositories/types"
)

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

type GitHubClient struct {
	client       *github.Client
	config       *githubConfig
	StarredRepos []*github.Repository
	Repos        *Repos
}

type githubConfig struct {
	username     string
	page         int
	itemsPerPage int
	maxItems     int
	workers      int
}

func NewGitHubClient() (*GitHubClient, error) {
	if GITHUB_TOKEN == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}
	cachePath := filepath.Join(cacheDir, "awesome-repositories", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	httpClient := &http.Client{
		Transport: ghtransport.NewTransport(
			bboltstorage.MustOpen(
				cachePath,
				0644,
				&bbolt.Options{Timeout: 15 * time.Second},
				nil),
			http.DefaultTransport,
		),
		Timeout: 15 * time.Second,
	}

	return &GitHubClient{
		client: github.NewClient(httpClient).WithAuthToken(GITHUB_TOKEN),
		config: &githubConfig{},
		Repos: &Repos{
			ByLanguage: make(map[string][]*types.Repo),
			ByCategory: make(map[string][]*types.Repo),
		},
	}, nil
}

func (c *GitHubClient) setConfig(username string, itemsPerPage, maxItems, workers int) {
	c.config = &githubConfig{
		username:     username,
		itemsPerPage: itemsPerPage,
		maxItems:     maxItems,
		workers:      workers,
	}

	if c.config.itemsPerPage <= 0 || c.config.itemsPerPage > 100 {
		c.config.itemsPerPage = 100
	}
	if c.config.maxItems <= 0 {
		c.config.maxItems = 999999
	}
	if c.config.workers <= 0 {
		c.config.workers = 15
	}
}
