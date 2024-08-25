package main

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/vaultr/cmd"
	"github.com/fmjstudios/gopskit/pkg/stamp"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
)

func main() {
	fmt.Printf("Running vaultr at version: %s\n", stamp.Version)
	fmt.Printf("Running vaultr with Go version: %s\n", stamp.GoVersion)
	fmt.Printf("Running vaultr built at: %s\n", stamp.BuildDate)
	fmt.Printf("Running vaultr at commit SHA: %s\n", stamp.CommitSHA)
	fmt.Printf("Running vaultr on branch: %s\n", stamp.Branch)
	fmt.Printf("Running vaultr built on: %s\n", stamp.Platform)
	cmd.Execute()
}
