package modelers

import (
	"wutka.com/icfpc/builder"
)

type Ledges struct {
}

func (b *Ledges) Model(modelBytes []byte, startBot *builder.Bot) {
	startBot.Start()

	r := int(modelBytes[0])

	xdir := 1
	zdir := 1
	for y := 1; y <= r-1; y++ {
		done := true
		for z := 1; done && z < r-1; z++ {
			for x := 1; x < r-1; x++ {
				if builder.Filled(x, y-1, z, modelBytes) {
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
			startX, endX := 1, r
			if xdir < 0 {
				startX, endX = r-1, 0
			}
			//			startBot.Goto(builder.Coord{startX, y, z})
			for x := startX; (xdir > 0 && x < endX) || (xdir < 0 && x >= endX); x += xdir {
				moved := false
				debugPrintf("x=%d y=%d z=%d, xdir=%d\n", x, y, z, xdir)

				for i := 0; i < 3; i++ {
					if ((zdir < 0 && z < r-1) || ((zdir > 1) && (z > 0))) && builder.Filled(x, y-1, z-zdir, modelBytes) &&
						!startBot.IsFilled(x, y-1, z-zdir) && startBot.GetBuilder().IsSupported(x, y-1, z-zdir) {
						if !moved {
							startBot.Goto(builder.Coord{x, y, z})
							moved = true
						}
						startBot.Fill(0, -1, -zdir)
					}
					if builder.Filled(x, y-1, z, modelBytes) && !startBot.IsFilled(x, y-1, z) && startBot.GetBuilder().IsSupported(x, y-1, z) {
						if !moved {
							startBot.Goto(builder.Coord{x, y, z})
							moved = true
						}
						startBot.Fill(0, -1, 0)
					}
					if ((zdir < 0 && z > 0) || ((zdir > 1) && (z < r-1))) && builder.Filled(x, y-1, z+zdir, modelBytes) &&
						!startBot.IsFilled(x, y-1, z+zdir) && startBot.GetBuilder().IsSupported(x, y-1, z+zdir) {
						if !moved {
							startBot.Goto(builder.Coord{x, y, z})
							moved = true
						}
						startBot.Fill(0, -1, zdir)
					}
				}
			}

			xdir = -xdir
			if (zdir > 0 && z < r-3) || (zdir < 0 && z > 3) {
				//				startBot.Goto(builder.Coord{0, 0, 3 * zdir})
			}
		}

		for {
			ok, c1, c2 := startBot.FindClosestUnconnectedLine(y - 1)
			if !ok {
				break
			}
			debugPrintf("Found unconnected line from %d,%d,%d to %d,%d,%d\n",
				c1.X, c1.Y, c1.Z, c2.X, c2.Y, c2.Z)

			startBot.Goto(builder.Coord{c1.X, y, c1.Z})
			if c1.X == c2.X {
				debugPrintf("Filling along Z axis\n")
				zdir := 1
				if c2.Z < c1.Z {
					zdir = -1
				}
				for z := c1.Z + zdir; (zdir > 0 && z < c2.Z+3) || (zdir < 0 && z > c1.Z-3); z += 3 * zdir {
					moved := false
					targetZ := z - zdir
					if ((zdir > 0 && targetZ <= c2.Z) || (zdir < 0 && targetZ >= c1.Z)) &&
						builder.Filled(c1.X, y-1, targetZ, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y-1, targetZ) && !startBot.IsFilled(c1.X, y-1, targetZ) {
						if !moved {
							startBot.Goto(builder.Coord{c1.X, y, z})
							moved = true
						}
						startBot.Fill(0, -1, -zdir)
					}
					if ((zdir > 0 && z <= c2.Z) || (zdir < 0 && z >= c1.Z)) &&
						builder.Filled(c1.X, y-1, z, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y-1, z) && !startBot.IsFilled(c1.X, y-1, z) {
						if !moved {
							startBot.Goto(builder.Coord{c1.X, y, z})
							moved = true
						}
						startBot.Fill(0, -1, 0)
					}
					targetZ = z + zdir
					if ((zdir > 0 && targetZ <= c2.Z) || (zdir < 0 && targetZ >= c1.Z)) &&
						builder.Filled(c1.X, y-1, targetZ, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y-1, targetZ) && !startBot.IsFilled(c1.X, y-1, targetZ) {
						if !moved {
							startBot.Goto(builder.Coord{c1.X, y, z})
							moved = true
						}
						startBot.Fill(0, -1, zdir)
					}
				}
			} else {
				debugPrintf("Filling along X axis\n")
				xdir := 1
				if c2.X < c1.X {
					xdir = -1
				}
				for x := c1.X + xdir; (xdir > 0 && x < c2.X+3) || (xdir < 0 && x > c1.X-3); x += 3 * xdir {
					moved := false
					targetX := x - xdir
					if ((xdir > 0 && targetX <= c2.X) || (xdir < 0 && targetX >= c1.X)) &&
						builder.Filled(targetX, y-1, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(targetX, y-1, c1.Z) && !startBot.IsFilled(targetX, y-1, c1.Z) {
						if !moved {
							startBot.Goto(builder.Coord{x, y, c1.Z})
							moved = true
						}
						debugPrintf("Filling %d,%d,%d\n", targetX, y-1, c1.Z)
						startBot.Fill(-xdir, -1, 0)
					}
					if ((xdir > 0 && x <= c2.X) || (xdir < 0 && x >= c1.X)) &&
						builder.Filled(x, y-1, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(x, y-1, c1.Z) && !startBot.IsFilled(x, y-1, c1.Z) {
						if !moved {
							startBot.Goto(builder.Coord{x, y, c1.Z})
							moved = true
						}
						debugPrintf("Filling %d,%d,%d\n", x, y-1, c1.Z)
						startBot.Fill(0, -1, 0)
					}
					targetX = x + xdir
					if ((xdir > 0 && targetX <= c2.X) || (xdir < 0 && targetX >= c1.X)) &&
						builder.Filled(targetX, y-1, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(targetX, y-1, c1.Z) && !startBot.IsFilled(targetX, y-1, c1.Z) {
						if !moved {
							startBot.Goto(builder.Coord{x, y, c1.Z})
							moved = true
						}
						debugPrintf("Filling %d,%d,%d\n", targetX, y-1, c1.Z)
						startBot.Fill(xdir, -1, 0)
					}
				}
			}
		}
		zdir = -zdir
		//		startBot.SMove(0, 1, 0)
	}

	botPos := startBot.Pos()

	changed := true
	for changed {
		changed = false

		for y := botPos.Y - 1; y >= 1; y-- {
			for {
				ok, c1, c2 := startBot.FindClosestUnconnectedLine(y)
				if !ok {
					break
				}
				debugPrintf("Found unconnected line from %d,%d,%d to %d,%d,%d\n",
					c1.X, c1.Y, c1.Z, c2.X, c2.Y, c2.Z)

				startBot.GotoShortest(c1.X, y-1, c1.Z)
				if c1.X == c2.X {
					debugPrintf("Filling along Z axis\n")
					zdir := 1
					if c2.Z < c1.Z {
						zdir = -1
					}
					for z := c1.Z + zdir; (zdir > 0 && z < c2.Z+3) || (zdir < 0 && z > c1.Z-3); z += 3 * zdir {
						moved := false
						targetZ := z - zdir
						if ((zdir > 0 && targetZ <= c2.Z) || (zdir < 0 && targetZ >= c1.Z)) &&
							builder.Filled(c1.X, y, targetZ, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y, targetZ) && !startBot.IsFilled(c1.X, y, targetZ) {
							if !moved {
								startBot.Goto(builder.Coord{c1.X, y - 1, z})
								moved = true
							}
							changed = true
							startBot.Fill(0, 1, -zdir)
						}
						if ((zdir > 0 && z <= c2.Z) || (zdir < 0 && z >= c1.Z)) &&
							builder.Filled(c1.X, y, z, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y, z) && !startBot.IsFilled(c1.X, y, z) {
							if !moved {
								startBot.Goto(builder.Coord{c1.X, y - 1, z})
								moved = true
							}
							changed = true
							startBot.Fill(0, 1, 0)
						}
						targetZ = z + zdir
						if ((zdir > 0 && targetZ <= c2.Z) || (zdir < 0 && targetZ >= c1.Z)) &&
							builder.Filled(c1.X, y-1, targetZ, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y, targetZ) && !startBot.IsFilled(c1.X, y, targetZ) {
							if !moved {
								startBot.Goto(builder.Coord{c1.X, y - 1, z})
								moved = true
							}
							changed = true
							startBot.Fill(0, 1, zdir)
						}
					}
				} else {
					debugPrintf("Filling along X axis\n")
					xdir := 1
					if c2.X < c1.X {
						xdir = -1
					}
					for x := c1.X + xdir; (xdir > 0 && x < c2.X+3) || (xdir < 0 && x > c1.X-3); x += 3 * xdir {
						moved := false
						targetX := x - xdir
						if ((xdir > 0 && targetX <= c2.X) || (xdir < 0 && targetX >= c1.X)) &&
							builder.Filled(targetX, y, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(targetX, y, c1.Z) && !startBot.IsFilled(targetX, y, c1.Z) {
							if !moved {
								startBot.Goto(builder.Coord{x, y - 1, c1.Z})
								moved = true
							}
							debugPrintf("Filling %d,%d,%d\n", targetX, y, c1.Z)
							changed = true
							startBot.Fill(-xdir, 1, 0)
						}
						if ((xdir > 0 && x <= c2.X) || (xdir < 0 && x >= c1.X)) &&
							builder.Filled(x, y, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(x, y, c1.Z) && !startBot.IsFilled(x, y, c1.Z) {
							if !moved {
								startBot.Goto(builder.Coord{x, y - 1, c1.Z})
								moved = true
							}
							debugPrintf("Filling %d,%d,%d\n", x, y, c1.Z)
							changed = true
							startBot.Fill(0, 1, 0)
						}
						targetX = x + xdir
						if ((xdir > 0 && targetX <= c2.X) || (xdir < 0 && targetX >= c1.X)) &&
							builder.Filled(targetX, y, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(targetX, y, c1.Z) && !startBot.IsFilled(targetX, y, c1.Z) {
							if !moved {
								startBot.Goto(builder.Coord{x, y - 1, c1.Z})
								moved = true
							}
							debugPrintf("Filling %d,%d,%d\n", targetX, y, c1.Z)
							changed = true
							startBot.Fill(xdir, 1, 0)
						}
					}
				}
			}
		}

		for y := 1; y <= r-1; y++ {
			for {
				ok, c1, c2 := startBot.FindClosestUnconnectedLine(y - 1)
				if !ok {
					break
				}
				debugPrintf("Found unconnected line from %d,%d,%d to %d,%d,%d\n",
					c1.X, c1.Y, c1.Z, c2.X, c2.Y, c2.Z)

				startBot.Goto(builder.Coord{c1.X, y, c1.Z})
				if c1.X == c2.X {
					debugPrintf("Filling along Z axis\n")
					zdir := 1
					if c2.Z < c1.Z {
						zdir = -1
					}
					for z := c1.Z + zdir; (zdir > 0 && z < c2.Z+3) || (zdir < 0 && z > c1.Z-3); z += 3 * zdir {
						moved := false
						targetZ := z - zdir
						if ((zdir > 0 && targetZ <= c2.Z) || (zdir < 0 && targetZ >= c1.Z)) &&
							builder.Filled(c1.X, y-1, targetZ, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y-1, targetZ) && !startBot.IsFilled(c1.X, y-1, targetZ) {
							if !moved {
								startBot.GotoShortest(c1.X, y, z)
								moved = true
							}
							changed = true
							startBot.Fill(0, -1, -zdir)
						}
						if ((zdir > 0 && z <= c2.Z) || (zdir < 0 && z >= c1.Z)) &&
							builder.Filled(c1.X, y-1, z, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y-1, z) && !startBot.IsFilled(c1.X, y-1, z) {
							if !moved {
								startBot.GotoShortest(c1.X, y, z)
								moved = true
							}
							changed = true
							startBot.Fill(0, -1, 0)
						}
						targetZ = z + zdir
						if ((zdir > 0 && targetZ <= c2.Z) || (zdir < 0 && targetZ >= c1.Z)) &&
							builder.Filled(c1.X, y-1, targetZ, modelBytes) && startBot.GetBuilder().IsSupported(c1.X, y-1, targetZ) && !startBot.IsFilled(c1.X, y-1, targetZ) {
							if !moved {
								startBot.GotoShortest(c1.X, y, z)
								moved = true
							}
							changed = true
							startBot.Fill(0, -1, zdir)
						}
					}
				} else {
					debugPrintf("Filling along X axis\n")
					xdir := 1
					if c2.X < c1.X {
						xdir = -1
					}
					for x := c1.X + xdir; (xdir > 0 && x < c2.X+3) || (xdir < 0 && x > c1.X-3); x += 3 * xdir {
						moved := false
						targetX := x - xdir
						if ((xdir > 0 && targetX <= c2.X) || (xdir < 0 && targetX >= c1.X)) &&
							builder.Filled(targetX, y-1, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(targetX, y-1, c1.Z) && !startBot.IsFilled(targetX, y-1, c1.Z) {
							if !moved {
								startBot.GotoShortest(x, y, c1.Z)
								moved = true
							}
							debugPrintf("Filling %d,%d,%d\n", targetX, y-1, c1.Z)
							changed = true
							startBot.Fill(-xdir, -1, 0)
						}
						if ((xdir > 0 && x <= c2.X) || (xdir < 0 && x >= c1.X)) &&
							builder.Filled(x, y-1, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(x, y-1, c1.Z) && !startBot.IsFilled(x, y-1, c1.Z) {
							if !moved {
								startBot.GotoShortest(x, y, c1.Z)
								moved = true
							}
							debugPrintf("Filling %d,%d,%d\n", x, y-1, c1.Z)
							changed = true
							startBot.Fill(0, -1, 0)
						}
						targetX = x + xdir
						if ((xdir > 0 && targetX <= c2.X) || (xdir < 0 && targetX >= c1.X)) &&
							builder.Filled(targetX, y-1, c1.Z, modelBytes) && startBot.GetBuilder().IsSupported(targetX, y-1, c1.Z) && !startBot.IsFilled(targetX, y-1, c1.Z) {
							if !moved {
								startBot.GotoShortest(x, y, c1.Z)
								moved = true
							}
							debugPrintf("Filling %d,%d,%d\n", targetX, y-1, c1.Z)
							changed = true
							startBot.Fill(xdir, -1, 0)
						}
					}
				}
			}
			zdir = -zdir
			//		startBot.SMove(0, 1, 0)
		}
	}

	/*
		botPos := startBot.Pos()
		if botPos.X > 0 {
			startBot.Goto(builder.Coord{0, botPos.Y, botPos.Z})
		}
		if botPos.Z > 0 {
			startBot.Goto(builder.Coord{0, botPos.Y, 0})
		}
		startBot.Goto(builder.Coord{0, 0, 0})
	*/
	startBot.GotoShortest(0, 0, 0)
	startBot.Halt()
}
