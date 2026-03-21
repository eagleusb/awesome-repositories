package md

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/eagleusb/awesome-repositories/types"
	u "github.com/eagleusb/awesome-repositories/utils"
)

// Writer handles writing markdown files for GitHub repositories
type Writer struct {
	repos map[string][]*types.Repo
}

// NewWriter creates a new Writer with the given repository map
func NewWriter(repos map[string][]*types.Repo) *Writer {
	return &Writer{repos: repos}
}

// WriteIndex generates the main README.md with language index
func (w *Writer) WriteIndex(filename string, topN int) error {
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

	// Helper struct for sorting languages by percentage
	type languageData struct {
		name       string
		count      int
		percentage float64
	}

	sortedLanguages := slices.Collect(maps.Keys(w.repos))
	slices.SortFunc(sortedLanguages, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	totalRepos := 0
	for _, language := range sortedLanguages {
		totalRepos += len(w.repos[language])
	}

	// Create language data slice with percentages
	languageStats := make([]languageData, 0, len(sortedLanguages))
	for _, language := range sortedLanguages {
		repos := w.repos[language]
		percentage := float64(len(repos)) / float64(totalRepos) * 100
		languageStats = append(languageStats, languageData{
			name:       language,
			count:      len(repos),
			percentage: percentage,
		})
	}

	// Sort by percentage descending
	topLanguages := make([]languageData, len(languageStats))
	copy(topLanguages, languageStats)
	slices.SortFunc(topLanguages, func(a, b languageData) int {
		if a.percentage > b.percentage {
			return -1
		} else if a.percentage < b.percentage {
			return 1
		}
		return 0
	})

	// Write top N languages
	topN = min(topN, len(topLanguages))

	if topN > 0 {
		_, err = fmt.Fprintf(writer, "\n## Top %d Languages\n", topN)
		if err != nil {
			return fmt.Errorf("failed to write top languages header to %s: %w", filename, err)
		}

		for i := range topN {
			lang := topLanguages[i]
			_, err = fmt.Fprintf(writer, "- [%s](stars/byLanguage/%s.md) (%d repositories, %.2f%%)\n",
				lang.name, u.SanitizeLanguage(lang.name), lang.count, lang.percentage)
			if err != nil {
				return fmt.Errorf("failed to write top language %s to %s: %w", lang.name, filename, err)
			}
		}
	}

	// Write all languages alphabetically
	_, err = fmt.Fprintf(writer, "\n## All Languages\n")
	if err != nil {
		return fmt.Errorf("failed to write all languages header to %s: %w", filename, err)
	}

	for _, lang := range languageStats {
		_, err = fmt.Fprintf(writer, "- [%s](stars/byLanguage/%s.md) (%d repositories, %.2f%%)\n",
			lang.name, u.SanitizeLanguage(lang.name), lang.count, lang.percentage)
		if err != nil {
			return fmt.Errorf("failed to write language %s to %s: %w", lang.name, filename, err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush %s: %w", filename, err)
	}

	return nil
}

// WriteRepos generates individual markdown files for each language
func (w *Writer) WriteRepos(dir string) error {
	if err := u.EnsureDirectory(dir); err != nil {
		return err
	}

	for language, repos := range w.repos {
		sanitized := u.SanitizeLanguage(language)
		filename := filepath.Join(dir, sanitized+".md")

		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)

		sortedRepos := make([]*types.Repo, len(repos))
		copy(sortedRepos, repos)
		slices.SortFunc(sortedRepos, func(a, b *types.Repo) int {
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
