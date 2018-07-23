package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"wutka.com/icfpc/builder"
	"wutka.com/icfpc/modelers"
)

func main() {
	traceName := os.Args[3]
	ioutil.WriteFile(traceName, []byte{}, 0644)

	oldModelName := os.Args[1]
	oldModelBytes, err := ioutil.ReadFile(oldModelName)
	if err != nil {
		fmt.Printf("Error loading file %s: %+v\n", oldModelName, err)
		return
	}

	fmt.Printf("Processing %s\n", oldModelName)

	processor := &modelers.Ledges{}

	startBot := builder.NewFromModel(oldModelBytes)
	b := startBot.GetBuilder()

	go processor.Model(oldModelBytes, startBot)

	trace := b.GetTraceAsList()

	ok, wrong, wasSet := b.CheckAgainstModel()
	if !ok {
		if wasSet {
			panic(fmt.Sprintf("Generated grid doesn't match model at %d,%d,%d (model didn't have pixel set)", wrong.X, wrong.Y, wrong.Z))
		} else {
			panic(fmt.Sprintf("Generated grid doesn't match model at %d,%d,%d (model had pixel set)", wrong.X, wrong.Y, wrong.Z))
		}
		return
	}

	reverseTrace := []byte{}
	for i := len(trace) - 1; i >= 0; i-- {
		traceBytes := trace[i]
		b := traceBytes[0]
		if b == 0xff {
			// ignore
		} else if b == 0xfe {
			reverseTrace = append(reverseTrace, b)
		} else if b == 0xfd {
			reverseTrace = append(reverseTrace, b)
		} else if b&0xcf == 0x04 && traceBytes[1]&0xe0 == 0x00 {
			i := int(traceBytes[1]&0x1f) - 15

			reverseTrace = append(reverseTrace, b, byte(-i+15))
		} else if b&0x0f == 0x0c {
			i1 := int(traceBytes[1]&0x0f) - 5
			i2 := int((traceBytes[1]&0xf0)>>4) - 5
			reverseTrace = append(reverseTrace, b, byte(-i1+5+((-i2+5)<<4)))
		} else if b&0x7 == 0x7 {
			panic("Can't reverse fission and fusion yet")
			return
		} else if b&0x7 == 0x6 {
			panic("Can't reverse fission and fusion yet")
			return
		} else if b&0x7 == 0x5 {
			panic("Can't reverse fission and fusion yet")
			return
		} else if b&0x7 == 0x3 {
			nd := int(b >> 3)
			reverseTrace = append(reverseTrace, byte(0x2+(nd<<3)))
		} else if b&0x7 == 0x2 {
			nd := int(b >> 3)
			reverseTrace = append(reverseTrace, byte(0x3+(nd<<3)))
		} else if b&0x7 == 0x1 {
			nd := int(b >> 3)
			x := int(traceBytes[1] - 30)
			y := int(traceBytes[2] - 30)
			z := int(traceBytes[3] - 30)
			reverseTrace = append(reverseTrace, byte(nd<<3), byte(x+30), byte(y+30), byte(z+30))
		} else if b&0x7 == 0x0 {
			nd := int(b >> 3)
			x := int(traceBytes[1] - 30)
			y := int(traceBytes[2] - 30)
			z := int(traceBytes[3] - 30)
			reverseTrace = append(reverseTrace, byte(0x1+(nd<<3)), byte(x+30), byte(y+30), byte(z+30))
		} else {
			panic("Unknown opcode in my own trace")
		}
	}

	newModelName := os.Args[2]
	newModelBytes, err := ioutil.ReadFile(newModelName)
	if err != nil {
		fmt.Printf("Error loading file %s: %+v\n", newModelName, err)
		return
	}

	processor = &modelers.Ledges{}

	startBot = builder.NewFromModel(newModelBytes)
	b = startBot.GetBuilder()

	go processor.Model(newModelBytes, startBot)

	reconTrace := b.GetTrace()

	ok, wrong, wasSet = b.CheckAgainstModel()
	if !ok {
		if wasSet {
			panic(fmt.Sprintf("Generated grid doesn't match model at %d,%d,%d (model didn't have pixel set)", wrong.X, wrong.Y, wrong.Z))
		} else {
			panic(fmt.Sprintf("Generated grid doesn't match model at %d,%d,%d (model had pixel set)", wrong.X, wrong.Y, wrong.Z))
		}
		return
	}
	traceName = os.Args[3]

	reverseTrace = append(reverseTrace, reconTrace...)

	ioutil.WriteFile(traceName, reverseTrace, 0644)
}
