package builder

func Filled(x, y, z int, modelBytes []byte) bool {
	r := int(modelBytes[0])
	targetBit := x*r*r + y*r + z
	return (modelBytes[1+targetBit/8] & (1 << uint(targetBit%8))) != 0
}

func HighestAlongX(y, z int, modelBytes []byte) int {
	r := int(modelBytes[0])
	for x := r - 1; x >= 0; x-- {
		if Filled(x, y, z, modelBytes) {
			return x
		}
	}
	return -1
}

func LowestAlongX(y, z int, modelBytes []byte) int {
	r := int(modelBytes[0])
	for x := 0; x < r; x++ {
		if Filled(x, y, z, modelBytes) {
			return x
		}
	}

	return -1
}

func HighestAlongY(x, z int, modelBytes []byte) int {
	r := int(modelBytes[0])
	for y := r - 1; y >= 0; y-- {
		if Filled(x, y, z, modelBytes) {
			return y
		}
	}
	return -1
}

func LowestAlongY(x, z int, modelBytes []byte) int {
	r := int(modelBytes[0])
	for y := 0; y < r; y++ {
		if Filled(x, y, z, modelBytes) {
			return y
		}
	}

	return -1
}

func HighestAlongZ(x, y int, modelBytes []byte) int {
	r := int(modelBytes[0])
	for z := r - 1; z >= 0; z-- {
		if Filled(x, y, z, modelBytes) {
			return z
		}
	}
	return -1
}

func LowestAlongZ(x, y int, modelBytes []byte) int {
	r := int(modelBytes[0])
	for z := 0; z < r; z++ {
		if Filled(x, y, z, modelBytes) {
			return z
		}
	}

	return -1
}
