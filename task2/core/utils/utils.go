package utils

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/webp"
)

func SaveImageToPNG(filename, folderName string, img image.Image) error {
	file, err := os.Create(fmt.Sprintf("./%s/%s", folderName, filename))
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func ReadImage(filePath string) image.Image {
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
