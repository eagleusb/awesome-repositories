package main

import (
	"flag"
	"fmt"

	"github.com/eagleusb/awesome-repositories/github"
)

var (
	version  = "dev"
	help     = flag.Bool("help", false, "show help")
	username = flag.String("username", "", "GitHub username")
	page     = flag.Int("page-size", 100, "number of repositories to fetch per page")
	limit    = flag.Int("limit", 1000, "maximum number of repositories to fetch")
)

func init() {}

func main() {
	flag.Parse()

	if *help || *username == "" {
		flag.Usage()
		return
	}

	fmt.Printf("Connecting to gh as: %s\n", *username)

	githubClient, err := github.NewGitHubClient()
	if err != nil {
		fmt.Printf("Error creating gh client: %v\n", err)
		return
	}

	repos, err := githubClient.GetStarredRepos(*username, *limit, *page)
	if err != nil {
		fmt.Printf("Error fetching starred repositories with %v\n", err)
	}

	repos.ClassifyRepos()
	for language, repo := range githubClient.Repos.ByLanguage {
		fmt.Printf("%s: %d\n", language, len(repo))
	}
}
