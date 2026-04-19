package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/semver"
)

type release struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Updater struct {
	BinaryName string
	RepoOwner  string
	RepoName   string
	Version    string
	UpdateHook func() error
	httpClient *http.Client
	baseURL    string
}

type Option func(*Updater)

func WithHTTPClient(client *http.Client) Option {
	return func(u *Updater) {
		u.httpClient = client
	}
}

func withBaseURL(url string) Option {
	return func(u *Updater) {
		u.baseURL = url
	}
}

func WithUpdateHook(fnc func() error) Option {
	return func(u *Updater) {
		u.UpdateHook = fnc
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
		httpClient: http.DefaultClient,
		baseURL:    "https://api.github.com",
	}

	for _, opt := range opts {
		opt(u)
	}
	return u
}

func (u *Updater) SelfUpdate() (bool, error) {
	ctx := context.Background()

	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", u.baseURL, u.RepoOwner, u.RepoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return false, fmt.Errorf("failed to decode release response: %w", err)
	}

	slog.Debug("Latest release", "tag", rel.TagName)

	if !CompareVersions(u.Version, rel.TagName) {
		slog.Debug("Already up-to-date")
		return false, nil
	}

	assetURL, err := findAssetURL(u.BinaryName, rel.Assets)
	if err != nil {
		return false, err
	}

	tmpFile, err := downloadAsset(assetURL)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			slog.Debug("Failed to remove temp file", "error", err)
		}
	}()

	currentBinary, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get current executable path: %w", err)
	}

	if err := replaceBinary(tmpFile, currentBinary); err != nil {
		return false, err
	}

	if u.UpdateHook != nil {
		if err := u.UpdateHook(); err != nil {
			return false, fmt.Errorf("failed to run update hook: %w", err)
		}
	}

	return true, nil
}

func findAssetURL(binaryName string, assets []releaseAsset) (string, error) {
	slog.Debug("Finding asset URL")
	for _, asset := range assets {
		slog.Debug("Checking asset", "name", asset.Name, "url", asset.BrowserDownloadURL)
		if matchBinaryName(binaryName, asset.Name) {
			return asset.BrowserDownloadURL, nil
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Debug("Failed to close response body", "error", err)
		}
	}()

	tmpFile, err := os.CreateTemp("", "go-self-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			slog.Debug("Failed to close temp file", "error", err)
		}
	}()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save asset: %w", err)
	}

	return tmpFile.Name(), nil
}

func replaceBinary(src, dst string) error {
	// Ensure the target binary directory exists
	targetDir := filepath.Dir(dst)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("failed target directory doesn't exists: %w", err)
	}

	// Create the target file in the same directory as the target binary
	targetTemp := dst + ".tmp"

	err := copyFile(src, targetTemp)
	if err != nil {
		return fmt.Errorf("failed to copy temp binary: %w", err)
	}

	defer func() {
		if err := os.Remove(targetTemp); err != nil {
			slog.Debug("Failed to remove target temp file", "error", err)
		}
	}()
	if err := os.Rename(targetTemp, dst); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	if err := os.Chmod(dst, 0755); err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			slog.Debug("Failed to close source file", "error", err)
		}
	}()

	// Open destination file
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open destination file: %w", err)
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			slog.Debug("Failed to close destination file", "error", err)
		}
	}()

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
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
