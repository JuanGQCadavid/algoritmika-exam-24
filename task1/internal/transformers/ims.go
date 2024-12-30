package transformers

import (
	"image"
	"log"
	"os"
	"strings"

	"golang.org/x/image/webp"
)

func testReadImage() {
	img := readImage("./imgs/A/img_43.webp")

	log.Println(img.Bounds())

	point0 := img.At(0, 0)
	log.Println(point0.RGBA())
}

func WalkThrough(basePath string) []image.Image {
	fs, err := os.ReadDir(basePath)

	if err != nil {
		log.Panicln("Error while reading path ", basePath, err.Error())
	}

	imgs := make([]image.Image, 0)

	for _, fileId := range fs {

		if basePath[len(basePath)-1] != '/' {
			basePath = basePath + "/"
		}

		if !fileId.IsDir() && strings.Contains(fileId.Name(), ".webp") {
			pathToImage := basePath + fileId.Name()
			log.Println(pathToImage)
			imgs = append(imgs, readImage(pathToImage))
		}
	}
	return imgs
}

func readImage(filePath string) image.Image {
	f, err := os.Open(filePath)

	if err != nil {
		log.Panicln("Error while opening file at readImage", err.Error())
	}

	defer f.Close()

	img, err := webp.Decode(f)

	if err != nil {
		log.Panicln("Error while deconing image ", err.Error())
	}

	return img
}
