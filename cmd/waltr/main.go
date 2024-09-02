package main

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/cmd"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
)

func main() {
	a := app.New()
	rCmd := cmd.NewRootCommand(a)

	if err := rCmd.Execute(); err != nil {
		a.Logger.Sugar().Fatalf("%s exited with error: %v\n", a.Name, err)
	}
}
