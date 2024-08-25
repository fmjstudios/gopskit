package main

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/opsctl/cmd"
	"github.com/fmjstudios/gopskit/pkg/stamp"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
)

func main() {
	fmt.Printf("Running opsctl at version: %s\n", stamp.Version)
	fmt.Printf("Running opsctl with Go version: %s\n", stamp.GoVersion)
	fmt.Printf("Running opsctl built at: %s\n", stamp.BuildDate)
	fmt.Printf("Running opsctl at commit SHA: %s\n", stamp.CommitSHA)
	fmt.Printf("Running opsctl on branch: %s\n", stamp.Branch)
	fmt.Printf("Running opsctl built on: %s\n", stamp.Platform)
	cmd.Execute()
}
