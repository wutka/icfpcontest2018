package main

import (
	"os"
	"io/ioutil"
	"wutka.com/icfpc/modelers"
	"fmt"
)

func main() {
	for _, filename := range os.Args[1:] {
		modelBytes, err := ioutil.ReadFile(filename)
//		fmt.Printf("%s: Dimension %d\n", filename, int(modelBytes[0]))
		if err != nil {
			fmt.Printf("Can't load file %s: %+v\n", filename, err)
			continue
		}
		if modelers.CanBuildBottomUp(modelBytes) {
			fmt.Printf("%s\n", filename)
		}
	}
}
