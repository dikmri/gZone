package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/minio/selfupdate"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	repoOwner = "dikmri"
	repoName  = "gZone"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetVersion returns the current version string to the frontend.
func (a *App) GetVersion() string {
	return Version
}

// StartUpdateCheck is called by the frontend after event listeners are set up.
// Running in a goroutine avoids blocking the UI.
func (a *App) StartUpdateCheck() {
	go a.checkAndApplyUpdate()
}

func (a *App) checkAndApplyUpdate() {
	// dev ビルドはアップデートチェックをスキップ
	if Version == "dev" {
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}
	req.Header.Set("User-Agent", "gZone/"+Version)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}
	defer resp.Body.Close()

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}

	if !isNewerVersion(rel.TagName, Version) {
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}

	// 現在のプラットフォーム向けアセットを探す
	want := platformAssetName()
	var dlURL string
	for _, asset := range rel.Assets {
		if asset.Name == want {
			dlURL = asset.BrowserDownloadURL
			break
		}
	}
	if dlURL == "" {
		// 対応アセットなし → 通常起動
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}

	runtime.EventsEmit(a.ctx, "update:available", rel.TagName)

	// ダウンロード
	dlReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, dlURL, nil)
	if err != nil {
		runtime.EventsEmit(a.ctx, "update:error", err.Error())
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}
	dlReq.Header.Set("User-Agent", "gZone/"+Version)

	dlResp, err := client.Do(dlReq)
	if err != nil {
		runtime.EventsEmit(a.ctx, "update:error", err.Error())
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}
	defer dlResp.Body.Close()

	pr := &progressReader{r: dlResp.Body, total: dlResp.ContentLength, ctx: a.ctx}

	// バイナリ置換
	if err := selfupdate.Apply(pr, selfupdate.Options{}); err != nil {
		runtime.EventsEmit(a.ctx, "update:error", err.Error())
		runtime.EventsEmit(a.ctx, "update:none")
		return
	}

	// macOS: Gatekeeper 検疫属性を除去
	if goruntime.GOOS == "darwin" {
		exe, _ := os.Executable()
		exec.Command("xattr", "-d", "com.apple.quarantine", exe).Run()
	}

	runtime.EventsEmit(a.ctx, "update:done", rel.TagName)
	time.Sleep(1500 * time.Millisecond)

	// 再起動
	exe, err := os.Executable()
	if err == nil {
		cmd := exec.Command(exe, os.Args[1:]...)
		if err := cmd.Start(); err == nil {
			runtime.Quit(a.ctx)
			return
		}
	}
	// 再起動失敗 → 通常起動を続ける
	runtime.EventsEmit(a.ctx, "update:none")
}

// progressReader wraps io.Reader and emits update:progress events (0-100).
type progressReader struct {
	r       io.Reader
	total   int64
	read    int64
	lastPct int
	ctx     context.Context
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 && pr.total > 0 {
		pr.read += int64(n)
		pct := int(pr.read * 100 / pr.total)
		if pct != pr.lastPct {
			pr.lastPct = pct
			runtime.EventsEmit(pr.ctx, "update:progress", pct)
		}
	}
	return n, err
}

// platformAssetName returns the expected GitHub release asset name for the current OS/arch.
func platformAssetName() string {
	name := fmt.Sprintf("gZone_%s_%s", goruntime.GOOS, goruntime.GOARCH)
	if goruntime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// isNewerVersion returns true when latestTag (e.g. "v0.2.0") is newer than current (e.g. "v0.1.0").
func isNewerVersion(latestTag, current string) bool {
	l := parseVer(strings.TrimPrefix(latestTag, "v"))
	c := parseVer(strings.TrimPrefix(current, "v"))
	for i := range l {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}

func parseVer(v string) [3]int {
	parts := strings.SplitN(v, ".", 3)
	var r [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		r[i], _ = strconv.Atoi(parts[i])
	}
	return r
}
