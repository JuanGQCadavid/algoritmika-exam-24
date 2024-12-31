package main

import (
	"exam/task1/internal/mathstuff"
	"exam/task1/internal/transformers"
	"image"
	"log"
	"os"

	"image/color"
	"image/draw"
	"image/png"

	"github.com/nfnt/resize"
)

const (
	resultsFile = "data.txt"
)

func main() {
	imgs := transformers.WalkThrough("./imgs/A")

	//256
	//128
	var reshapeSize uint = 64
	a, b := resize.Resize(reshapeSize, 0, imgs[0], resize.Lanczos3), resize.Resize(reshapeSize, 0, imgs[1], resize.Lanczos3)

	if err := saveImageToPNG("a.png", a); err != nil {
		panic(err)
	}
	if err := saveImageToPNG("b.png", b); err != nil {
		panic(err)
	}

	flatA, flatB := mathstuff.FlatImage(a), mathstuff.FlatImage(b)
	log.Println(len(flatA), flatA[0])
	log.Println(len(flatB), flatB[0])

	// Write data to file
	post, val, insertOps, deleteOps, matchOps := mathstuff.DTW(flatA, flatB)
	// if err := utls.WriteDataToFile(resultsFile, post, val); err != nil {
	// 	fmt.Println("Error writing to file:", err)
	// 	return
	// }

	// Read data back from file
	// post, val, err := utls.ReadDataFromFile(resultsFile)
	// if err != nil {
	// 	fmt.Println("Error reading from file:", err)
	// 	return
	// }

	log.Println(len(post), len(val), len(insertOps), len(deleteOps), len(matchOps))

	// min := math.Inf(0)
	// max := 0.0
	// avg := 0.0

	// for _, v := range val {
	// 	min = math.Min(v, min)
	// 	max = math.Max(v, max)
	// 	avg += v
	// }

	// avg = avg / float64(len(val))

	// log.Println(min, max, avg)
	// originalShape := a.Bounds()

	bounds := a.Bounds()
	ca := image.NewRGBA(bounds)
	cb := image.NewRGBA(bounds)

	draw.Draw(ca, bounds, a, bounds.Min, draw.Src)
	draw.Draw(cb, bounds, b, bounds.Min, draw.Src)

	log.Println(a.Bounds(), b.Bounds())

	for _, v := range deleteOps {
		i := int(v[0] / a.Bounds().Dx())
		j := int(v[0] % a.Bounds().Dx())

		ca.Set(i, j, color.RGBA{255, 0, 0, 255})
		cb.Set(i, j, color.RGBA{255, 0, 0, 255})

	}

	if err := saveImageToPNG("ca.png", ca); err != nil {
		panic(err)
	}
	if err := saveImageToPNG("cb.png", cb); err != nil {
		panic(err)
	}
	// if err := saveImageToPNG("b.png", b); err != nil {
	// 	panic(err)
	// }

}

func readAndSave() {

}

func saveImageToPNG(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}
