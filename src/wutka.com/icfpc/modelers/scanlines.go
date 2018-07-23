package modelers

import (
	"fmt"
	"wutka.com/icfpc/builder"
)

type Scanlines struct {
}

type OrderGoTo struct {
	prim builder.Coord
	sec  builder.Coord
}

type OrderMakeLine struct {
	prim builder.Coord
	sec  builder.Coord
}

type OrderWaitForSupport struct {
	c builder.Coord
}

type OrderFusion struct {
	prim builder.Coord
	sec  builder.Coord
}

func RunScanline(modelBytes []byte, bot, partner *builder.Bot, isPrimary bool) {
	/*
		for {
			command := bot.GetOrder()

			switch c := command.(type) {
			case OrderGoTo:
			case OrderMakeLine:
			case OrderWaitForSupport:
			case OrderFusion:
			}
		}
	*/
}

func (b *Scanlines) Model(modelBytes []byte, startBot *builder.Bot) {
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
