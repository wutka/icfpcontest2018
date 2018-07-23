package modelers

import (
	"fmt"
	"wutka.com/icfpc/builder"
)

var debug = false

type Blobs struct {
}

const (
	BLOB_MOVE = 1
	BLOB_FILL = 2
)

const (
	AXIS_X = 1
	AXIS_Z = 2
)

type BlobCommand struct {
	command int
	x, z    int
}

func OtherAxis(axis int) int {
	if axis == AXIS_X {
		return AXIS_Z
	} else {
		return AXIS_X
	}
}
func MoveN(x, z int, n int, axis, dir int) (int, int) {
	if axis == AXIS_X {
		x = x + n*dir
	} else {
		z = z + n*dir
	}
	return x, z
}

func CanFill(blob builder.Slice, b *builder.Builder, x, z int) bool {
	debugPrintf("Checking CanFill at %d,%d\n", x, z)
	if blob.IsFilled(x-1, z) {
		return true
	}
	debugPrintf("CanFill %d,%d not filled\n", x-1, z)
	if blob.IsFilled(x+1, z) {
		return true
	}
	debugPrintf("CanFill %d,%d not filled\n", x+1, z)
	if blob.IsFilled(x, z-1) {
		return true
	}
	debugPrintf("CanFill %d,%d not filled\n", x, z-1)
	if blob.IsFilled(x, z+1) {
		return true
	}
	debugPrintf("CanFill %d,%d not filled\n", x, z+1)
	return b.IsSupported(x, blob.Y, z)
}

func RowSplit(blob builder.Slice, x, z, axis int) (int, int) {
	negLength := 0
	if axis == AXIS_X {
		for currZ := z - 1; currZ >= 0; currZ-- {
			n, p := LengthSplit(blob, x, currZ, axis)
			if n == 0 && p == 0 {
				break
			}
			negLength++
		}
	} else {
		for currX := x - 1; currX >= 0; currX-- {
			n, p := LengthSplit(blob, currX, z, axis)
			if n == 0 && p == 0 {
				break
			}
			negLength++
		}
	}
	posLength := 0
	if axis == AXIS_X {
		for currZ := z; currZ < blob.R; currZ++ {
			n, p := LengthSplit(blob, x, currZ, axis)
			if n == 0 && p == 0 {
				break
			}
			posLength++
		}
	} else {
		for currX := x; currX < blob.R; currX++ {
			n, p := LengthSplit(blob, currX, z, axis)
			if n == 0 && p == 0 {
				break
			}
			posLength++
		}
	}

	return negLength, posLength
}

func LengthSplit(blob builder.Slice, x, z, axis int) (int, int) {
	negLength := 0

	var lastFilled int
	if axis == AXIS_X {
		lastFilled = x
	} else {
		lastFilled = z
	}

	currX, currZ := x, z
	for currX > 0 && currZ > 0 {
		newX, newZ := MoveN(currX, currZ, 1, axis, -1)
		if newX == x && newZ == z {
			break
		}
		if blob.IsFilled(newX, newZ) {
			if axis == AXIS_X {
				lastFilled = newX
			} else {
				lastFilled = newZ
			}
		}
		currX, currZ = newX, newZ
	}

	if axis == AXIS_X {
		negLength = x - lastFilled
	} else {
		negLength = z - lastFilled
	}

	posLength := 0
	if axis == AXIS_X {
		lastFilled = x
	} else {
		lastFilled = z
	}

	currX, currZ = x, z
	for currX < blob.R-1 && currZ < blob.R-1 {
		newX, newZ := MoveN(currX, currZ, 1, axis, 1)
		if newX == x && newZ == z {
			break
		}
		if blob.IsFilled(newX, newZ) {
			if axis == AXIS_X {
				lastFilled = newX
			} else {
				lastFilled = newZ
			}
		}
		currX, currZ = newX, newZ
	}

	if axis == AXIS_X {
		posLength = lastFilled - x
	} else {
		posLength = lastFilled - z
	}

	return negLength, posLength
}

func FillsNeeded(blob, filler builder.Slice, x, z, axis int, dir int) bool {
	if blob.IsFilled(x, z) && !filler.IsFilled(x, z) {
		return true
	}

	currX, currZ := x, z
	for currX < blob.R-1 && currZ > blob.R-1 {
		newX, newZ := MoveN(currX, currZ, 1, axis, dir)
		if newX == x && newZ == z {
			return false
		}
		if blob.IsFilled(newX, newZ) && !filler.IsFilled(newX, newZ) {
			return true
		}
		currX, currZ = newX, newZ
	}
	return false
}

