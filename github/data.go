package github

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	u "github.com/eagleusb/awesome-repositories/utils"
	"github.com/google/go-github/v83/github"
	"golang.org/x/time/rate"
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

func (c *GitHubClient) GetStarredRepos(username string, limit, page, workers int) (*GitHubClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	c.setConfig(username, page, limit, workers)

	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	opts := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: c.config.itemsPerPage},
	}

	firstPage, resp, err := c.client.Activity.ListStarred(ctx, c.config.username, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch first page: %w", err)
	}

	totalPages := resp.LastPage
	fmt.Printf("Total pages to fetch: %d\n", totalPages)

	c.StarredRepos = make([]*github.Repository, 0, c.config.maxItems)
	for _, starredRepo := range firstPage {
		c.StarredRepos = append(c.StarredRepos, starredRepo.Repository)
	}

	if totalPages <= 1 {
		fmt.Printf("Total repositories fetched: %d\n", len(c.StarredRepos))
		return c, nil
	}

	pageChan := make(chan int, 10)
	errorsChan := make(chan error, c.config.workers*totalPages)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < c.config.workers; i++ {
		wg.Add(1)
		go c.fetchPageWorker(ctx, &wg, &mu, limiter, pageChan, errorsChan, username, opts)
	}

	go func() {
		for p := 2; p <= totalPages; p++ {
			pageChan <- p
		}
		close(pageChan)
	}()

	wg.Wait()
	close(errorsChan)

	var errors []error
	for err := range errorsChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to fetch %d pages: %v", len(errors), errors)
	}

	if len(c.StarredRepos) > c.config.maxItems {
		c.StarredRepos = c.StarredRepos[:c.config.maxItems]
	}

	fmt.Printf("Total repositories fetched: %d\n", len(c.StarredRepos))
	return c, nil
}

func (c *GitHubClient) fetchPageWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	mu *sync.Mutex,
	limiter *rate.Limiter,
	pageChan <-chan int,
	errorsChan chan<- error,
	username string,
	opts *github.ActivityListStarredOptions,
) {
	defer wg.Done()

	maxRetries := 3

	for page := range pageChan {
		var lastErr error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			if err := limiter.Wait(ctx); err != nil {
				errorsChan <- fmt.Errorf("rate limiter wait failed on page %d (attempt %d): %w", page, attempt, err)
				break
			}

			pageOpts := &github.ActivityListStarredOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: c.config.itemsPerPage,
				},
			}

			starredPage, _, err := c.client.Activity.ListStarred(ctx, username, pageOpts)
			if err != nil {
				lastErr = err
				if attempt < maxRetries {
					backoffDuration := time.Duration(attempt) * time.Second
					fmt.Printf("Failed to fetch page %d (attempt %d/%d): %v - retrying in %v...\n",
						page, attempt, maxRetries, err, backoffDuration)
					time.Sleep(backoffDuration)
					continue
				}
				errorsChan <- fmt.Errorf("failed to fetch page %d after %d attempts: %w", page, maxRetries, lastErr)
				break
			}

			mu.Lock()
			for _, starredRepo := range starredPage {
				c.StarredRepos = append(c.StarredRepos, starredRepo.Repository)
			}
			mu.Unlock()

			fmt.Printf("Fetched page %d with %d repositories\n", page, len(starredPage))
			break
		}
	}
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

func (c *GitHubClient) WriteIndex() error {
	filename := "README.md"
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	_, err = fmt.Fprintf(writer, "# awesome-repositories\n")
	if err != nil {
		return fmt.Errorf("failed to write header to %s: %w", filename, err)
	}

	sortedLanguages := slices.Collect(maps.Keys(c.Repos.ByLanguage))
	slices.SortFunc(sortedLanguages, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	totalRepos := 0
	for _, language := range sortedLanguages {
		totalRepos += len(c.Repos.ByLanguage[language])
	}

	for _, language := range sortedLanguages {
		repos := c.Repos.ByLanguage[language]
		percentage := float64(len(repos)) / float64(totalRepos) * 100
		_, err = fmt.Fprintf(writer, "- [%s](stars/byLanguage/%s.md) (%d repositories, %.2f%%)\n", language, u.SanitizeLanguage(language), len(repos), percentage)
		if err != nil {
			return fmt.Errorf("failed to write language %s to %s: %w", language, filename, err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush %s: %w", filename, err)
	}

	return nil
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
