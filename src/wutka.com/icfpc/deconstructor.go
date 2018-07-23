package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"wutka.com/icfpc/builder"
	"wutka.com/icfpc/modelers"
)

func main() {
	traceName := os.Args[2]
	ioutil.WriteFile(traceName, []byte{}, 0644)

	modelName := os.Args[1]
	modelBytes, err := ioutil.ReadFile(modelName)
	if err != nil {
		fmt.Printf("Error loading file %s: %+v\n", modelName, err)
		return
	}

	fmt.Printf("Processing %s\n", modelName)

	processor := &modelers.Ledges{}

	startBot := builder.NewFromModel(modelBytes)
	b := startBot.GetBuilder()

	go processor.Model(modelBytes, startBot)

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
	reverseTrace = append(reverseTrace, 0xff) // Halt at the end

	traceName = os.Args[2]
	ioutil.WriteFile(traceName, reverseTrace, 0644)
}

func reverseNearDir(nd int) int {
	x := (nd / 9) - 1
	y := ((nd % 9) / 3) - 1
	z := (nd % 3) - 1
	return (-x+1)*9 + (-y+1)*3 + -z + 1
}
