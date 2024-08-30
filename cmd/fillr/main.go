package main

import (
	"fmt"
	"os"

	_ "github.com/fmjstudios/gopskit/pkg/stamp"
	"github.com/fmjstudios/gopskit/pkg/tools"
)

func main() {
	// m := map[string]interface{}{
	// 	"hello":      "from where?",
	// 	"definitely": "from the other side",
	// }

	// m2 := map[string]interface{}{
	// 	"hello":      "from where, by who?",
	// 	"mostLikely": "from the other side",
	// }

	// err := util.DeepMergeMap(m, m2)
	// if err != nil {
	// 	panic(err.Error())
	// }

	path := os.Args[1]
	mp, err := tools.AddSecretValue(path, ".whatever", true)
	if err != nil {
		panic(err.Error())
	}

	for k, v := range mp {
		fmt.Printf("Key: %s - Value: %s\n", k, v)
	}

	// value, err := tools.GetSecretValue(path, ".kubectl.image.tag", true)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// fmt.Printf("value: %s\n", value)
}
