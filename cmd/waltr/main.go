package main

import (
	"log"
	"os"
	"strings"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/cmd"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
)

func main() {
	kern, err := app.New()
	if err != nil {
		log.Fatalf("couldn't initialize %s!\nError: %v", app.Name, err)
	}

	cmdRoot := cmd.NewRootCommand(kern)
	if err := cmdRoot.Execute(); err != nil {
		kern.Log.Fatalf("command: '%s' resulted in error: %v\n", strings.Join(os.Args[1:2], " "), err)
	}

	os.Exit(0)
}
