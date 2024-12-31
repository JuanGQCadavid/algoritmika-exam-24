package main

import (
	"exam/task1/internal/transformers"
	"exam/task1/internal/utls"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	"image/png"

	"github.com/nfnt/resize"
)

const (
	resultsFile = "data.txt"
)

func main() {
	imgs := transformers.WalkThrough("./imgs/A")
	log.Println(Cosine(imgs[0].At(0, 0), imgs[1].At(0, 0)))

	//256
	//

	a, b := resize.Resize(128, 0, imgs[0], resize.Lanczos3), resize.Resize(256, 0, imgs[1], resize.Lanczos3)

	err := saveImageToPNG("a.png", a)
	if err != nil {
		panic(err)
	}

	flatA, flatB := make([]color.Color, a.Bounds().Dx()*a.Bounds().Dy()), make([]color.Color, b.Bounds().Dx()*b.Bounds().Dy())

	counter := 0
	for i := 0; i < a.Bounds().Dx(); i++ {
		for j := 0; j < a.Bounds().Dy(); j++ {
			flatA[counter] = a.At(i, j)
			counter++
		}
	}

	counter = 0
	for i := 0; i < b.Bounds().Dx(); i++ {
		for j := 0; j < b.Bounds().Dy(); j++ {
			flatB[counter] = b.At(i, j)
			counter++
		}
	}

	log.Println(len(flatA), flatA[0])

	post, val := DTW(flatA, flatB)

	log.Println(len(post), len(val))

	if err := utls.WriteDataToFile(resultsFile, post, val); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	// Read data back from file
	readIntData, readFloatData, err := utls.ReadDataFromFile(resultsFile)
	if err != nil {
		fmt.Println("Error reading from file:", err)
		return
	}

	// for i := range val {

	if len(readFloatData) != len(val) {
		log.Panicln("Different values for val")
	}
	// }

	for i := range post {
		for j := range post[i] {
			if post[i][j] != readIntData[i][j] {
				log.Panicln("Different values for post: ", post[i][j], readIntData[i][j])
			}
		}

	}

}

func readImages() {
	imgs := transformers.WalkThrough("./imgs/A")
	// for _, img := range imgs {
	// 	log.Println(img.Bounds())
	// 	point0 := img.At(0, 0)
	// 	log.Println(point0.RGBA())
	// }
	log.Println(Cosine(imgs[0].At(0, 0), imgs[1].At(0, 0)))
}

func saveImageToPNG(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

// https://en.wikipedia.org/wiki/Cosine_similarity
func Cosine(pixelA, pixelB color.Color) float64 {
	// log.Println(pixelA, pixelB)

	aR, aG, aB, aA := pixelA.RGBA()
	vect1 := []uint32{
		aR, aG, aB, aA,
	}

	bR, bG, bB, bA := pixelB.RGBA()
	vect2 := []uint32{
		bR, bG, bB, bA,
	}

	// dot-product two vectors
	// to calculate A·B
	dotProduct := 0.0
	for i := range vect1 {
		dotProduct += float64(vect1[i]) * float64(vect2[i])
	}

	// to calculate |A|*|B|
	sum1 := 0.0
	for _, v := range vect1 {
		sum1 += math.Pow(float64(v), 2)
	}
	sum2 := 0.0
	for _, v := range vect2 {
		sum2 += math.Pow(float64(v), 2)
	}

	magnitude := math.Sqrt(sum1) * math.Sqrt(sum2)
	if magnitude == 0 {
		return 0.0
	}
	return float64(dotProduct) / float64(magnitude)
}

func DummyDistance(a, b float64) float64 {
	return math.Abs(a - b)
}

func DTW(vecA, vecB []color.Color) ([][]int, []float64) {
	dtw := make([][]float64, len(vecA)+1)

	for i := range len(vecA) + 1 {
		dtw[i] = make([]float64, len(vecB)+1)
		dtw[i][0] = math.Inf(0)
	}

	for j := range len(vecB) + 1 {
		dtw[0][j] = math.Inf(0)
	}

	dtw[0][0] = 0

	log.Println("Matrix created")

	for i := 1; i <= len(vecA); i++ {
		for j := 1; j <= len(vecB); j++ {
			var (
				insertion = dtw[i-1][j]
				deletion  = dtw[i][j-1]
				match     = dtw[i-1][j-1]
			)
			dtw[i][j] = Cosine(vecA[i-1], vecB[j-1]) + math.Min(math.Min(insertion, deletion), match)
		}
	}

	log.Println("Matrix filled")
	// for i := range len(dtw) {
	// 	log.Println(dtw[len(dtw)-i-1])
	// }

	positions := make([][]int, 0)
	pathValues := make([]float64, 0)

	i, j := len(dtw)-1, len(dtw[0])-1

	for {
		if i+j == 0 {
			break
		}

		positions = append(positions, []int{i, j})
		pathValues = append(pathValues, dtw[i][j])

		var (
			insertion = dtw[i-1][j]
			deletion  = dtw[i][j-1]
			match     = dtw[i-1][j-1]
		)

		smallest := insertion
		newI, newJ := i-1, j
		if deletion < smallest {
			smallest = deletion
			newI, newJ = i, j-1
		}
		if match < smallest {
			newI, newJ = i-1, j-1
		}

		i, j = newI, newJ
	}

	log.Println("Positions and values done")

	return positions, pathValues
}
