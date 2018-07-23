package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"wutka.com/icfpc/builder"
	"wutka.com/icfpc/modelers"
)

func main() {
	f, err := os.Create("blobsprofile.prof")
	if err != nil {
		log.Fatal("Can't open proffile", err)
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Fatal("Can't start profiler", err)
	}
	defer pprof.StopCPUProfile()

	// Clear old trace
	traceName := os.Args[2]
	ioutil.WriteFile(traceName, []byte{}, 0644)

	modelName := os.Args[1]
	fmt.Printf("Processing %s\n", modelName)

	modelBytes, err := ioutil.ReadFile(modelName)
	if err != nil {
		fmt.Printf("Error loading file %s: %+v\n", modelName, err)
		return
	}

	processor := &modelers.Ledges{}

	startBot := builder.NewFromModel(modelBytes)
	b := startBot.GetBuilder()

	go processor.Model(modelBytes, startBot)

	trace := b.GetTrace()

	ok, wrong, wasSet := b.CheckAgainstModel()
	if !ok {
		if wasSet {
			panic(fmt.Sprintf("Generated grid doesn't match model at %d,%d,%d (model didn't have pixel set)", wrong.X, wrong.Y, wrong.Z))
		} else {
			panic(fmt.Sprintf("Generated grid doesn't match model at %d,%d,%d (model had pixel set)", wrong.X, wrong.Y, wrong.Z))
		}
		return
	}
	ioutil.WriteFile(traceName, trace, 0644)
}
