package main

import (
	"github.com/fmjstudios/gopskit/internal/ssolo/app"
	"github.com/fmjstudios/gopskit/internal/ssolo/cmd"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
	"log"
)

func main() {
	kern, err := app.New()
	if err != nil {
		log.Fatalf("could not initialize %s: %v", kern.Name, err)
	}

	cmdRoot := cmd.NewRootCommand(kern)
	if err := cmdRoot.Execute(); err != nil {
		kern.Log.Fatalf("%s exited with error: %v\n", kern.Name, err)
	}
}
