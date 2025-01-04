package main

import (
	"exam/task3/core/mathstuff"
	"exam/task3/core/utils"
	"image"
	"image/color"
	"log"
	"math"
	"sync"
	"time"

	"math/rand"

	"github.com/nfnt/resize"
)

const (
	targetImgPath = "./img/Lord.webp"
)

var (
	// Color params
	colorPalette [][]color.Color
	numOfColors  = 300

	// Image Params
	width   uint = 1024
	boxSize int  = 16

	// Generative params
	initPopulationSize = int(1e4)

	// Parallel params
	maxSubprocess = 8
)

func init() {
	colorPalette = make([][]color.Color, numOfColors)
	for i := range colorPalette {
		colorSelected := color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 255,
		}

		colorPalette[i] = make([]color.Color, boxSize*boxSize)
		for j := range colorPalette[i] {
			colorPalette[i][j] = colorSelected
		}
	}
}

func createOne(colorPaletteRange, boxSize int, width uint) [][]int {
	shape := int(width) / boxSize
	child := make([][]int, shape)
	for i := range child {
		child[i] = make([]int, shape)
		for j := range child[i] {
			child[i][j] = rand.Intn(colorPaletteRange)
		}
	}
	return child
}

func generatePopulation(initPopulationSize, colorPaletteRange, boxSize int, width uint) [][][]int {
	var (
		popu = make([][][]int, initPopulationSize)
	)
	for i := range popu {
		popu[i] = createOne(colorPaletteRange, boxSize, width)
	}
	return popu
}

func generateBoxLimits(a image.Image, size int) [][]int {
	result := make([][]int, 0)

	for y := 0; y < a.Bounds().Dy(); y += size {
		for x := 0; x < a.Bounds().Dx(); x += size {
			var (
				xLimit = x + size
				yLimit = y + size
			)

			if xLimit > a.Bounds().Dx() {
				xLimit = a.Bounds().Dx() - 1
			}
			if yLimit > a.Bounds().Dy() {
				yLimit = a.Bounds().Dy() - 1
			}

			result = append(result, []int{x, y, xLimit, yLimit})
		}
	}

	return result
}

func extractVectorBox(img image.Image, startX, startY, endX, endY int) []color.Color {
	result := make([]color.Color, (endX-startX)*(endY-startY))
	counter := 0
	for x := startX; x < endX; x++ {
		for y := startY; y < endY; y++ {
			result[counter] = img.At(x, y)
			counter += 1
		}
	}
	return result
}

func fitnessFunction(targetImage image.Image, boxSize int, offSprint [][]int) float64 {
	totalSum := 0.0
	for y, row := range offSprint {
		for x := range row {
			var (
				startX = x * boxSize
				endX   = x*boxSize + boxSize

				startY = y * boxSize
				endY   = y*boxSize + boxSize
			)
			// println(fmt.Sprintf("(%d,%d) - (%d,%d)", startX, startY, endX, endY))
			imgBox := extractVectorBox(targetImage, startX, startY, endX, endY)
			totalSum += mathstuff.DTW(imgBox, colorPalette[row[x]])
		}
	}
	return totalSum
}

func parallelFitnessFunction(targetImage image.Image, boxSize int, offSprint [][]int) float64 {
	chunks := len(offSprint) / maxSubprocess
	ch := make(chan float64, chunks)
	wg := sync.WaitGroup{}

	for i := 0; i < chunks; i++ {
		startRow := i * chunks
		endRow := startRow + chunks
		wg.Add(1)

		go func(id, startRow, endRow int) {
			defer wg.Done()
			mineTotalSum := 0.0
			for y := startRow; y < endRow; y++ {
				row := offSprint[y]
				for x := range row {
					var (
						startX = x * boxSize
						endX   = x*boxSize + boxSize

						startY = y * boxSize
						endY   = y*boxSize + boxSize
					)
					// println(fmt.Sprintf("(%d,%d) - (%d,%d)", startX, startY, endX, endY))
					imgBox := extractVectorBox(targetImage, startX, startY, endX, endY)
					mineTotalSum += mathstuff.DTW(imgBox, colorPalette[row[x]])
				}
			}
			ch <- mineTotalSum
		}(i, startRow, endRow)
	}

	wg.Wait()
	close(ch)

	globalSum := 0.0
	for val := range ch {
		globalSum += val
	}
	return globalSum
}

func main() {
	var (
		targetImage image.Image = utils.ReadImage(targetImgPath)
	)
	a := resize.Resize(width, width, targetImage, resize.Lanczos3)
	utils.SaveImageToPNG("test.png", "./", a)
	initPopulation := generatePopulation(initPopulationSize, len(colorPalette), boxSize, width)

	log.Println(len(initPopulation[0]), len(initPopulation[0][0]))
	// log.Println(initPopulation[0][0])

	// fitnessFunction(targetImage, boxSize, int(width), initPopulation[0])
	log.Print("Single: ")
	tick := time.Now()
	s := fitnessFunction(targetImage, boxSize, initPopulation[0])
	println(int(s))
	log.Println(time.Since(tick))

	log.Print("Parallel: ")
	tick = time.Now()
	p := parallelFitnessFunction(targetImage, boxSize, initPopulation[0])
	println(int(p))
	log.Println(time.Since(tick))

	log.Println(math.Abs(s - p))
}
