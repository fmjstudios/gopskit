package main

import (
	"fmt"

	_ "github.com/fmjstudios/gopskit/pkg/stamp"
	"github.com/fmjstudios/gopskit/pkg/tools"
)

func main() {
	// dir, err := filesystem.TempDir("fillr")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer filesystem.Remove(dir)

	// fmt.Printf("Created temporary directory: %s", dir)

	// Tools
	m, err := tools.Find()
	if err != nil {
		panic(err)
	}

	for k, v := range m {
		fmt.Printf("Tool: %s - Location: %s\n", k, v)
	}
}
