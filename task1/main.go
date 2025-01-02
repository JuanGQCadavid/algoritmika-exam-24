package main

import (
	"exam/task1/internal/mathstuff"
	"exam/task1/internal/transformers"
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"image/color"
	"image/draw"
	"image/png"

	"github.com/nfnt/resize"
)

const (
	resultsFile = "data.txt"
)

type DTWDistance struct {
	TargetX  int
	TargetY  int
	DestX    int
	DestY    int
	Distance float64
	Target   *image.Image
	Dest     *image.Image
}

type DWTPerLine struct {
	Target        *image.Image
	Dest          *image.Image
	TotalDistance float64
	Distances     []DTWDistance
}

type MetaData struct {
	Name          string
	Img           *image.Image
	TotalDistance float64
}

func extractVectorLine(img image.Image, pos int) []color.Color {
	result := make([]color.Color, img.Bounds().Dx())

	for i := 0; i < img.Bounds().Dx(); i++ {
		result[i] = img.At(i, pos)
	}

	return result
}

var (
	colors = []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 0, G: 255, B: 0, A: 255},     // Green
		{R: 0, G: 0, B: 255, A: 255},     // Blue
		{R: 255, G: 255, B: 0, A: 255},   // Yellow
		{R: 0, G: 255, B: 255, A: 255},   // Cyan
		{R: 255, G: 0, B: 255, A: 255},   // Magenta
		{R: 128, G: 0, B: 128, A: 255},   // Purple
		{R: 255, G: 165, B: 0, A: 255},   // Orange
		{R: 128, G: 128, B: 128, A: 255}, // Gray
		{R: 0, G: 128, B: 0, A: 255},     // Dark Green
	}
)

func GetDTWPerLine(target, dest image.Image, reshapeSize uint) *DWTPerLine {
	a, b := resize.Resize(reshapeSize, 0, target, resize.Lanczos3), resize.Resize(reshapeSize, 0, dest, resize.Lanczos3)

	response := DWTPerLine{
		Target:    &a,
		Dest:      &b,
		Distances: make([]DTWDistance, a.Bounds().Dy()),
	}

	report := int(a.Bounds().Dy() / 10)
	for i := 0; i < a.Bounds().Dy(); i++ {
		var (
			vectorA      = extractVectorLine(a, i)
			minVectorPos = 0
			minValue     = math.Inf(0)
		)

		if i%report == 0 {
			log.Println(i, "/", a.Bounds().Dy())
		}

		for j := 0; j < b.Bounds().Dy(); j++ {
			vectorB := extractVectorLine(b, j)
			_, val, _, _, _ := mathstuff.DTW(vectorA, vectorB)

			if val[0] < minValue {
				minVectorPos = j
				minValue = val[0]
			}
		}

		response.Distances[i] = DTWDistance{
			TargetY:  i,
			DestY:    minVectorPos,
			Distance: minValue,
			Target:   &a,
			Dest:     &b,
		}

		response.TotalDistance += minValue
	}

	return &response
}

var (
	folderName = strings.ReplaceAll(time.Now().Format(time.DateOnly)+"-"+time.Now().Format(time.Kitchen), ":", "_")
)

func init() {
	os.MkdirAll(folderName, 0755)
}

