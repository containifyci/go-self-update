package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpdater(t *testing.T) {
	u := NewUpdater("test", "test", "test", "test")
	assert.NotNil(t, u)
}

func TestSelfUpdate(t *testing.T) {
	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock binary data"))
	}))
	defer assetServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/test/test/releases/latest" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(release{
				TagName: "v1.0.0",
				Assets: []releaseAsset{
					{
						Name:               fmt.Sprintf("test_%s_%s", runtime.GOOS, runtime.GOARCH),
						BrowserDownloadURL: assetServer.URL,
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer apiServer.Close()

	u := NewUpdater("test", "test", "test", "test", withBaseURL(apiServer.URL))
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
