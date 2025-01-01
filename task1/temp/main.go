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
	var reshapeSize uint = 256
	a, b := resize.Resize(reshapeSize, 0, imgs[0], resize.Lanczos3), resize.Resize(reshapeSize, 0, imgs[1], resize.Lanczos3)

	if err := saveImageToPNG("a.png", a); err != nil {
		panic(err)
	}
	if err := saveImageToPNG("b.png", b); err != nil {
		panic(err)
	}

	flatA, flatB := mathstuff.FlatImage(a), mathstuff.FlatImage(b)

	post, val, insertOps, deleteOps, matchOps := mathstuff.DTW(flatA, flatB)
	log.Println(len(post), len(val), len(insertOps), len(deleteOps), len(matchOps))

	bounds := a.Bounds()
	ca := image.NewRGBA(bounds)
	cb := image.NewRGBA(bounds)

	draw.Draw(ca, bounds, a, bounds.Min, draw.Src)
	draw.Draw(cb, bounds, b, bounds.Min, draw.Src)

	log.Println(a.Bounds().Dx(), b.Bounds().Dy())

	for _, v := range deleteOps {
		i := int(v[0] / a.Bounds().Dx())
		j := int(v[0] % a.Bounds().Dx())
		ca.Set(i, j, color.RGBA{255, 0, 0, 255})
		cb.Set(i, j, color.RGBA{255, 0, 0, 255})
	}

	for _, v := range insertOps {
		i := int(v[0] / a.Bounds().Dx())
		j := int(v[0] % a.Bounds().Dx())
		ca.Set(i, j, color.RGBA{0, 255, 0, 255})
		cb.Set(i, j, color.RGBA{0, 255, 0, 255})
	}

	if err := saveImageToPNG("ca.png", ca); err != nil {
		panic(err)
	}
	if err := saveImageToPNG("cb.png", cb); err != nil {
		panic(err)
	}

}

func saveImageToPNG(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}