func main() {
	tick := time.Now()

	// ////////////////////
	//
	// INIT
	//
	// ////////////////////

	dest := transformers.WalkThrough("./img/all")
	target := transformers.WalkThrough("./img/B_target")[0]
	var reshapeSize uint = 128 // 256 128 64

	wg := sync.WaitGroup{}
	ch := make(chan *DWTPerLine, len(dest))

	for _, d := range dest {
		wg.Add(1)

		go func(destination image.Image, reshapeSize uint) {
			defer wg.Done()
			ch <- GetDTWPerLine(target, destination, reshapeSize)
		}(d, reshapeSize)
	}

	// ////////////////////
	//
	// Joint
	//
	// ////////////////////

	wg.Wait()
	close(ch)

	resultsMetadata := make(map[*image.Image]MetaData)
	resultsData := make([]DTWDistance, 0)
	counter := 0

	for data := range ch {
		resultsData = append(resultsData, data.Distances...)
		resultsMetadata[data.Dest] = MetaData{
			Name:          fmt.Sprintf("DEST-%d", counter),
			TotalDistance: data.TotalDistance,
			Img:           data.Dest,
		}
		counter += 1
	}

	// ////////////////////
	//
	// SORTING
	//
	// ////////////////////

	// Lines

	sort.Slice(resultsData, func(i, j int) bool {
		return resultsData[i].Distance < resultsData[j].Distance
	})

	for i := 1; i < len(resultsData); i++ {
		if resultsData[i-1].Distance > resultsData[i].Distance {
			log.Fatal("It is not sorted")
		}
	}

	// Files -> metaDatakeys will be the sorted keys

	metaDatakeys := make([]*image.Image, 0)
	for key := range resultsMetadata {
		metaDatakeys = append(metaDatakeys, key)
	}

	sort.Slice(metaDatakeys, func(i, j int) bool {
		return resultsMetadata[metaDatakeys[i]].TotalDistance < resultsMetadata[metaDatakeys[j]].TotalDistance
	})

	// ////////////////////
	//
	// REPORTING
	//
	// ////////////////////

	// Lines
	topTenLines := make(map[*image.Image][]DTWDistance)
	stringReport := "Pos; Target Row; Dest ID; Dest Row; Distance \n"
	log.Println("Top ten lines")
	for i, val := range resultsData[0:10] {
		if topTenLines[val.Dest] == nil {
			topTenLines[val.Dest] = make([]DTWDistance, 0)
		}
		topTenLines[val.Dest] = append(topTenLines[val.Dest], val)
		stringReport += fmt.Sprintf(" %d; %d; %s; %d; %f \n", i, val.TargetY, resultsMetadata[val.Dest].Name, val.DestY, val.Distance)
	}
	log.Println(stringReport)

	targetImage := *resultsData[0].Target
	bounds := targetImage.Bounds()

	ca := image.NewRGBA(bounds)
	draw.Draw(ca, bounds, targetImage, bounds.Min, draw.Src)
	counter = 0

	for i, val := range topTenLines {
		destImage := *i
		cb := image.NewRGBA(bounds)

		draw.Draw(cb, bounds, destImage, bounds.Min, draw.Src)

		for i := range len(val) {
			line := resultsData[i]
			for x := 0; x < targetImage.Bounds().Dx(); x++ {
				ca.Set(x, line.TargetY, colors[counter])
				cb.Set(x, line.DestY, colors[counter])
			}
			counter += 1
		}

		if err := saveImageToPNG(fmt.Sprintf("%s-lines.png", resultsMetadata[i].Name), cb); err != nil {
			panic(err)
		}
	}

	if err := saveImageToPNG("target-lines.png", ca); err != nil {
		panic(err)
	}

	// Files
	log.Println("Top 5 files: ")
	topFiles := "FileName; Distance \n"
	for i, val := range metaDatakeys[0:5] {
		log.Println(val)
		log.Println(resultsMetadata[val])

		topFiles += fmt.Sprintf("%d; %s; %f \n", i, resultsMetadata[val].Name, resultsMetadata[val].TotalDistance)
		if err := saveImageToPNG(fmt.Sprintf("%s.png", resultsMetadata[val].Name), *resultsMetadata[val].Img); err != nil {
			panic(err)
		}
	}
	log.Println(topFiles)

	if f, err := os.Create(fmt.Sprintf("./%s/results.txt", folderName)); err != nil {
		log.Fatalln(err.Error())
	} else {
		f.WriteString(fmt.Sprintf("Resolution: %d, time taken: %s \n", reshapeSize, time.Since(tick)))
		f.WriteString("Lines\n")
		f.WriteString(stringReport)
		f.WriteString("\n")
		f.WriteString("Files\n")
		f.WriteString(topFiles)
		f.WriteString("\n")

	}

}

func saveImageToPNG(filename string, img image.Image) error {
	file, err := os.Create(fmt.Sprintf("./%s/%s", folderName, filename))
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}
