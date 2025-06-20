package main

import "fmt"

const (
	Version   = "v1.3.0"
	BuildTime = "2025-06-20T10:02:00Z"
	GitCommit = "abc123"
)

func printVersion() {
	fmt.Printf("Coolify Go v%s (built %s, commit %s)\n", Version, BuildTime, GitCommit)
}
