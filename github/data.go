package github

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	u "github.com/eagleusb/awesome-repositories/utils"
	"github.com/google/go-github/v83/github"
)

type githubRepo struct {
	Name        string
	Description string
	Language    string
	Category    []string
	URL         string
	Stars       int
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

func (c *GitHubClient) ClassifyRepos() *GitHubClient {
	for _, r := range c.StarredRepos {
		if r == nil {
			fmt.Printf("Skipping nil repository\n")
			continue
		}

		name := r.GetName()
		language := r.GetLanguage()
		description := r.GetDescription()
		url := r.GetHTMLURL()
		stars := r.GetStargazersCount()
		category := r.Topics

		if language == "" {
			language = "Unknown"
		}
		if description == "" {
			description = "Unknown"
		}
		if len(category) == 0 {
			category = []string{"Unknown"}
		}

		repo := &githubRepo{
			Name:        name,
			Language:    language,
			Description: description,
			URL:         url,
			Stars:       stars,
			Category:    category,
		}
		c.Repos.ByLanguage[language] = append(c.Repos.ByLanguage[language], repo)
	}
	return c
}

func (c *GitHubClient) WriteRepos() error {
	dir := "stars/byLanguage"
	if err := u.EnsureDirectory(dir); err != nil {
		return err
	}

	for language, repos := range c.Repos.ByLanguage {
		sanitized := u.SanitizeLanguage(language)
		filename := filepath.Join(dir, sanitized+".md")

		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()

		sortedRepos := make([]*githubRepo, len(repos))
		copy(sortedRepos, repos)
		slices.SortFunc(sortedRepos, func(a, b *githubRepo) int {
			return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		})

		_, err = fmt.Fprintf(writer, "## %s (%d repositories) \n", language, len(sortedRepos))
		if err != nil {
			return fmt.Errorf("failed to write header to %s: %w", filename, err)
		}

		for _, repo := range sortedRepos {
			_, err := fmt.Fprintf(writer, "- [%s](%s) (%d stars) - %s\n", repo.Name, repo.URL, repo.Stars, repo.Description)
			if err != nil {
				return fmt.Errorf("failed to write repo %s to %s: %w", repo.Name, filename, err)
			}
		}

		if err := writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush %s: %w", filename, err)
		}

		fmt.Printf("Wrote %d repositories to %s\n", len(sortedRepos), filename)
	}

	return nil
}
