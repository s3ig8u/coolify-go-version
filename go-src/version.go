package main

import "fmt"

const (
	Version   = "v1.4.1"
	BuildTime = "2025-06-21T17:13:42Z"
	GitCommit = "a2624e7"
)

func printVersion() {
	fmt.Printf("Coolify Go v%s (built %s, commit %s)\n", Version, BuildTime, GitCommit)
}
