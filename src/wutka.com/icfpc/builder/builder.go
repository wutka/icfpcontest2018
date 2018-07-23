package builder

import (
	"fmt"
	"sort"
	"sync"
)

type Coord struct {
	X, Y, Z int
}

type Region struct {
	c1, c2 Coord
}

type BotMotionSync struct {
	numBots int
	ready   bool
}

type Bot struct {
	id            int
	pos           Coord
	seeds         []int
	volatile      Region
	builder       *Builder
	pendingFusion bool
	fusionTarget  int
	motionSync    *BotMotionSync
	destination   Coord
	traceChannel  chan []byte
	orderChannel  chan interface{}
}

type GroupOp struct {
	opType      int
	numBotsLeft int
	region      Region
}

type Builder struct {
	r               int
	grid            []byte
	supported       []bool
	model           []byte
	pendingGroupOps []*GroupOp
	harmonics       int
	energy          int
	motionSyncs     []*BotMotionSync

	bots []*Bot
}

type Slice struct {
	R     int
	Y     int
	slice []byte
}

func NewFromModel(modelBytes []byte) *Bot {
	builder := Builder{r: int(modelBytes[0]), model: modelBytes}
	builder.grid = make([]byte, builder.r*builder.r*builder.r)
	builder.supported = make([]bool, builder.r*builder.r*builder.r)
	builder.harmonics = 0
	builder.pendingGroupOps = []*GroupOp{}

	newBot := Bot{id: 1, pos: Coord{}, volatile: Region{}, builder: &builder, traceChannel: make(chan []byte),
		seeds: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}}
	builder.bots = []*Bot{&newBot}

	return &newBot
}

func (b *Builder) GetTrace() []byte {
	botList := make([]*Bot, 20)
	dummy := []byte{}

	trace := []byte{}
	for {
		for i := 0; i < 20; i++ {
			botList[i] = nil
		}
		for _, bot := range b.bots {
			botList[bot.id-1] = bot
		}
		for i := 0; i < 20; i++ {
			if botList[i] != nil {
				botList[i].traceChannel <- dummy
				botTrace := <-botList[i].traceChannel
				trace = append(trace, botTrace...)
				if len(botTrace) == 1 && botTrace[0] == 0xff {
					return trace
				}
			}
		}
	}
}

func (b *Builder) GetTraceAsList() [][]byte {
	botList := make([]*Bot, 20)
	dummy := []byte{}

	trace := [][]byte{}
	for {
		for i := 0; i < 20; i++ {
			botList[i] = nil
		}
		for _, bot := range b.bots {
			botList[bot.id-1] = bot
		}
		for i := 0; i < 20; i++ {
			if botList[i] != nil {
				botList[i].traceChannel <- dummy
				botTrace := <-botList[i].traceChannel
				trace = append(trace, botTrace)
				if len(botTrace) == 1 && botTrace[0] == 0xff {
					return trace
				}
			}
		}
	}
}

func Iabs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func (b *Bot) Start() {
	_ = <-b.traceChannel
}

func (b *Bot) checkFusion() {
	if b.pendingFusion {
		panic(fmt.Sprintf("Bot %d was supposed to fuse with %d", b.id, b.fusionTarget))
		return
	}
}

func (b *Bot) GetBuilder() *Builder {
	return b.builder
}

func (b *Bot) GetOrder() interface{} {
	return <-b.orderChannel
}

func (b *Builder) CheckAgainstModel() (bool, Coord, bool) {
	for x := 1; x < b.r-1; x++ {
		for y := 1; y < b.r-1; y++ {
			for z := 1; z < b.r-1; z++ {
				gridFilled := b.grid[b.GridOffset(Coord{x, y, z})] != 0
				if gridFilled != Filled(x, y, z, b.model) {
					return false, Coord{x, y, z}, gridFilled
				}
			}
		}

	}
	return true, Coord{}, false
}

func (b *Builder) GetSlice(y int) Slice {
	slice := make([]byte, b.r*b.r)
	for x := 0; x < b.r; x++ {
		for z := 0; z < b.r; z++ {
			if Filled(x, y, z, b.model) {
				slice[x*b.r+z] = 1
			}
		}
	}
	return Slice{R: b.r, Y: y, slice: slice}
}

func (b *Builder) GetUnfilledSlice(y int) Slice {
	slice := make([]byte, b.r*b.r)
	for x := 0; x < b.r; x++ {
		for z := 0; z < b.r; z++ {
			if Filled(x, y, z, b.model) && !b.IsFilled(x, y, z) {
				slice[x*b.r+z] = 1
			}
		}
	}
	return Slice{R: b.r, Y: y, slice: slice}
}

func (s *Slice) IsFilled(x, z int) bool {
	if x < 0 || z < 0 || x >= s.R || z >= s.R {
		return false
	}
	return s.slice[x*s.R+z] != 0
}

