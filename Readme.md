# go-self-update

The **Updater** library simplifies self-updating binary applications in Go. It automates the process of checking for newer versions of a binary hosted on a GitHub repository, downloading the latest release, and replacing the current executable. Additionally, it supports customizable post-update hooks to integrate application-specific actions.

---

## Features

- **Version Checking**: Compares the current binary version with the latest release on GitHub.
- **Asset Selection**: Dynamically selects the appropriate binary asset based on the operating system and architecture.
- **Automatic Updates**: Downloads and replaces the current binary with the latest release.
- **Customizable Hooks**: Run custom logic after a successful update.
- **SemVer Support**: Ensures accurate version comparison using Semantic Versioning.

---

## Installation

To use the Updater library, include it in your Go project:

```bash
go get github.com/containifyci/go-self-update/pkg/updater
```

---

## Usage

### Basic Example

Below is an example of how to use the library in a Go application:

```go
package main

import (
	"log"
	"github.com/containifyci/go-self-update/pkg/updater"
)

func main() {
	// Initialize the updater
	u := updater.NewUpdater(
		"my-app",             // Binary name
		"my-org",             // GitHub repository owner
		"my-repo",            // GitHub repository name
		"v1.0.0",             // Current version
		updater.WithUpdateHook(func() error { // Custom hook (optional)
			log.Println("Update completed successfully!")
			return nil
		}),
	)

	// Perform the update check
	updated, err := u.SelfUpdate()
	if err != nil {
		log.Fatalf("Failed to update: %v", err)
	}

	if updated {
		log.Println("Application updated to the latest version!")
	} else {
		log.Println("Application is already up-to-date.")
	}
}
```

### Custom Options

The library provides options to customize its behavior:

- **Client Customization**: Use a custom `*github.Client` (e.g., for authentication or API rate limits):

```go
u := updater.NewUpdater("my-app", "my-org", "my-repo", "v1.0.0", updater.WithClient(customClient))
```

- **Post-Update Hook**: Execute a function after an update:

```go
u := updater.NewUpdater("my-app", "my-org", "my-repo", "v1.0.0", updater.WithUpdateHook(myUpdateHook))
```

---

## Key Functions

### `NewUpdater`
Creates and configures an `Updater` instance.

### `SelfUpdate`
Performs the self-update process:
1. Fetches the latest release from GitHub.
2. Compares the current version with the latest version.
3. Downloads and replaces the binary if an update is available.

### `CompareVersions`
Compares two Semantic Version strings.

### `findAssetURL`
Selects the appropriate binary asset from the GitHub release assets.

---

## How It Works

1. **Initialization**: You configure the `Updater` with the binary name, repository details, and current version.
2. **Version Check**: The library queries the latest release from the GitHub API.
3. **Asset Download**: If an update is available, it downloads the correct binary asset matching the OS and architecture.
4. **Binary Replacement**: Safely replaces the running binary with the downloaded one.
5. **Custom Logic**: Optionally, you can execute a post-update hook for further actions.

---

## Error Handling

The library uses structured logging (`slog`) for debugging. You can enable debug logs by configuring the `slog` logger.

---

Start integrating **Updater** today to provide seamless updates for your Go applications! 🎉
