package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/fmjstudios/gopskit/pkg/stamp"
	"github.com/fmjstudios/gopskit/pkg/tools"
	"github.com/fmjstudios/gopskit/pkg/util"
)

func main() {
	// YAML Merge

	// path := os.Args[1]

	// mp, err := tools.AddSecretValue(path, map[string]interface{}{
	// 	"hooks": map[string]interface{}{
	// 		"awxToken":       "fick dich",
	// 		"kubescapeToken": "bastard",
	// 		"vaultToken":     "thisIsANewValue",
	// 	},
	// }, true)

	// if err != nil {
	// 	panic(err.Error())
	// }

	// for k, v := range mp {
	// 	fmt.Printf("Key: %s - Value: %s\n", k, v)
	// }

	// content, err := yaml.Marshal(mp)
	// if err != nil {
	// 	panic(err)
	// }

	// if err := os.WriteFile("/tmp/gopskit-test/fillr-out-values.yaml", content, 0644); err != nil {
	// 	panic(err)
	// }

	// GIT

	// dir, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }

	// git, err := filesystem.RevParseGitRoot(dir)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("Found Git directory at: %s\n", git)

	// SmallStep
	res, err := tools.GenerateStepValues()
	if err != nil {
		log.Fatal(err)
	}

	mp, err := tools.AddSecretStepValues(res, util.GeneratePassphrase(util.WithLength(48)), os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range mp {
		fmt.Printf("Key: %s - Value: %s\n", k, v)
	}
}
