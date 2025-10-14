package version

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/mod/semver"

	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/session"
)

const (
	version  = "v0.1.0"
	registry = "https://proxy.golang.org/github.com/amrakk/zcago/@latest"
)

type response struct {
	Version string
	Time    string
	Origin  struct {
		VCS  string
		URL  string
		Hash string
		Ref  string
	}
}

func GetVersion() string {
	return version
}

func CheckUpdate(ctx context.Context, sc session.Context) {
	if !sc.CheckUpdate() {
		return
	}

	info := getVersionInfo(ctx)
	if info.Version == "" {
		logger.Log(sc).Debug("No version information available from registry")
		return
	}

	if semver.IsValid(version) && semver.IsValid(info.Version) {
		if semver.Compare(info.Version, version) > 0 {
			logger.Log(sc).Infof("A new version of zcago is available: %s → %s (released %s)", version, info.Version, info.Time)
		} else {
			logger.Log(sc).Info("zcago is up to date")
		}
	} else {
		logger.Log(sc).Warn("Invalid version format detected").
			Debug("Current version valid:", semver.IsValid(version)).
			Debug("Registry version valid:", semver.IsValid(info.Version))
	}
}

func getVersionInfo(ctx context.Context) response {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, registry, nil)
	if err != nil {
		return response{}
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return response{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response{}
	}

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return response{}
	}
	return r
}
