package main

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/cmd"
	"github.com/fmjstudios/gopskit/pkg/kv"
	_ "github.com/fmjstudios/gopskit/pkg/stamp"
	"log"
	"os"
)

func main() {
	kern, err := app.New()
	if err != nil {
		log.Fatalf("could not initialize waltr application kernel: %v", err)
	}
	defer func(KV *kv.Database) {
		err := KV.Close()
		if err != nil {
			log.Fatalf("could not shut down waltr database connection: %v", err)
		}
	}(kern.KV)

	cmdRoot := cmd.NewRootCommand(kern)
	if err := cmdRoot.Execute(); err != nil {
		kern.Log.Fatalf("%s exited with error: %v\n", kern.Name, err)
	}

	os.Exit(0)
}
