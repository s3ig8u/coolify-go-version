package main

import "fmt"

const (
	Version   = "v1.4.0"
	BuildTime = "2025-06-20T11:22:00Z"
	GitCommit = "production-ready"
)

func printVersion() {
	fmt.Printf("Coolify Go v%s (built %s, commit %s)\n", Version, BuildTime, GitCommit)
}
