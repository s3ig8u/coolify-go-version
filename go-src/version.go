package main

import "fmt"

const (
	Version   = "v1.2.0"
	BuildTime = "2025-06-19T20:30:00Z"
	GitCommit = "def456"
)

func printVersion() {
	fmt.Printf("Coolify Go v%s (built %s, commit %s)\n", Version, BuildTime, GitCommit)
}
