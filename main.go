package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/containifyci/go-self-update/pkg/systemd"
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

	// Define flags
	command := flag.String("command", "", "Command to execute")

	// Parse the flags
	flag.Parse()

	if *command == "update" {
		updater := updater.NewUpdater(repoName, repoOwner, repoName, version, updater.WithUpdateHook(systemd.SystemdRestartHook("go-self-update")))
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
	} else {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
		fmt.Println("Blocking, press ctrl+c to continue...")
		<-done // Will block here until user hits ctrl+c
	}
}