func ScoreFillPlan(plan []BlobCommand, fromX, fromZ int) int {
	sum := 0
	for _, command := range plan {
		if command.command == BLOB_MOVE {
			sum += builder.Iabs(fromX - command.x)
			sum += builder.Iabs(fromZ - command.z)
		}
	}
	return sum
}

func MakeFillPlan(blob builder.Slice, b *builder.Builder, startX, startZ int, axis int, bestScore int, botX, botZ int) (bool, []BlobCommand) {
	currScore := 0
	currBotX := botX
	currBotZ := botZ

	debugPrintf("Making a fill plan starting at %d,%d\n", startX, startZ)
	filler := blob.Empty()
	plan := []BlobCommand{}

	if axis == AXIS_X {
		debugPrintf("Trying fills along X axis\n")
	} else {
		debugPrintf("Trying fills along Z axis\n")
	}

	if OtherAxis(axis) == AXIS_X {
		debugPrintf("Other axis is X axis\n")
	} else {
		debugPrintf("Other axis is Z axis\n")
	}

	negRows, posRows := RowSplit(blob, startX, startZ, axis)
	debugPrintf("Row Split at %d,%d = %d,%d\n", startX, startZ, negRows, posRows)

	var rowProcessing [][]int
	if negRows < posRows {
		if negRows > 0 {
			rowProcessing = [][]int{{0, negRows + 1, -1}, {0, posRows + 1, 1}}
		} else {
			rowProcessing = [][]int{{0, negRows, -1}, {0, posRows + 1, 1}}
		}
	} else {
		if posRows > 0 {
			rowProcessing = [][]int{{0, posRows + 1, 1}, {0, negRows + 1, -1}}
		} else {
			rowProcessing = [][]int{{0, posRows, 1}, {0, negRows + 1, -1}}
		}
	}

	lastX, lastZ := startX, startZ

	for _, rp := range rowProcessing {

		rowStart := rp[0]
		rowMax := rp[1]
		rowDir := rp[2]

		debugPrintf("rowMax=%d rowDir = %d\n", rowMax, rowDir)
		// off is offset from either start startZ or startX
		for off := rowStart; off < rowMax; off++ {
			debugPrintf("Offset = %d\n", off)
			var x, z int
			if axis == AXIS_X {
				x, z = lastX, startZ+off*rowDir
			} else {
				x, z = startX+off*rowDir, lastZ
			}

			debugPrintf("Starting row at %d,%d\n", x, z)
			rowFilled := true
			for i := 0; i < blob.R; i++ {
				if axis == AXIS_X {
					if blob.IsFilled(i, z) && !filler.IsFilled(i, z) {
						rowFilled = false
						break
					}
				} else {
					if blob.IsFilled(x, i) && !filler.IsFilled(x, i) {
						rowFilled = false
						break
					}
				}
			}
			if rowFilled {
				debugPrintf("Skipping filled row\n")
				continue
			}

			// We can move up one more row if the next row is not shorter than the current at either end
			if off < rowMax-1 {
				currNeg, currPos := LengthSplit(blob, x, z, axis)
				nextX, nextZ := MoveN(x, z, 1, OtherAxis(axis), rowDir)
				nextNeg, nextPos := LengthSplit(blob, nextX, nextZ, axis)
				if nextNeg >= currNeg && nextPos >= currPos {
					if blob.IsFilled(x, z) && !filler.IsFilled(x, z) && ((x == startX && z == startZ) || CanFill(filler, b, x, z)) {
						debugPrintf("Filling at %d,%d for row skip\n", x, z)
						plan = append(plan, BlobCommand{BLOB_MOVE, x, z})
						currScore += builder.Iabs(currBotX - x)
						currScore += builder.Iabs(currBotZ - z)
						if bestScore >= 0 && currScore > bestScore {
							//							return false, plan
						}
						currBotX = x
						currBotZ = z
						filler.Fill(x, z)
						plan = append(plan, BlobCommand{BLOB_FILL, x, z})
					}
					off++
					if axis == AXIS_X {
						x, z = lastX, startZ+off*rowDir
					} else {
						x, z = startX+off*rowDir, lastZ
					}
				}
			}

			// Go ahead and fill the row

			negLength, posLength := LengthSplit(blob, x, z, axis)
			if negLength == 0 && posLength == 0 && filler.IsFilled(x, z) {
				debugPrintf("Nothing to fill on the row\n")
				continue
			}

			offsets := []int{}
			dirs := []int{}

			if negLength < posLength {
				for i := 0; i <= negLength; i++ {
					offsets = append(offsets, -i)
					dirs = append(dirs, -1)
				}
				for i := 0; i < posLength+1; i++ {
					offsets = append(offsets, i)
					dirs = append(dirs, 1)
				}
			} else {
				for i := 0; i <= posLength; i++ {
					offsets = append(offsets, i)
					dirs = append(dirs, 1)
				}
				for i := 0; i < negLength+1; i++ {
					offsets = append(offsets, -i)
					dirs = append(dirs, -1)
				}
			}

			debugPrintf("Starting point loop from %d,%d\n", x, z)
			for canBackup := 0; canBackup < 2; canBackup++ {

				for _, pointOffset := range offsets {
					debugPrintf("Point offset = %d\n", pointOffset)
					//				moveDir := dirs[poi]

					var currX, currZ int

					if axis == AXIS_X {
						currX, currZ = x+pointOffset, z
						lastX = currX
					} else {
						currX, currZ = x, z+pointOffset
						lastZ = currZ
					}

					debugPrintf("CurrX,currZ = %d,%d\n", currX, currZ)
					if blob.IsFilled(currX, currZ) {
						if !filler.IsFilled(currX, currZ) {
							debugPrintf("%d,%d still needs filling\n", currX, currZ)
						} else {
							debugPrintf("%d,%d is filled\n", currX, currZ)
						}
					} else {
						debugPrintf("%d,%d doesn't need filling\n", currX, currZ)
					}

					//				for FillsNeeded(blob, filler, currX, currZ, OtherAxis(axis), moveDir) {
					//					filled := false
					moved := false
					if blob.IsFilled(currX, currZ) && !filler.IsFilled(currX, currZ) && CanFill(filler, b, currX, currZ) {
						debugPrintf("Filling at %d,%d\n", currX, currZ)
						if !moved {
							plan = append(plan, BlobCommand{BLOB_MOVE, currX, currZ})
							currScore += builder.Iabs(currBotX - x)
							currScore += builder.Iabs(currBotZ - z)
							if bestScore >= 0 && currScore > bestScore {
								//								return false, plan
							}
							currBotX = x
							currBotZ = z
							moved = true
						}
						//						filled = true
						filler.Fill(currX, currZ)
						plan = append(plan, BlobCommand{BLOB_FILL, currX, currZ})
					}
					otherX, otherZ := MoveN(currX, currZ, 1, OtherAxis(axis), -1)
					if blob.IsFilled(otherX, otherZ) && !filler.IsFilled(otherX, otherZ) && CanFill(filler, b, otherX, otherZ) {
						debugPrintf("Filling at %d,%d\n", otherX, otherZ)
						if !moved {
							plan = append(plan, BlobCommand{BLOB_MOVE, currX, currZ})
							currScore += builder.Iabs(currBotX - x)
							currScore += builder.Iabs(currBotZ - z)
							if bestScore >= 0 && currScore > bestScore {
								//								return false, plan
							}
							currBotX = x
							currBotZ = z
							moved = true
						}
						//						filled = true
						filler.Fill(otherX, otherZ)
						plan = append(plan, BlobCommand{BLOB_FILL, otherX, otherZ})
					}

					if blob.IsFilled(currX, currZ) && !filler.IsFilled(currX, currZ) && CanFill(filler, b, currX, currZ) {
						debugPrintf("Filling at %d,%d\n", currX, currZ)
						if !moved {
							plan = append(plan, BlobCommand{BLOB_MOVE, currX, currZ})
							currScore += builder.Iabs(currBotX - x)
							currScore += builder.Iabs(currBotZ - z)
							if bestScore >= 0 && currScore > bestScore {
								//								return false, plan
							}
							currBotX = x
							currBotZ = z
							moved = true
						}
						//						filled = true
						filler.Fill(currX, currZ)
						plan = append(plan, BlobCommand{BLOB_FILL, currX, currZ})
					}

					otherX, otherZ = MoveN(currX, currZ, 1, OtherAxis(axis), 1)
					if blob.IsFilled(otherX, otherZ) && !filler.IsFilled(otherX, otherZ) && CanFill(filler, b, otherX, otherZ) {
						debugPrintf("Filling at %d,%d\n", otherX, otherZ)
						if !moved {
							plan = append(plan, BlobCommand{BLOB_MOVE, currX, currZ})
							currScore += builder.Iabs(currBotX - x)
							currScore += builder.Iabs(currBotZ - z)
							if bestScore >= 0 && currScore > bestScore {
								//								return false, plan
							}
							currBotX = x
							currBotZ = z
							moved = true
						}
						//						filled = true
						filler.Fill(otherX, otherZ)
						plan = append(plan, BlobCommand{BLOB_FILL, otherX, otherZ})
					}
					if blob.IsFilled(currX, currZ) && !filler.IsFilled(currX, currZ) && CanFill(filler, b, currX, currZ) {
						debugPrintf("Filling at %d,%d\n", currX, currZ)
						if !moved {
							plan = append(plan, BlobCommand{BLOB_MOVE, currX, currZ})
							currScore += builder.Iabs(currBotX - x)
							currScore += builder.Iabs(currBotZ - z)
							if bestScore >= 0 && currScore > bestScore {
								//								return false, plan
							}
							currBotX = x
							currBotZ = z
							moved = true
						}
						//						filled = true
						filler.Fill(currX, currZ)
						plan = append(plan, BlobCommand{BLOB_FILL, currX, currZ})
					}

					otherX, otherZ = MoveN(currX, currZ, 1, OtherAxis(axis), -1)
					if blob.IsFilled(otherX, otherZ) && !filler.IsFilled(otherX, otherZ) && CanFill(filler, b, otherX, otherZ) {
						debugPrintf("Filling at %d,%d\n", otherX, otherZ)
						if !moved {
							plan = append(plan, BlobCommand{BLOB_MOVE, currX, currZ})
							currScore += builder.Iabs(currBotX - x)
							currScore += builder.Iabs(currBotZ - z)
							if bestScore >= 0 && currScore > bestScore {
								//								return false, plan
							}
							currBotX = x
							currBotZ = z
							moved = true
						}
						//						filled = true
						filler.Fill(otherX, otherZ)
						plan = append(plan, BlobCommand{BLOB_FILL, otherX, otherZ})
					}

					//					if !filled {
					//						debugPrintf("Needed to fill at %d,%d but couldn't\n", currX, currZ)
					//						return false, plan
					//					}
					//				}
				}

				rowFilled = true
				for i := 0; i < blob.R; i++ {
					if axis == AXIS_X {
						if blob.IsFilled(i, z) && !filler.IsFilled(i, z) {
							rowFilled = false
							break
						}
					} else {
						if blob.IsFilled(x, i) && !filler.IsFilled(x, i) {
							rowFilled = false
							break
						}
					}
				}
				if rowFilled {
					break
				}
				newOffsets := []int{}
				newDirs := []int{}
				for i := len(offsets) - 1; i >= 0; i-- {
					newOffsets = append(newOffsets, offsets[i])
					newDirs = append(newDirs, dirs[i])
				}
				offsets = newOffsets
				dirs = newDirs
			}
		}
	}

	if !blob.Equal(filler) {
		for x := 0; x < blob.R; x++ {
			for z := 0; z < blob.R; z++ {
				if blob.IsFilled(x, z) && !filler.IsFilled(x, z) {
					debugPrintf("Blob/filler mismatch at %d,%d,  filler not filled\n", x, z)
				} else if !blob.IsFilled(x, z) && filler.IsFilled(x, z) {
					debugPrintf("Blob/filler mismatch at %d,%d,  filler was filled but shouldn't be\n", x, z)
				}
			}
		}
		//panic("Fill routine didn't fill the whole blob")
		return false, plan
	}
	return true, plan
}

