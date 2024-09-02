package main

import (
	"log"

	"github.com/fmjstudios/gopskit/internal/fillr/app"
	"github.com/fmjstudios/gopskit/internal/fillr/cmd"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
)

func main() {
	rCmd := cmd.NewRootCommand()

	if err := rCmd.Execute(); err != nil {
		log.Fatalf("%s exited with error: %v\n", app.APP_NAME, err)
	}
}
