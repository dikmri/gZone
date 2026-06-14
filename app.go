package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	videoServer *http.Server
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(_ context.Context) {
	if a.videoServer != nil {
		a.videoServer.Close()
	}
}

var videoExts = map[string]bool{
	".mp4":  true,
	".webm": true,
	".ogv":  true,
	".mov":  true,
	".mkv":  true,
	".avi":  true,
	".m4v":  true,
	".flv":  true,
	".wmv":  true,
	".ts":   true,
}

// SelectFolder opens a native directory picker dialog.
func (a *App) SelectFolder() string {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "動画フォルダを選択",
	})
	if err != nil || path == "" {
		return ""
	}
	return path
}

// StartVideoServer closes any prior server, starts a new local HTTP file server
// for the given folder, and returns the port number.
func (a *App) StartVideoServer(folderPath string) (int, error) {
	if a.videoServer != nil {
		a.videoServer.Close()
		a.videoServer = nil
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port

	fs := http.FileServer(http.Dir(folderPath))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Range")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		fs.ServeHTTP(w, r)
	})

	a.videoServer = &http.Server{Handler: mux}
	go func() {
		_ = a.videoServer.Serve(listener)
	}()

	return port, nil
}

// GetVideoFiles returns a paginated slice of video filenames sorted alphabetically.
func (a *App) GetVideoFiles(folderPath string, offset, limit int) ([]string, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if videoExts[strings.ToLower(filepath.Ext(e.Name()))] {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	if offset >= len(names) {
		return []string{}, nil
	}
	end := offset + limit
	if end > len(names) {
		end = len(names)
	}
	return names[offset:end], nil
}

// GetTotalVideoCount returns the total count of video files in the folder.
func (a *App) GetTotalVideoCount(folderPath string) (int, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && videoExts[strings.ToLower(filepath.Ext(e.Name()))] {
			count++
		}
	}
	return count, nil
}
