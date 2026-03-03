package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v83/github"
)

type githubRepo struct {
	Name        string
	Description string
	Language    string
	Category    []string
}

type githubRepos struct {
	ByLanguage map[string][]*githubRepo
	ByCategory map[string][]*githubRepo
}

func (c *GitHubClient) GetStarredRepos(username string, limit, page int) (*GitHubClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	c.setConfig(username, page, limit)

	opts := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: c.config.itemsPerPage},
	}

	iter := c.client.Activity.ListStarredIter(ctx, c.config.username, opts)
	count := 0

	for starredRepo, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("failed to fetch starred repos: %w", err)
		}

		if count >= c.config.maxItems {
			fmt.Printf("Reached maximum limit of %d starred repositories.\n", count)
			break
		}

		c.StarredRepos = append(c.StarredRepos, starredRepo.Repository)

		count++
		if count%c.config.itemsPerPage == 0 {
			fmt.Printf("Processed %d repositories...\n", count)
		}
	}

	fmt.Printf("Total repositories fetched: %d\n", count)
	return c, nil
}

func (c *GitHubClient) ClassifyRepos() {
	for _, r := range c.StarredRepos {
		func() {
			var language string
			var name string
			var description string
			var category []string

			name = r.GetName()

			defer func() {
				if rec := recover(); rec != nil {
					fmt.Printf("Panic recovered for repo '%s': %v\n", name, rec)
				}
			}()

			language = r.GetLanguage()
			description = r.GetDescription()

			if language == "" {
				language = "Unknown"
			}
			if description == "" {
				description = "Unknown"
			}
			if len(r.Topics) == 0 {
				category = []string{"Unknown"}
			} else {
				category = r.Topics
			}

			repo := &githubRepo{
				Name:        name,
				Description: description,
				Language:    language,
				Category:    category,
			}
			c.Repos.ByLanguage[language] = append(c.Repos.ByLanguage[language], repo)
		}()
	}
}
