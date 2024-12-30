package main

import (
	"exam/task1/internal/transformers"
	"image/color"
	"log"
	"math"
)

func main() {
	DTW([]float64{0, 2, 0, 1, 0, 0}, []float64{0, 0, 0.5, 2, 0, 1, 0})
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

// https://en.wikipedia.org/wiki/Cosine_similarity
func Cosine(pixelA, pixelB color.Color) float64 {
	log.Println(pixelA, pixelB)

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

func DTW(vecA, vecB []float64) []float64 {
	dtw := make([][]float64, len(vecA)+1)

	for i := range len(vecA) + 1 {
		dtw[i] = make([]float64, len(vecB)+1)
		dtw[i][0] = math.Inf(0)
	}

	for j := range len(vecB) + 1 {
		dtw[0][j] = math.Inf(0)
	}

	dtw[0][0] = 0

	for i := 1; i <= len(vecA); i++ {
		for j := 1; j <= len(vecB); j++ {
			var (
				insertion = dtw[i-1][j]
				deletion  = dtw[i][j-1]
				match     = dtw[i-1][j-1]
			)
			dtw[i][j] = DummyDistance(vecA[i-1], vecB[j-1]) + math.Min(math.Min(insertion, deletion), match)
		}
	}

	for i := range len(dtw) {
		log.Println(dtw[len(dtw)-i-1])
	}

	// positions := make([]int)
	// pathValues := make([]float64, len(vecB))
	return nil
}
