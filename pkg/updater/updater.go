package updater

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/google/go-github/v66/github"
	"golang.org/x/mod/semver"
)

type Updater struct {
	BinaryName string
	RepoOwner  string
	RepoName   string
	Version    string

	client *github.Client
}

type Option func(*Updater)

func WithClient(client *github.Client) Option {
	return func(u *Updater) {
		u.client = client
	}
}

func NewUpdater(BinaryName, repoOwner, repoName, version string, opts ...Option) *Updater {

	logOpts := slog.HandlerOptions{
		Level:       slog.LevelDebug,
		AddSource:   false,
		ReplaceAttr: nil,
	}

	slogger := slog.New(slog.NewTextHandler(os.Stdout, &logOpts))
	slog.SetDefault(slogger)

	u := &Updater{
		BinaryName: BinaryName,
		RepoOwner:  repoOwner,
		RepoName:   repoName,
		Version:    version,
		client:     github.NewClient(nil),
	}

	for _, opt := range opts {
		opt(u)
	}
	return u
}

func (u *Updater) SelfUpdate() (bool, error) {
	ctx := context.Background()
	release, _, err := u.client.Repositories.GetLatestRelease(ctx, u.RepoOwner, u.RepoName)
	if err != nil {
		return false, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	slog.Debug("Latest release", "tag", *release.TagName)

	if !CompareVersions(u.Version, *release.TagName) {
		slog.Debug("Already up-to-date")
		return false, nil
	}

	assetURL, err := findAssetURL(u.BinaryName, release.Assets)
	if err != nil {
		return false, err
	}

	tmpFile, err := downloadAsset(assetURL)
	if err != nil {
		return false, err
	}
	defer os.Remove(tmpFile)

	currentBinary, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get current executable path: %w", err)
	}

	if err := replaceBinary(tmpFile, currentBinary); err != nil {
		return false, err
	}

	return true, nil
}

func findAssetURL(binaryName string, assets []*github.ReleaseAsset) (string, error) {
	slog.Debug("Finding asset URL")
	for _, asset := range assets {
		if asset.Name != nil && asset.BrowserDownloadURL != nil {
			slog.Debug("Checking asset", "name", *asset.Name, "url", *asset.BrowserDownloadURL)
			if matchBinaryName(binaryName, *asset.Name) {
				return *asset.BrowserDownloadURL, nil
			}
		}
	}
	return "", fmt.Errorf("no matching binary found for current %s %s", runtime.GOOS, runtime.GOARCH)
}

func matchBinaryName(binaryName, name string) bool {
	return name == fmt.Sprintf("%s_%s_%s", binaryName, runtime.GOOS, runtime.GOARCH)
}

func downloadAsset(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "go-self-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save asset: %w", err)
	}

	return tmpFile.Name(), nil
}

func replaceBinary(newBinary, currentBinary string) error {
	if err := os.Rename(newBinary, currentBinary); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	if err := os.Chmod(currentBinary, 0755); err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	return nil
}

func normalizeVersion(version string) string {
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

func CompareVersions(v1, v2 string) bool {
	v1 = normalizeVersion(v1)
	v2 = normalizeVersion(v2)

	if !semver.IsValid(v1) {
		return true
	}

	if !semver.IsValid(v2) {
		panic(fmt.Sprintf("invalid version format: v1=%s, v2=%s", v1, v2))
	}

	ret := semver.Compare(v1, v2)
	if ret == 0 || ret == 1 {
		return false
	}
	return true
}
