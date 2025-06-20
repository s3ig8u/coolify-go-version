package main

import "fmt"

const (
	Version   = "v1.4.0"
	BuildTime = "2025-06-20T11:38:00Z"
	GitCommit = "azure-registry-v1.4.0"
)

func printVersion() {
	fmt.Printf("Coolify Go v%s (built %s, commit %s)\n", Version, BuildTime, GitCommit)
}
