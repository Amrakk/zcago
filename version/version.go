package version

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/mod/semver"

	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/session"
)

const version = "v0.0.1"
const registry = "https://proxy.golang.org/github.com/amrakk/zcago/@latest"

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

func CheckUpdate(sc session.Context) {
	if !sc.CheckUpdate() {
		return
	}

	info := getVersionInfo()
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

func getVersionInfo() response {
	reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, registry, nil)
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
