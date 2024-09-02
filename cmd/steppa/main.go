package main

import (
	"github.com/fmjstudios/gopskit/internal/steppa/app"
	"github.com/fmjstudios/gopskit/internal/steppa/cmd"
)

func main() {
	a := app.New()
	rCmd := cmd.NewRootCommand(a)

	if err := rCmd.Execute(); err != nil {
		a.Logger.Sugar().Fatalf("%s exited with error: %v\n", a.Name, err)
	}
}
