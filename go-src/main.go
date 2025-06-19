// main.go
package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Coolify Go v%s: Hello, world!\nBuild Time: %s\nCommit: %s", Version, BuildTime, GitCommit)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","version":"%s","buildTime":"%s","commit":"%s"}`, Version, BuildTime, GitCommit)
	})

	fmt.Printf("🚀 Coolify Go v%s server running at http://localhost:8080\n", Version)
	http.ListenAndServe(":8080", nil)
}
