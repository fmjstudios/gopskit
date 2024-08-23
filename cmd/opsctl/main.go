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
	cmd.Execute()
}
