package main

import (
	"fmt"
	"os"

	"github.com/containifyci/go-self-update/pkg/updater"
)

const (
	repoOwner = "containifyci"
	repoName  = "go-self-update"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	fmt.Printf("go-self-update %s, commit %s, built at %s\n", version, commit, date)
	updater := updater.NewUpdater(repoName, repoOwner, repoName, version)
	updated, err := updater.SelfUpdate()
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
		os.Exit(1)
	}
	if updated {
		fmt.Println("Update completed successfully!")
		return
	}
	fmt.Println("Already up-to-date")
}
