package updater

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdater(t *testing.T) {
	u := NewUpdater("test", "test", "test", "test")
	assert.NotNil(t, u)
}

func TestSelfUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock binary data"))
	}))
	defer server.Close()
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			github.RepositoryRelease{
				Name:    github.String("test"),
				TagName: github.String("v1.0.0"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.String(fmt.Sprintf("test_%s_%s", runtime.GOOS, runtime.GOARCH)),
						BrowserDownloadURL: github.String(server.URL),
					},
				},
			},
		),
	)
	c := github.NewClient(mockedHTTPClient)
	u := NewUpdater("test", "test", "test", "test", WithClient(c))
	updated, err := u.SelfUpdate()
	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		version, want string
	}{
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
		{"", "v"},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := normalizeVersion(tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.1", true},
		{"1.0.1", "1.0.0", false},
		{"1.0.0", "1.1.0", true},
		{"1.1.0", "1.0.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.v1+"-"+tt.v2, func(t *testing.T) {
			got := CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.want, got)
		})
	}
}
