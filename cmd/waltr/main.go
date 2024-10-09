package main

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/cmd"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
)

func main() {
	kern, err := app.New()
	if err != nil {
		log.Fatalf("could not initialize waltr application kernel: %v", err)
	}

	cmdRoot := cmd.NewRootCommand(kern)
	if err := cmdRoot.Execute(); err != nil {
		kern.Log.Fatalf("%s exited with error: %v\n", kern.Name, err)
	}

	g := new(errgroup.Group)
	g.Go(func() error {
		return kern.KV.Close()
	})

	err = g.Wait()
	if err != nil {
		log.Fatalf("could not shut down waltr database connection: %v", err)
	}

	os.Exit(0)
}
