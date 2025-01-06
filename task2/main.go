package main

import (
	"exam/task3/core/mathstuff"
	"exam/task3/core/utils"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/nfnt/resize"
)

const (
	// targetImgPath = "./img/batman.webp"
	targetImgPath = "./img/test2.webp"
	// targetImgPath = "./img/Lord.webp"
)

var (
	// Color params
	colorPalette [][]color.Color
	numOfColors  = 300

	// Image Params
	width   uint = 1024
	boxSize int  = 16

	// Generative params
	initPopulationSize = 100

	// Parallel params
	maxSubprocess = 8

	// Process time
	runningLimit time.Duration = time.Hour * 1
	folderName                 = strings.ReplaceAll(time.Now().Format(time.DateOnly)+"-"+time.Now().Format(time.Kitchen), ":", "_")

	// Hyperparameter
	crossOverProb float64 = 0.65
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

	os.MkdirAll(folderName, 0755)
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

type PopStat struct {
	Data  [][]int
	Score float64
}

func generatePopulation(initPopulationSize, colorPaletteRange, boxSize int, width uint) []PopStat {
	var (
		popu = make([]PopStat, initPopulationSize)
	)
	for i := range popu {
		newBorn := createOne(colorPaletteRange, boxSize, width)
		popu[i] = PopStat{
			Data:  newBorn,
			Score: math.Inf(0),
		}
	}
	return popu
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

func getNParents(n, limit int) []int {
	result := make([]int, n)

	for i := range result {
		for {
			id := rand.Intn(limit)
			unique := true
			for j := 0; j < i; j++ {
				if result[j] == id {
					unique = false
				}
			}

			if unique {
				result[i] = id
				break
			}
		}
	}
	return result
}

func Off(father int, others []int, populationArray []PopStat, crossOverProb float64) [][]int {
	rowSize, colSize := len(populationArray[0].Data), len(populationArray[0].Data[0])

	newBorn := make([][]int, rowSize)

	for i := 0; i < rowSize; i++ {
		newBorn[i] = make([]int, colSize)
		for j := 0; j < colSize; j++ {
			var colorFrom int
			if rand.Float64() < crossOverProb {
				getFrom := rand.Intn(len(others) - 1)
				if getFrom >= len(others) {
					getFrom = len(others) - 1
				}
				colorFrom = others[getFrom]
			} else {
				colorFrom = father
			}
			newBorn[i][j] = populationArray[colorFrom].Data[i][j]
		}
	}

	return newBorn
}

func DrawOverImage(targetImage image.Image, offSpritn PopStat, boxSize int) image.Image {
	bounds := targetImage.Bounds()
	ca := image.NewRGBA(bounds)
	draw.Draw(ca, bounds, targetImage, bounds.Min, draw.Src)
	mod := targetImage.Bounds().Dx() / boxSize
	// rowSize, colSize := len(offSpritn.Data), len(offSpritn.Data[0])
	for x := 0; x < targetImage.Bounds().Dx(); x++ {
		for y := 0; y < targetImage.Bounds().Dy(); y++ {
			colorX, colorY := int(x/mod), int(y/mod)

			cols := colorPalette[offSpritn.Data[colorX][colorY]]
			ca.Set(x, y, cols[0])
		}
	}

	return ca
}

func main() {
	var (
		targetImageOrigin image.Image = utils.ReadImage(targetImgPath)
		targetImage                   = resize.Resize(width, width, targetImageOrigin, resize.Lanczos3)
		initPopulation                = generatePopulation(initPopulationSize, len(colorPalette), boxSize, width)
	)

	startTime := time.Now()
	iterationsCounter := 0

	deltasReport := runningLimit / 10
	lastReport := time.Now()
	for time.Since(startTime) < runningLimit {
		if time.Since(lastReport) > deltasReport {
			log.Println("Time of computing: ", time.Since(startTime)-runningLimit, "/", runningLimit)
			log.Println("Remaining computing time:: ", time.Until(startTime.Add(runningLimit)))
			lastReport = time.Now()
		}

		var (
			parentsIndex  = getNParents(4, initPopulationSize)
			offSprint     = Off(parentsIndex[0], parentsIndex[1:], initPopulation, crossOverProb)
			offFitness    = 0.0
			fatherFitness = initPopulation[parentsIndex[0]].Score
			wg            = sync.WaitGroup{}
		)

		if fatherFitness == math.Inf(0) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fatherFitness = parallelFitnessFunction(targetImage, boxSize, initPopulation[parentsIndex[0]].Data)
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			offFitness = parallelFitnessFunction(targetImage, boxSize, offSprint)
		}()
		wg.Wait()
		if offFitness < fatherFitness {
			log.Println("Off is better than father ", offFitness, " < ", fatherFitness)
			initPopulation[parentsIndex[0]] = PopStat{
				Data:  offSprint,
				Score: offFitness,
			}
		}
		if initPopulation[parentsIndex[0]].Score == math.Inf(0) {
			initPopulation[parentsIndex[0]].Score = fatherFitness
		}
		iterationsCounter += 1
	}

	log.Println("Total iterations: ", iterationsCounter)

	sort.Slice(initPopulation, func(i, j int) bool {
		return initPopulation[i].Score < initPopulation[j].Score
	})

	log.Println(initPopulation[0].Score, initPopulation[1].Score, initPopulation[2].Score)

	for i, data := range initPopulation[:5] {
		dst := DrawOverImage(targetImage, data, boxSize)
		if err := utils.SaveImageToPNG(fmt.Sprintf("TOP-%d-%d.png", i+1, int(data.Score)), folderName, dst); err != nil {
			log.Println("Errpr saving the last image", err.Error())
		}

		overlapped := overLapWinner(targetImage, dst)
		if err := utils.SaveImageToPNG(fmt.Sprintf("TOP-%d-overlapped.png", i+1), folderName, overlapped); err != nil {
			log.Println("Errpr saving the overlapped image", err.Error())
		}
	}

	utils.SaveImageToPNG("origin.png", folderName, targetImage)

	bigger := 0.0
	index := 0
	for i := range initPopulation {
		if initPopulation[i].Score != math.Inf(0) && bigger < initPopulation[i].Score {
			bigger = initPopulation[i].Score
			index = i
		}
	}

	dst := DrawOverImage(targetImage, initPopulation[index], boxSize)
	utils.SaveImageToPNG(fmt.Sprintf("worst-%d.png", int(initPopulation[index].Score)), folderName, dst)
	overlapped := overLapWinner(targetImage, dst)
	if err := utils.SaveImageToPNG("worst-overlapped.png", folderName, overlapped); err != nil {
		log.Println("Errpr saving the overlapped image", err.Error())
	}

	log.Printf("Biger: %.2f, Index: %d, Population without being checked: %d \n", bigger, index, len(initPopulation)-index)
	log.Println("Done")

}

/*
 * ChatGPT Code Section
 *
 */

func blendColors(dst, src color.Color, alpha uint8) color.Color {
	// Convert colors to RGBA
	dr, dg, db, da := dst.RGBA()
	sr, sg, sb, sa := src.RGBA()

	// Normalize alpha to a 0-1 scale
	alphaF := float64(alpha) / 255.0

	// Perform alpha blending
	r := uint8((1-alphaF)*float64(dr/257) + alphaF*float64(sr/257))
	g := uint8((1-alphaF)*float64(dg/257) + alphaF*float64(sg/257))
	b := uint8((1-alphaF)*float64(db/257) + alphaF*float64(sb/257))
	a := uint8((1-alphaF)*float64(da/257) + alphaF*float64(sa/257))

	return color.RGBA{r, g, b, a}
}
func overLapWinner(targetImage, winner image.Image) image.Image {
	// Create a 16x16 RGBA image
	img := image.NewRGBA(image.Rect(0, 0, targetImage.Bounds().Dx(), targetImage.Bounds().Dy()))
	alpha := uint8(128) // 50% transparency

	// Fill the image with a solid background color
	bgColor := color.RGBA{255, 255, 255, 255} // White
	for y := 0; y < targetImage.Bounds().Dy(); y++ {
		for x := 0; x < targetImage.Bounds().Dx(); x++ {
			img.Set(x, y, bgColor)
			blendedColor := blendColors(targetImage.At(x, y), winner.At(x, y), alpha)
			img.Set(x, y, blendedColor)
		}
	}

	return img
}

/*
 *  END ChatGPT Code Section
 *
 */