func (s *Slice) Copy() Slice {
	sliceCopy := make([]byte, len(s.slice))
	copy(sliceCopy, s.slice)
	return Slice{R: s.R, Y: s.Y, slice: sliceCopy}
}

func (s *Slice) Empty() Slice {
	sliceCopy := make([]byte, len(s.slice))
	return Slice{R: s.R, Y: s.Y, slice: sliceCopy}
}

func (s *Slice) Equal(other Slice) bool {
	for i := range s.slice {
		if s.slice[i] != other.slice[i] {
			return false
		}
	}
	return true
}

func (s *Slice) Fill(x, z int) {
	s.slice[x*s.R+z] = 1
}

func (s *Slice) Clear(x, z int) {
	s.slice[x*s.R+z] = 0
}

func (s *Slice) ToBlobList() []Slice {
	blobList := []Slice{}
	sliceCopy := s.Copy()

	for {
		found := false
		blobX := -1
		blobZ := -1
		for x := 0; x < s.R; x++ {
			for z := 0; z < s.R; z++ {
				if sliceCopy.IsFilled(x, z) {
					found = true
					blobX = x
					blobZ = z
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return blobList
		}

		newBlob := Slice{R: s.R, Y: s.Y, slice: make([]byte, len(s.slice))}
		workQueue := [][]int{{blobX, blobZ}}
		for len(workQueue) > 0 {
			currItem := workQueue[0]
			workQueue = workQueue[1:]
			x := currItem[0]
			z := currItem[1]
			if !sliceCopy.IsFilled(x, z) {
				continue
			}
			newBlob.Fill(x, z)
			sliceCopy.Clear(x, z)
			if sliceCopy.IsFilled(x, z) {
				panic("Cleared spot still shows filled")
			}

			if sliceCopy.IsFilled(x-1, z) {
				workQueue = append(workQueue, []int{x - 1, z})
			}
			if sliceCopy.IsFilled(x, z-1) {
				workQueue = append(workQueue, []int{x, z - 1})
			}
			if sliceCopy.IsFilled(x+1, z) {
				workQueue = append(workQueue, []int{x + 1, z})
			}
			if sliceCopy.IsFilled(x, z+1) {
				workQueue = append(workQueue, []int{x, z + 1})
			}
		}
		blobList = append(blobList, newBlob)
	}
}

func (b *Builder) IsBlobSupported(blob Slice) bool {
	for x := 0; x < b.r; x++ {
		for z := 0; z < b.r; z++ {
			if b.IsSupported(x, blob.Y, z) {
				return true
			}
		}
	}
	return false
}

func (b *Builder) IsBlobSupportedBy(blob Slice, supportBlob Slice) bool {
	for x := 0; x < b.r; x++ {
		for z := 0; z < b.r; z++ {
			if !supportBlob.IsFilled(x, z) {
				continue
			}
			if b.IsSupported(x, blob.Y, z) {
				return true
			}
		}
	}
	return false
}

func (b *Builder) GetBlobSupportPoints(blob Slice) []Coord {
	coords := []Coord{}
	for x := 0; x < b.r; x++ {
		for z := 0; z < b.r; z++ {
			if blob.IsFilled(x, z) && b.IsSupported(x, blob.Y, z) {
				coords = append(coords, Coord{x, blob.Y, z})
			}
		}
	}
	return coords
}

func ShortestMDist(from Coord, to []Coord) int {
	m := MDist(from, to[0])
	for _, c2 := range to[1:] {
		m2 := MDist(from, c2)
		if m2 < m {
			m = m2
		}
	}
	return m
}

func (b *Bot) FillBlob(blob Slice, startingPoint Coord, botPos int) {

}

func (b *Bot) FindClosestUnconnectedLine(y int) (bool, Coord, Coord) {
	bestRun := 0
	bestDist := 0
	bestC1 := Coord{}
	bestC2 := Coord{}

	//	fmt.Printf("Checking for unconnected lines\n")
	for x := 1; x < b.builder.r-1; x++ {
		for z := 1; z < b.builder.r-1; z++ {
			//			if Filled(x, y, z, b.builder.model) && !b.IsFilled(x, y, z) && !b.IsSupported(x, y, z) {
			//				fmt.Printf("Pixel at %d,%d,%d isn't supported\n", x, y, z)
			//			}
			if Filled(x, y, z, b.builder.model) && !b.IsFilled(x, y, z) && b.builder.IsSupported(x, y, z) {
				lastZ := z
				for lastZ+1 < b.builder.r-1 && Filled(x, y, lastZ+1, b.builder.model) && !b.IsFilled(x, y, lastZ+1) {
					lastZ++
				}
				zrun := lastZ - z + 1
				zc1 := Coord{x, y, z}
				zc2 := Coord{x, y, lastZ}
				if zrun > bestRun || (zrun == bestRun && MDist(b.pos, zc1) < bestDist) {
					bestRun = zrun
					bestC1, bestC2 = zc1, zc2
				}

				lastX := x
				for lastX+1 < b.builder.r-1 && Filled(lastX+1, y, z, b.builder.model) && !b.IsFilled(lastX+1, y, z) {
					lastX++
				}
				xrun := lastX - x + 1
				xc1 := Coord{x, y, z}
				xc2 := Coord{lastX, y, z}
				if xrun > bestRun || (xrun == bestRun && (MDist(b.pos, xc1) < bestDist)) {
					bestRun = xrun
					bestC1, bestC2 = xc1, xc2
				}

				lastZR := z
				for lastZR > 0 && Filled(x, y, lastZR-1, b.builder.model) && !b.IsFilled(x, y, lastZR-1) {
					lastZR--
				}
				zrunR := z - lastZR + 1
				zc1 = Coord{x, y, z}
				zc2 = Coord{x, y, lastZR}
				if zrunR > bestRun || (zrunR == bestRun && (MDist(b.pos, zc1) < bestDist)) {
					bestRun = zrunR
					bestC1, bestC2 = zc1, zc2
				}

				lastXR := x
				for lastXR > 0 && Filled(lastXR-1, y, z, b.builder.model) && !b.IsFilled(lastXR-1, y, z) {
					lastXR--
				}
				xrunR := x - lastXR + 1
				xc1 = Coord{x, y, z}
				xc2 = Coord{lastXR, y, z}
				if xrunR > bestRun || (xrunR == bestRun && (MDist(b.pos, xc1) < bestDist)) {
					bestRun = xrunR
					bestC1, bestC2 = xc1, xc2
				}
			}
		}
	}
	return bestRun > 0, bestC1, bestC2
}

func MDist(c1, c2 Coord) int {
	return Iabs(c1.X-c2.X) + Iabs(c1.Y-c2.Y) + Iabs(c1.Z-c2.Z)
}

func (b *Bot) GridOffset(c Coord) int {
	return c.X*b.builder.r*b.builder.r + c.Y*b.builder.r + c.Z
}

func (b *Builder) GridOffset(c Coord) int {
	return c.X*b.r*b.r + c.Y*b.r + c.Z
}

func (b *Bot) Pos() Coord {
	return b.pos
}

var Origin = Coord{0, 0, 0}

func (b *Bot) Halt() {
	fmt.Printf("Halt\n")
	if len(b.builder.bots) != 1 {
		panic("Tried to halt with more than one active bot")
		return
	}
	if b.pos != Origin {
		panic("Tried to halt with remaining bot not at 0,0,0")
		return
	}
	if b.builder.harmonics != 0 {
		panic("Tried to halt while harmonics high")
		return
	}
	b.traceChannel <- []byte{0xff}
	_ = <-b.traceChannel
}

func (b *Bot) Wait() {
	//	fmt.Printf("Wait\n")
	b.checkFusion()
	b.volatile = Region{b.pos, b.pos}
	b.traceChannel <- []byte{0xfe}
	_ = <-b.traceChannel
}
func (b *Bot) Flip() {
	//	fmt.Printf("Flip\n")
	b.checkFusion()
	b.builder.harmonics = 1 - b.builder.harmonics
	b.volatile = Region{b.pos, b.pos}
	b.traceChannel <- []byte{0xfd}
	_ = <-b.traceChannel
}

func (b *BotMotionSync) Arrived() {
	b.numBots--
}

func (b *BotMotionSync) Check() {
	if b.numBots == 0 {
		b.ready = true
	} else if b.numBots < 0 {
		panic("BotMotionSync got too many arrivals")
		return
	}
}

func (b *BotMotionSync) IsReady() bool {
	return b.ready
}

func (b *Bot) SMove(dx, dy, dz int) {
	//	fmt.Printf("SMove <%d,%d,%d>\n", dx, dy, dz)
	b.checkFusion()
	a, d := encodeLongCoordDiff(dx, dy, dz)
	b1 := byte((a << 4) + 0x04)
	b2 := byte(d)

	oldPos := b.pos
	newX := b.pos.X + dx
	newY := b.pos.Y + dy
	newZ := b.pos.Z + dz
	b.volatile = Region{oldPos, Coord{newX, newY, newZ}}

	if b.motionSync != nil {
		if b.pos == b.destination {
			b.motionSync.Arrived()
		}
	}
	b.CheckSpaceViolations()
	b.traceChannel <- []byte{b1, b2}
	_ = <-b.traceChannel
	b.pos = Coord{newX, newY, newZ}
}

func (b *Bot) LMove(dx1, dy1, dz1, dx2, dy2, dz2 int) {
	//	fmt.Printf("LMove <%d,%d,%d> <%d,%d,%d>\n", dx1, dy1, dz1, dx2, dy2, dz2)
	b.checkFusion()
	a1, d1 := encodeShortCoordDiff(dx1, dy1, dz1)
	a2, d2 := encodeShortCoordDiff(dx2, dy2, dz2)
	//	fmt.Printf("a1,d1 = %d,%d   a2,d2 = %d,%d\n", a1, d1, a2, d2)
	b1 := byte((a2 << 6) + (a1 << 4) + 0x0c)
	b2 := byte((d2 << 4) + d1)

	oldPos := b.pos
	newX := b.pos.X + dx1 + dx2
	newY := b.pos.Y + dy1 + dy2
	newZ := b.pos.Z + dz1 + dz2
	b.volatile = Region{oldPos, Coord{newX, newY, newZ}}

	if b.motionSync != nil {
		if b.pos == b.destination {
			b.motionSync.Arrived()
		}
	}

	b.CheckSpaceViolations()
	b.traceChannel <- []byte{b1, b2}
	_ = <-b.traceChannel
	b.pos = Coord{newX, newY, newZ}
}

func (b *Bot) doFusion(otherPos Coord, isPrimary bool) {
	otherOffset := -1
	for i, otherBot := range b.builder.bots {
		if otherBot.pos == otherPos {
			otherOffset = i
			break
		}
	}
	myOffset := -1
	for i := range b.builder.bots {
		if b.id == i {
			myOffset = i
			break
		}
	}
	if otherOffset < 0 {
		panic("Tried to merge with non-existent bot")
		return
	}
	otherBot := b.builder.bots[otherOffset]
	if b.pendingFusion && b.fusionTarget == otherBot.id {
		b.pendingFusion = false
		b.fusionTarget = 0
		if isPrimary {
			b.seeds = append(b.seeds, otherBot.seeds...)
			sort.Ints(b.seeds)
			b.builder.bots = append(b.builder.bots[0:otherOffset], b.builder.bots[otherOffset+1:]...)
		} else {
			otherBot.seeds = append(otherBot.seeds, b.seeds...)
			sort.Ints(otherBot.seeds)
			b.builder.bots = append(b.builder.bots[0:myOffset], b.builder.bots[myOffset+1:]...)

		}
	} else {
		otherBot.pendingFusion = true
		otherBot.fusionTarget = b.id
	}
}

func (b *Bot) FusionP(dx, dy, dz int) {
	//	fmt.Printf("FusionP <%d,%d,%d>\n", dx, dy, dz)
	nd := byte(encodeNearCoordDiff(dx, dy, dz))
	b1 := (nd << 3) + 0x7

	b.volatile = Region{b.pos, b.pos}

	otherPos := Coord{b.pos.X + dx, b.pos.Y + dy, b.pos.Z + dz}
	b.doFusion(otherPos, true)

	b.traceChannel <- []byte{b1}
	_ = <-b.traceChannel
}

func (b *Bot) FusionS(dx, dy, dz int) {
	//	fmt.Printf("FusionS <%d,%d,%d>\n", dx, dy, dz)
	nd := byte(encodeNearCoordDiff(dx, dy, dz))
	b1 := (nd << 3) + 0x6

	b.volatile = Region{b.pos, b.pos}

	otherPos := Coord{b.pos.X + dx, b.pos.Y + dy, b.pos.Z + dz}
	b.doFusion(otherPos, false)

	b.traceChannel <- []byte{b1}
	_ = <-b.traceChannel
}

func (b *Bot) Fission(dx, dy, dz int, m int) *Bot {
	//	fmt.Printf("Fission <%d,%d,%d> %d\n", dx, dy, dz, m)
	b.checkFusion()
	nd := byte(encodeNearCoordDiff(dx, dy, dz))
	b1 := (nd << 3) + 0x5

	newBotPos := Coord{b.pos.X + dx, b.pos.Y + dy, b.pos.Z + dz}
	newBot := Bot{id: b.seeds[0], pos: newBotPos, volatile: Region{newBotPos, newBotPos},
		traceChannel: make(chan []byte),
		seeds:        b.seeds[1 : m+1], builder: b.builder}
	b.seeds = b.seeds[m+1:]
	b.builder.bots = append(b.builder.bots, &newBot)

	b.traceChannel <- []byte{b1, byte(m)}
	_ = <-b.traceChannel
	return &newBot
}

func (b *Builder) IsSupported(x, y, z int) bool {
	if y == 0 {
		return true
	}
	if b.supported[b.GridOffset(Coord{x, y, z})] {
		return true
	} else if (x > 0) && b.supported[b.GridOffset(Coord{x - 1, y, z})] {
		return true
	} else if (x < b.r-1) && b.supported[b.GridOffset(Coord{x + 1, y, z})] {
		return true
	} else if (y > 0) && b.supported[b.GridOffset(Coord{x, y - 1, z})] {
		return true
	} else if (y < b.r-1) && b.supported[b.GridOffset(Coord{x, y + 1, z})] {
		return true
	} else if (z > 0) && b.supported[b.GridOffset(Coord{x, y, z - 1})] {
		return true
	} else if (z < b.r-1) && b.supported[b.GridOffset(Coord{x, y, z + 1})] {
		return true
	}
	return false
}

func (b *Bot) Fill(dx, dy, dz int) {
	//	fmt.Printf("Fill <%d,%d,%d>\n", dx, dy, dz)
	nd := byte(encodeNearCoordDiff(dx, dy, dz))
	b1 := (nd << 3) + 0x3

	gridx := b.pos.X + dx
	gridy := b.pos.Y + dy
	gridz := b.pos.Z + dz

	//fmt.Printf("Fill <%d,%d,%d>\n", gridx, gridy, gridz)
	if !b.builder.IsSupported(gridx, gridy, gridz) {
		panic(fmt.Sprintf("Tried to fill unsupported location at %d,%d,%d", gridx, gridy, gridz))
	}
	fillPos := Coord{gridx, gridy, gridz}
	//	fmt.Printf("Filling %d,%d,%d\n", gridx, gridy, gridz)
	b.builder.grid[b.GridOffset(fillPos)] = 1
	b.markSupported(fillPos)
	b.volatile = Region{b.pos, fillPos}

	b.CheckSpaceViolations()

	b.traceChannel <- []byte{b1}
	_ = <-b.traceChannel
}

func (b *Bot) Void(dx, dy, dz int) {
	//	fmt.Printf("Void <%d,%d,%d>\n", dx, dy, dz)
	nd := byte(encodeNearCoordDiff(dx, dy, dz))
	b1 := (nd << 3) + 0x2

	gridx := b.pos.X + dx
	gridy := b.pos.Y + dy
	gridz := b.pos.Z + dz

	fillPos := Coord{gridx, gridy, gridz}
	//	fmt.Printf("Voiding %d,%d,%d\n", gridx, gridy, gridz)
	b.builder.grid[b.GridOffset(fillPos)] = 0
	b.recomputeSupported()

	b.volatile = Region{b.pos, fillPos}

	b.CheckSpaceViolations()

	b.traceChannel <- []byte{b1}
	_ = <-b.traceChannel
}

func imin(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func imax(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func getRegion(x1, y1, z1 int, x2, y2, z2 int) Region {
	lowx := imin(x1, x2)
	lowy := imin(y1, y2)
	lowz := imin(z1, z2)

	highx := imax(x1, x2)
	highy := imax(y1, y2)
	highz := imax(z1, z2)

	return Region{Coord{lowx, lowy, lowz}, Coord{highx, highy, highz}}
}

func numCorners(reg Region) int {
	c := 1
	if reg.c1.X != reg.c2.X {
		c = c * 2
	}
	if reg.c1.Y != reg.c2.Y {
		c = c * 2
	}
	if reg.c1.Z != reg.c2.Z {
		c = c * 2
	}
	return c
}

func (b *Bot) GFill(dx1, dy1, dz1 int, dx2, dy2, dz2 int) {
	fmt.Printf("GFill <%d,%d,%d> <%d,%d,%d>\n", dx1, dy1, dz1, dx2, dy2, dz2)
	nd := byte(encodeNearCoordDiff(dx1, dy1, dz1))
	b1 := (nd << 3) + 0x1

	pos := b.pos
	b.doGroupOp(1, getRegion(pos.X+dx1, pos.Y+dy1, pos.Z+dz1, pos.X+dx2, pos.Y+dy2, pos.Z+dz2))

	b.traceChannel <- []byte{b1, byte(dx2 + 30), byte(dy2 + 30), byte(dz2 + 30)}
	_ = <-b.traceChannel
}

func (b *Bot) GVoid(dx1, dy1, dz1 int, dx2, dy2, dz2 int) {
	fmt.Printf("GFill <%d,%d,%d> <%d,%d,%d>\n", dx1, dy1, dz1, dx2, dy2, dz2)
	nd := byte(encodeNearCoordDiff(dx1, dy1, dz1))
	b1 := (nd << 3) + 0x0

	pos := b.pos
	b.doGroupOp(0, getRegion(pos.X+dx1, pos.Y+dy1, pos.Z+dz1, pos.X+dx2, pos.Y+dy2, pos.Z+dz2))

	b.traceChannel <- []byte{b1, byte(dx2 + 30), byte(dy2 + 30), byte(dz2 + 30)}
	_ = <-b.traceChannel
}

func (b *Bot) doGroupOp(opType int, reg Region) {
	for i, grpOp := range b.builder.pendingGroupOps {
		if grpOp.region == reg && grpOp.opType == opType {
			grpOp.numBotsLeft--
			if grpOp.numBotsLeft == 0 {
				for x := reg.c1.X; x <= reg.c2.X; x++ {
					for y := reg.c1.Y; y <= reg.c2.Y; y++ {
						for z := reg.c1.Z; z <= reg.c2.Z; z++ {
							b.builder.grid[b.GridOffset(Coord{x, y, z})] = byte(opType)
						}
					}
				}

				b.recomputeSupported()
				b.CheckSpaceViolations()
			}
			b.builder.pendingGroupOps = append(b.builder.pendingGroupOps[:i], b.builder.pendingGroupOps[i+1:]...)
			return
		}
	}
	newOp := GroupOp{opType: opType, region: reg, numBotsLeft: numCorners(reg)}
	b.builder.pendingGroupOps = append(b.builder.pendingGroupOps, &newOp)
}

func (b *Bot) recomputeSupported() {
	b.builder.supported = make([]bool, b.builder.r*b.builder.r*b.builder.r)
	for x := 1; x < b.builder.r-1; x++ {
		for z := 1; z < b.builder.r-1; z++ {
			if b.IsFilled(x, 0, z) {
				b.markSupported(Coord{x, 0, z})
			}
		}
	}
}

func (b *Bot) markSupported(c Coord) {
	if b.builder.supported[b.GridOffset(c)] {
		//fmt.Printf("%d,%d,%d already marked supported\n", c.X, c.Y, c.Z)
		return
	}

	b.builder.supported[b.GridOffset(c)] = true
	//fmt.Printf("Marking %d,%d,%d as supported\n", c.X, c.Y, c.Z)
	markQueue := []Coord{c}

	for len(markQueue) > 0 {
		curr := markQueue[0]
		markQueue = markQueue[1:]
		x, y, z := curr.X, curr.Y, curr.Z
		if x > 0 && b.IsFilled(x-1, y, z) && !b.builder.supported[b.GridOffset(Coord{x - 1, y, z})] {
			newC := Coord{x - 1, y, z}
			//fmt.Printf("Marking %d,%d,%d as supported\n", x-1, y, z)
			b.builder.supported[b.GridOffset(newC)] = true
			markQueue = append(markQueue, newC)
		}
		if x < b.builder.r-1 && b.IsFilled(x+1, y, z) && !b.builder.supported[b.GridOffset(Coord{x + 1, y, z})] {
			newC := Coord{x + 1, y, z}
			b.builder.supported[b.GridOffset(newC)] = true
			//fmt.Printf("Marking %d,%d,%d as supported\n", x+1, y, z)
			markQueue = append(markQueue, newC)
		}
		if y > 0 && b.IsFilled(x, y-1, z) && !b.builder.supported[b.GridOffset(Coord{x, y - 1, z})] {
			newC := Coord{x, y - 1, z}
			b.builder.supported[b.GridOffset(newC)] = true
			//fmt.Printf("Marking %d,%d,%d as supported\n", x, y-1, z)
			markQueue = append(markQueue, newC)
		}
		if y < b.builder.r-1 && b.IsFilled(x, y+1, z) && !b.builder.supported[b.GridOffset(Coord{x, y + 1, z})] {
			newC := Coord{x, y + 1, z}
			b.builder.supported[b.GridOffset(newC)] = true
			//fmt.Printf("Marking %d,%d,%d as supported\n", x, y+1, z)
			markQueue = append(markQueue, newC)
		}
		if z > 0 && b.IsFilled(x, y, z-1) && !b.builder.supported[b.GridOffset(Coord{x, y, z - 1})] {
			newC := Coord{x, y, z - 1}
			b.builder.supported[b.GridOffset(newC)] = true
			//fmt.Printf("Marking %d,%d,%d as supported\n", x, y-1, z)
			markQueue = append(markQueue, newC)
		}
		if z < b.builder.r-1 && b.IsFilled(x, y, z+1) && !b.builder.supported[b.GridOffset(Coord{x, y, z + 1})] {
			newC := Coord{x, y, z + 1}
			b.builder.supported[b.GridOffset(newC)] = true
			//fmt.Printf("Marking %d,%d,%d as supported\n", x, y+1, z)
			markQueue = append(markQueue, newC)
		}
	}

}
func (b *Bot) IsFilled(x, y, z int) bool {
	fillPos := Coord{x, y, z}
	return b.builder.grid[b.GridOffset(fillPos)] != 0
}

func (b *Builder) IsFilled(x, y, z int) bool {
	fillPos := Coord{x, y, z}
	return b.grid[b.GridOffset(fillPos)] != 0
}

func (b *Bot) Goto(coord Coord) {
	//	fmt.Printf("Goto %d,%d,%d\n", coord.X, coord.Y, coord.Z)
	dx := coord.X - b.pos.X
	dy := coord.Y - b.pos.Y
	dz := coord.Z - b.pos.Z

	if dx == 0 && dy == 0 && dz == 0 {
		return
	}

	for Iabs(dx) > 5 {
		if dx < -15 {
			b.SMove(-15, 0, 0)
			dx = dx + 15
		} else if dx > 15 {
			b.SMove(15, 0, 0)
			dx = dx - 15
		} else {
			b.SMove(dx, 0, 0)
			dx = 0
		}
	}

	for Iabs(dy) > 5 {
		if dy < -15 {
			b.SMove(0, -15, 0)
			dy = dy + 15
		} else if dy > 15 {
			b.SMove(0, 15, 0)
			dy = dy - 15
		} else {
			b.SMove(0, dy, 0)
			dy = 0
		}
	}

	for Iabs(dz) > 5 {
		if dz < -15 {
			b.SMove(0, 0, -15)
			dz = dz + 15
		} else if dz > 15 {
			b.SMove(0, 0, 15)
			dz = dz - 15
		} else {
			b.SMove(0, 0, dz)
			dz = 0
		}
	}

	if dx == 0 && dy == 0 && dz == 0 {
		return
	}

	if dx == 0 {
		if dy == 0 {
			b.SMove(0, 0, dz)
		} else if dz == 0 {
			b.SMove(0, dy, 0)
		} else {
			b.LMove(0, dy, 0, 0, 0, dz)
		}
	} else if dy == 0 {
		if dx == 0 {
			b.SMove(0, 0, dz)
		} else if dz == 0 {
			b.SMove(dx, 0, 0)
		} else {
			b.LMove(dx, 0, 0, 0, 0, dz)
		}
	} else if dz == 0 {
		if dx == 0 {
			b.SMove(0, dy, 0)
		} else if dy == 0 {
			b.SMove(dx, 0, 0)
		} else {
			b.LMove(dx, 0, 0, 0, dy, 0)
		}
	} else {
		b.LMove(dx, 0, 0, 0, dy, 0)
		b.SMove(0, 0, dz)
	}
}

func isBetween(x, x1, x2 int) bool {
	if x1 < x2 {
		return x >= x1 && x <= x2
	} else {
		return x >= x2 && x <= x1
	}
}

func regionsOverlap(r1, r2 Region) bool {
	return isBetween(r1.c1.X, r2.c1.X, r2.c2.Z) ||
		isBetween(r1.c2.X, r2.c1.X, r2.c2.Z) ||
		isBetween(r1.c1.Y, r2.c1.Y, r2.c2.Y) ||
		isBetween(r1.c2.Y, r2.c1.Y, r2.c2.Y) ||
		isBetween(r1.c1.Z, r2.c1.Z, r2.c2.Z) ||
		isBetween(r1.c2.Z, r2.c1.Z, r2.c2.Z)
}

func (b *Bot) CheckSpaceViolations() {
	for _, otherBot := range b.builder.bots {
		if otherBot.id == b.id {
			continue
		}
		if regionsOverlap(b.volatile, otherBot.volatile) {
			panic(fmt.Sprintf("Bot %d volatile region overlaps with %d", b.id, otherBot.id))
			return
		}
	}
}

func (b *Bot) GotoShortest(x, y, z int) {
	ok, shortestPath := b.GetShortestPath(Coord{x, y, z})
	if !ok {
		panic("Tried to go to shortest path, but there isn't one")
		return
	}
	lastDirType := 0
	lastPathElem := shortestPath[0]
	if lastPathElem.X == x && lastPathElem.Y == y {
		lastDirType = 3
	} else if lastPathElem.X == x && lastPathElem.Z == z {
		lastDirType = 1
	}

	for _, pathElem := range shortestPath[1:] {
		currDirType := 0
		if lastPathElem.X == pathElem.X && lastPathElem.Y == pathElem.Y {
			currDirType = 3
		} else if lastPathElem.X == pathElem.X && lastPathElem.Z == pathElem.Z {
			currDirType = 1
		}
		if currDirType != lastDirType {
			b.Goto(lastPathElem)
			lastDirType = currDirType
		}
		lastPathElem = pathElem
	}
	b.Goto(lastPathElem)
}

var pathDirections = []Coord{Coord{-1, 0, 0}, Coord{1, 0, 0},
	Coord{0, -1, 0}, Coord{0, 1, 0},
	Coord{0, 0, -1}, Coord{0, 0, 1}}

var spInit bool = false
var pathQueue [][]Coord
var pathQueueLen []int
var lengths []int
var shortestPath []Coord
var shortestPathLen int

var spLock sync.Mutex

func (b *Bot) GetShortestPath(c Coord) (bool, []Coord) {
	spLock.Lock()
	defer spLock.Unlock()

	if !spInit {
		pathQueue = make([][]Coord, b.builder.r*b.builder.r*b.builder.r)
		for i := 0; i < b.builder.r*b.builder.r*b.builder.r; i++ {
			pathQueue[i] = []Coord{}
		}
		pathQueueLen = make([]int, b.builder.r*b.builder.r*b.builder.r)
		lengths = make([]int, b.builder.r*b.builder.r*b.builder.r)
		shortestPath = make([]Coord, 500)
		spInit = true
	} else {
		for i := range pathQueue {
			if len(pathQueue[i]) > 20 && i > 200 {
				pathQueue[i] = []Coord{}
			}
			pathQueueLen[i] = 0
		}
		if len(shortestPath) > 5000 {
			shortestPath = make([]Coord, 500)
		}
		shortestPathLen = 0
	}

	highest := 1 + b.builder.r*b.builder.r*b.builder.r

	for i := range lengths {
		lengths[i] = highest
	}

	lengths[b.GridOffset(c)] = 0
	if len(pathQueue[0]) < 1 {
		pathQueue[0] = append(pathQueue[0], c)
	} else {
		pathQueue[0][0] = c
	}
	pathQueueLen[0] = 1

	done := false

	xmul := b.builder.r * b.builder.r
	ymul := b.builder.r

	for !done {
		found := false
		var nextCoord Coord
		currDist := highest
		for i := range pathQueue {
			if pathQueueLen[i] > 0 {
				nextCoord = pathQueue[i][pathQueueLen[i]-1]
				pathQueueLen[i]--
				currDist = i
				found = true
				break
			}
		}
		if !found {
			break
		}

		for _, dir := range pathDirections {
			workCoord := Coord{nextCoord.X + dir.X, nextCoord.Y + dir.Y, nextCoord.Z + dir.Z}
			if workCoord == b.pos {
				lengths[workCoord.X*xmul+workCoord.Y*ymul+workCoord.Z] = currDist + 1
				done = true
				break
			}

			if workCoord.X < 0 || workCoord.X > b.builder.r-1 || workCoord.Y < 0 || workCoord.Y > b.builder.r-1 ||
				workCoord.Z < 0 || workCoord.Z > b.builder.r-1 {
				continue
			}
			if !b.IsFilled(workCoord.X, workCoord.Y, workCoord.Z) &&
				lengths[workCoord.X*xmul+workCoord.Y*ymul+workCoord.Z] == highest {
				lengths[workCoord.X*xmul+workCoord.Y*ymul+workCoord.Z] = currDist + 1
				if len(pathQueue[currDist+1]) == pathQueueLen[currDist+1] {
					pathQueue[currDist+1] = append(pathQueue[currDist+1], workCoord)
				} else {
					pathQueue[currDist+1][pathQueueLen[currDist+1]] = workCoord
				}
				pathQueueLen[currDist+1]++
			}
		}
	}

	if !done {
		fmt.Printf("No path from %d,%d,%d to %d,%d,%d\n", b.pos.X, b.pos.Y, b.pos.Z, c.X, c.Y, c.Z)
		return false, []Coord{}
	}

	currPos := b.pos
	currDist := lengths[currPos.X*xmul+currPos.Y*ymul+currPos.Z]

	done = false
	for !done {
		found := false
		for _, dir := range pathDirections {
			workCoord := Coord{currPos.X + dir.X, currPos.Y + dir.Y, currPos.Z + dir.Z}
			if workCoord == c {
				if len(shortestPath) == shortestPathLen {
					shortestPath = append(shortestPath, workCoord)
				} else {
					shortestPath[shortestPathLen] = workCoord
				}
				shortestPathLen++
				done = true
				break
			}

			if workCoord.X < 0 || workCoord.X > b.builder.r-1 || workCoord.Y < 0 || workCoord.Y > b.builder.r-1 ||
				workCoord.Z < 0 || workCoord.Z > b.builder.r-1 {
				continue
			}

			if lengths[workCoord.X*xmul+workCoord.Y*ymul+workCoord.Z] == currDist-1 {
				if len(shortestPath) == shortestPathLen {
					shortestPath = append(shortestPath, workCoord)
				} else {
					shortestPath[shortestPathLen] = workCoord
				}
				shortestPathLen++
				currPos = workCoord
				currDist -= 1
				found = true
				break
			}
		}
		if done {
			break
		}
		if !found {
			panic("There should have been a shortest path, but now I can't find it")
			return false, nil
		}
	}
	return true, shortestPath[:shortestPathLen]
}

func encodeCoordDiff(dx, dy, dz, offset int) (int, int) {
	if dx != 0 {
		return 1, dx + offset
	} else if dy != 0 {
		return 2, dy + offset
	} else if dz != 0 {
		return 3, dz + offset
	} else {
		return 1, 0 + offset
	}
}

func encodeShortCoordDiff(dx, dy, dz int) (int, int) {
	return encodeCoordDiff(dx, dy, dz, 5)
}

func encodeLongCoordDiff(dx, dy, dz int) (int, int) {
	return encodeCoordDiff(dx, dy, dz, 15)
}

func encodeNearCoordDiff(dx, dy, dz int) int {
	if dx != 0 && dy != 0 && dz != 0 {
		panic(fmt.Sprintf("Invalid near coordinate: %d,%d,%d", dx, dy, dz))
		return 0
	}
	return 9*(dx+1) + 3*(dy+1) + dz + 1
}
