package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"wutka.com/icfpc/builder"
)

func main() {
	for _, filename := range os.Args[1:] {
		modelBytes, err := ioutil.ReadFile(filename)
		//		fmt.Printf("%s: Dimension %d\n", filename, int(modelBytes[0]))
		if err != nil {
			fmt.Printf("Can't load file %s: %+v\n", filename, err)
			continue
		}

		r := int(modelBytes[0])
		for y := 0; y < r; y++ {
			for z := 0; z < r; z++ {
				for x := 0; x < r; x++ {
					if builder.Filled(x, y, z, modelBytes) {
						fmt.Printf("*")
					} else {
						fmt.Printf(" ")
					}
				}
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
			for x := 0; x < r; x++ {
				fmt.Printf("-")
			}
			fmt.Printf("\n")
		}
	}
}
