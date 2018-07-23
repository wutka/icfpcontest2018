package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	for _, filename := range os.Args[1:] {
		fileBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("Unable to read file %s: %+v", err)
			continue
		}

		fmt.Printf("File %s:\n", filename)
		decodeTrace(fileBytes)
		fmt.Printf("\n\n")
	}
}

func dirToString(a, i int) string {
	x, y, z := 0, 0, 0
	switch a {
	case 1:
		x = i
	case 2:
		y = i
	case 3:
		z = i
	}
	return fmt.Sprintf("<%d,%d,%d>", x, y, z)
}

func nearDirToString(nd int) string {
	x := (nd / 9) - 1
	y := ((nd % 9) / 3) - 1
	z := (nd % 3) - 1
	return fmt.Sprintf("<%d,%d,%d>", x, y, z)
}

func decodeTrace(traceBytes []byte) {
	pos := 0

	for pos < len(traceBytes) {
		b := traceBytes[pos]
		b1 := byte(0)
		if pos < len(traceBytes)-1 {
			b1 = traceBytes[pos+1]
		}
		if b == 0xff {
			fmt.Printf("Halt\n")
			pos++
		} else if b == 0xfe {
			fmt.Printf("Wait\n")
			pos++
		} else if b == 0xfd {
			fmt.Printf("Flip\n")
			pos++
		} else if b&0xcf == 0x04 && b1&0xe0 == 0x00 {
			dir := int(b&0x30) >> 4
			i := int(b1&0x1f) - 15

			fmt.Printf("SMove %s\n", dirToString(dir, i))
			pos += 2
		} else if b&0x0f == 0x0c {
			d1 := int(b&0x30) >> 4
			d2 := int(b&0xc0) >> 6
			i1 := int(b1&0x0f) - 5
			i2 := int((b1&0xf0)>>4) - 5
			fmt.Printf("LMove %s %s\n", dirToString(d1, i1), dirToString(d2, i2))
			pos += 2
		} else if b&0x7 == 0x7 {
			nd := int(b >> 3)
			fmt.Printf("FusionP %s\n", nearDirToString(nd))
			pos++
		} else if b&0x7 == 0x6 {
			nd := int(b >> 3)
			fmt.Printf("FusionS %d\n", nearDirToString(nd))
			pos++
		} else if b&0x7 == 0x5 {
			nd := int(b >> 3)
			m := int(b1)
			fmt.Printf("Fission %s %d\n", nearDirToString(nd), m)
			pos++
		} else if b&0x7 == 0x3 {
			nd := int(b >> 3)
			fmt.Printf("Fill %s\n", nearDirToString(nd))
			pos++
		} else if b&0x7 == 0x2 {
			nd := int(b >> 3)
			fmt.Printf("Void %s\n", nearDirToString(nd))
			pos++
		} else if b&0x7 == 0x1 {
			nd := int(b >> 3)
			x := int(b1 - 30)
			y := int(traceBytes[pos+2] - 30)
			z := int(traceBytes[pos+3] - 30)
			fmt.Printf("GFill %s <%d,%d,%d>\n", nearDirToString(nd), x, y, z)
			pos += 4
		} else if b&0x7 == 0x0 {
			nd := int(b >> 3)
			x := int(b1 - 30)
			y := int(traceBytes[pos+2] - 30)
			z := int(traceBytes[pos+3] - 30)
			fmt.Printf("GVoid %s <%d,%d,%d>\n", nearDirToString(nd), x, y, z)
			pos += 4
		} else {
			fmt.Printf("Unknown\n")
			pos++
		}
	}
}