func ExecuteFillPlan(bot *builder.Bot, y int, botLocation int, plan []BlobCommand, useSafeGoto bool) {
	first := true
	for _, command := range plan {
		if command.command == BLOB_MOVE {
			if useSafeGoto || !first {
				bot.GotoShortest(command.x, y+botLocation, command.z)
			} else {
				//bot.GotoShortest(command.x, y+botLocation, command.z)
				bot.Goto(builder.Coord{command.x, y + botLocation, command.z})
				first = false
			}
		} else if command.command == BLOB_FILL {
			botPos := bot.Pos()
			botX, botZ := botPos.X, botPos.Z
			bot.Fill(command.x-botX, -botLocation, command.z-botZ)
		}
	}
	return
}

func imax(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func getBestDistances(coords []builder.Coord, base builder.Coord, n int) []builder.Coord {
	numCoords := 0
	bestDistances := []int{}
	bestCoords := []builder.Coord{}

	for _, coord := range coords {
		dist := builder.MDist(base, coord)
		if numCoords < n {
			bestDistances = append(bestDistances, dist)
			bestCoords = append(bestCoords, coord)
		} else {
			highestPos := 0
			highestDist := bestDistances[0]

			for i, bd := range bestDistances[1:] {
				if bd > highestDist {
					highestPos = i
					highestDist = bd
				}
			}
			if dist < highestDist {
				bestDistances[highestPos] = dist
				bestCoords[highestPos] = coord
			}
		}
	}

	return bestCoords
}
func (b_ *Blobs) Model(modelBytes []byte, startBot *builder.Bot) {
	startBot.Start()

	r := int(modelBytes[0])

	b := startBot.GetBuilder()

	first := true

	changed := true

	for changed {
		changed = false
		for y := 0; y < r-1; y++ {
			debugPrintf("at y=%d\n", y)
			slice := b.GetUnfilledSlice(y)
			blobList := slice.ToBlobList()

			debugPrintf("blobList has %d items\n", len(blobList))
			for len(blobList) > 0 {
				foundSupported := false
				var currBlob builder.Slice
				currBlobPos := -1

				for blobPos, blob := range blobList {
					if b.IsBlobSupported(blob) {
						currBlob = blob
						foundSupported = true
						currBlobPos = blobPos
						break
					}
				}
				if !foundSupported {
					debugPrintf("Blob isn't supported yet\n")
					break
				}

				supportPoints := b.GetBlobSupportPoints(currBlob)
				closestPoints := getBestDistances(supportPoints, startBot.Pos(), 100)

				gotBest := false
				bestScore := -1
				var bestPlan []BlobCommand
				for _, axis := range []int{AXIS_X, AXIS_Z} {
					for _, pt := range closestPoints {
						planOK, plan := MakeFillPlan(currBlob, startBot.GetBuilder(), pt.X, pt.Z, axis, bestScore, startBot.Pos().X, startBot.Pos().Z)
						if !planOK {
							continue
						}
						planScore := ScoreFillPlan(plan, startBot.Pos().X, startBot.Pos().Z)
						if !gotBest || (planScore < bestScore) {
							gotBest = true
							bestScore = planScore
							bestPlan = plan
						}
					}
				}
				if gotBest {
					debugPrintf("Executing plan\n")
					ExecuteFillPlan(startBot, y, 1, bestPlan, !first)
					first = false
					changed = true
				} else {
					debugPrintf("No plan at %d\n", y)
				}

				blobList = append(blobList[:currBlobPos], blobList[currBlobPos+1:]...)
			}
		}

		modelComplete, _, _ := b.CheckAgainstModel()
		if modelComplete {
			break
		}

		debugPrintf("Reversing\n")

		for y := r - 1; y >= 0; y-- {
			debugPrintf("at y=%d\n", y)
			slice := b.GetUnfilledSlice(y)
			blobList := slice.ToBlobList()

			debugPrintf("blobList has %d items\n", len(blobList))
			for len(blobList) > 0 {
				foundSupported := false
				var currBlob builder.Slice
				currBlobPos := -1

				for blobPos, blob := range blobList {
					if b.IsBlobSupported(blob) {
						currBlob = blob
						foundSupported = true
						currBlobPos = blobPos
						break
					}
				}
				if !foundSupported {
					debugPrintf("Blob isn't supported yet\n")
					break
				}

				supportPoints := b.GetBlobSupportPoints(currBlob)
				closestPoints := getBestDistances(supportPoints, startBot.Pos(), 100)

				gotBest := false
				bestScore := 0
				var bestPlan []BlobCommand
				for _, axis := range []int{AXIS_X, AXIS_Z} {
					for _, pt := range closestPoints {
						planOK, plan := MakeFillPlan(currBlob, startBot.GetBuilder(), pt.X, pt.Z, axis, bestScore, startBot.Pos().X, startBot.Pos().Z)
						if !planOK {
							continue
						}
						planScore := ScoreFillPlan(plan, startBot.Pos().X, startBot.Pos().Z)
						if !gotBest || (planScore < bestScore) {
							gotBest = true
							bestScore = planScore
							bestPlan = plan
						}
					}
				}

				if gotBest {
					debugPrintf("Executing plan\n")
					ExecuteFillPlan(startBot, y, -1, bestPlan, !first)
					first = false
					changed = true
				}

				blobList = append(blobList[:currBlobPos], blobList[currBlobPos+1:]...)

				modelComplete, _, _ := b.CheckAgainstModel()
				if modelComplete {
					break
				}
			}
		}
	}

	startBot.GotoShortest(0, 0, 0)
	startBot.Halt()
}

func debugPrintf(f string, args ...interface{}) {
	if debug {
		fmt.Printf(f, args...)
	}
}
