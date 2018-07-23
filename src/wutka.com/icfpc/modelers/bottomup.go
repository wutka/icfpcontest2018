package modelers

import (
	"fmt"
	"wutka.com/icfpc/builder"
)

type BottomUp struct {
}

func CanBuildBottomUp(modelBytes []byte) bool {
	r := int(modelBytes[0])
	if r < 4 {
		return true
	}
	for z := 1; z < r-1; z++ {
		for y := 1; y < r-1; y++ {
			for x := 1; x < r-1; x++ {
				if builder.Filled(x, y, z, modelBytes) && !builder.Filled(x, y-1, z, modelBytes) {
					//					fmt.Printf("Failed at %d,%d,%d\n", x, y, z)
					return false
				}
			}
		}
	}
	return true
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		} else {
			return c
		}
	} else if c < b {
		return c
	} else {
		return b
	}
}

func max3(a, b, c int) int {
	if a > b {
		if a > c {
			return a
		} else {
			return c
		}
	} else if c > b {
		return c
	} else {
		return b
	}
}

func getNextStartPos(y, currZ, zDir int, modelBytes []byte) (int, int, int) {
	r := int(modelBytes[0])
	x1 := -1
	x2 := -1
	x3 := -1
	beginZ := currZ + 3*zDir
	endZ := r - 1
	if zDir < 0 {
		endZ = 1
	}
	for startZ := beginZ; startZ != endZ; startZ += zDir {
		x1 = builder.LowestAlongX(y-1, startZ, modelBytes)
		x2 = builder.LowestAlongX(y, startZ, modelBytes)
		x3 = builder.LowestAlongX(y+1, startZ, modelBytes)
		if x1 < 0 && x2 < 0 && x3 < 0 {
			continue
		}
		minX := min3(x1, x2, x3)

		x1 = builder.HighestAlongX(y-1, startZ, modelBytes)
		x2 = builder.HighestAlongX(y, startZ, modelBytes)
		x3 = builder.HighestAlongX(y+1, startZ, modelBytes)

		if x1 < 0 && x2 < 0 && x3 < 0 {
			continue
		}

		maxX := max3(x1, x2, x3)

		return minX, maxX, startZ
	}

	return -1, -1, -1
}

func (b *BottomUp) Model(modelBytes []byte, startBot *builder.Bot) {
	startBot.Start()

	r := int(modelBytes[0])

	xdir := 1
	zdir := 1
	for y := 0; y < r-1; y++ {
		done := true
		for z := 1; done && z < r-1; z++ {
			for x := 1; x < r-1; x++ {
				if builder.Filled(x, y, z, modelBytes) {
					done = false
					break
				}
			}
		}
		if done {
			break
		}
		startZ, endZ := 1, r-1
		if zdir < 0 {
			startZ, endZ = r-2, 1
		}
		for z := startZ; (zdir > 0 && z <= endZ) || (zdir < 0 && z >= endZ); z += 3 * zdir {
			startX, endX := 2, r
			if xdir < 0 {
				startX, endX = r-2, 0
			}
			startBot.Goto(builder.Coord{startX, y, z})
			for x := startX; (xdir > 0 && x < endX) || (xdir < 0 && x >= endX); x += xdir {
				moved := false
				fmt.Printf("x=%d y=%d z=%d, xdir=%d\n", x, y, z, xdir)
				if builder.Filled(x-xdir, y, z, modelBytes) && !startBot.IsFilled(x-xdir, y, z) {
					if !moved {
						startBot.Goto(builder.Coord{x, y, z})
						moved = true
					}
					startBot.Fill(-xdir, 0, 0)
				}
				if z > 0 && builder.Filled(x-xdir, y, z-1, modelBytes) && !startBot.IsFilled(x-xdir, y, z-1) {
					if !moved {
						startBot.Goto(builder.Coord{x, y, z})
						moved = true
					}
					startBot.Fill(-xdir, 0, -1)
				}
				if z < r-1 && builder.Filled(x-xdir, y, z+1, modelBytes) && !startBot.IsFilled(x-xdir, y, z+1) {
					if !moved {
						startBot.Goto(builder.Coord{x, y, z})
						moved = true
					}
					startBot.Fill(-xdir, 0, 1)
				}
			}
			xdir = -xdir
			if (zdir > 0 && z < r-3) || (zdir < 0 && z > 3) {
				startBot.SMove(0, 0, 3*zdir)
			}
		}
		zdir = -zdir
		startBot.SMove(0, 1, 0)
	}
	botPos := startBot.Pos()
	if botPos.X > 0 {
		startBot.Goto(builder.Coord{0, botPos.Y, botPos.Z})
	}
	if botPos.Z > 0 {
		startBot.Goto(builder.Coord{0, botPos.Y, 0})
	}
	startBot.Goto(builder.Coord{0, 0, 0})
	startBot.Halt()

	return
}
