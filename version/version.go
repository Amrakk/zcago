package version

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/mod/semver"

	"github.com/Amrakk/zcago/internal/httpx"
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

func CheckUpdate(sc session.Context) {
	if !sc.CheckUpdate() {
		return
	}

	info := getVersionInfo()
	if info.Version == "" {
		return
	}

	if semver.IsValid(version) && semver.IsValid(info.Version) {
		if semver.Compare(info.Version, version) > 0 {
			httpx.Logger(sc).Infof("zcago: update available: %s → %s (released %s)\n", version, info.Version, info.Time)
		} else {
			httpx.Logger(sc).Info("zcago: up to date\n")
		}
	}
}
