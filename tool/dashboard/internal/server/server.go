package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	goredis "github.com/redis/go-redis/v9"

	"deepx/tool/dashboard/internal/handler"
)

type Server struct {
	rdb        *goredis.Client
	loaderBin  string
	redisAddr  string
	listenAddr string
	distDir    string // frontend build output directory
	statusHub  *handler.StatusHub
}

func New(rdb *goredis.Client, loaderBin, redisAddr, listenAddr string) *Server {
	distDir := findDistDir()
	if distDir == "" {
		log.Fatalf("FATAL: frontend not built — index.html not found.\n" +
			"  Expected next to binary: %s/index.html\n" +
			"  Or at: frontend/dist/index.html\n" +
			"  Run: make build-dashboard",
			exeDir())
	}

	s := &Server{
		rdb:        rdb,
		loaderBin:  loaderBin,
		redisAddr:  redisAddr,
		listenAddr: listenAddr,
		distDir:    distDir,
		statusHub:  handler.NewStatusHub(rdb),
	}
	go s.statusHub.Run()
	handler.RegisterTerminal(rdb, listenAddr)
	log.Printf("[server] frontend loaded from %s", distDir)
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		data, err := handler.GetStatus(s.rdb)
		writeJSON(w, data, err)
	})

	mux.HandleFunc("/api/status/stream", func(w http.ResponseWriter, r *http.Request) {
		s.statusHub.ServeSSE(w, r)
	})

	mux.HandleFunc("/api/vthreads", func(w http.ResponseWriter, r *http.Request) {
		data, err := handler.ListVThreads(s.rdb)
		writeJSON(w, data, err)
	})

	mux.HandleFunc("/api/vthread/", func(w http.ResponseWriter, r *http.Request) {
		handler.GetVThread(s.rdb, w, r)
	})

	mux.HandleFunc("/api/functions", func(w http.ResponseWriter, r *http.Request) {
		data, err := handler.ListFunctions(s.rdb)
		writeJSON(w, data, err)
	})

	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeTerminal(w, r)
	})

	mux.HandleFunc("/api/term/stdout", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeTermStdout(w, r)
	})

	mux.HandleFunc("/api/term/stderr", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeTermStderr(w, r)
	})

	mux.HandleFunc("/api/term/stdin", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeTermStdin(w, r)
	})

	mux.HandleFunc("/api/ops/", func(w http.ResponseWriter, r *http.Request) {
		handler.GetOps(s.rdb, w, r)
	})

	mux.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.Run(s.loaderBin, s.redisAddr, w, r)
	})

	// Static files (React build) + SPA fallback
	fs := http.FileServer(http.Dir(s.distDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := filepath.Join(s.distDir, r.URL.Path)
		if r.URL.Path == "/" {
			path = filepath.Join(s.distDir, "index.html")
		}

		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for all non-file routes
		http.ServeFile(w, r, filepath.Join(s.distDir, "index.html"))
	})

	return withCORS(mux)
}

// ── distDir resolution ──

// findDistDir locates the React frontend build output.
//
//	Deployment: dash-server + index.html in same dir → os.Executable()
//	Development: go run from tool/dashboard/ → CWD-relative fallback
//
// Returns "" if index.html is not found anywhere (caller must fatal).
func findDistDir() string {
	// 1. Deployment: look next to the binary
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			return dir
		}
	}
	// 2. Development: CWD-relative paths
	for _, d := range []string{
		filepath.Join("frontend", "dist"),
		filepath.Join("tool", "dashboard", "frontend", "dist"),
	} {
		if _, err := os.Stat(filepath.Join(d, "index.html")); err == nil {
			return d
		}
	}
	return ""
}

// exeDir returns the directory of the running binary, or "?" if unknown.
func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "?"
	}
	return filepath.Dir(exe)
}

func writeJSON(w http.ResponseWriter, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(data)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
