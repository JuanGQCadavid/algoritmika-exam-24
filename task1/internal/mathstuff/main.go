package mathstuff

import (
	"image"
	"image/color"
	"math"
)

func FlatImage(a image.Image) []color.Color {

	flatA := make([]color.Color, a.Bounds().Dx()*a.Bounds().Dy())

	counter := 0
	for i := 0; i < a.Bounds().Dx(); i++ {
		for j := 0; j < a.Bounds().Dy(); j++ {
			flatA[counter] = a.At(i, j)
			counter++
		}
	}

	return flatA
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

	distance := float64(dotProduct) / float64(magnitude)

	if distance < 0 {
		return 1 + (distance * -1)
	}

	return 1 - distance
}

func DTW(vecA, vecB []color.Color) ([][]int, []float64, [][]int, [][]int, [][]int) {
	dtw := make([][]float64, len(vecA)+1)

	for i := range len(vecA) + 1 {
		dtw[i] = make([]float64, len(vecB)+1)
		dtw[i][0] = math.Inf(0)
	}

	for j := range len(vecB) + 1 {
		dtw[0][j] = math.Inf(0)
	}

	dtw[0][0] = 0

	// log.Println("Matrix created")

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

	// log.Println("Matrix filled")
	// for i := range len(dtw) {
	// 	log.Println(dtw[len(dtw)-i-1])
	// }

	positions := make([][]int, 0)
	pathValues := make([]float64, 0)

	insertOps := make([][]int, 0)
	deleteOps := make([][]int, 0)
	matchOps := make([][]int, 0)

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
			opType    = "insertion"
		)

		smallest := insertion
		newI, newJ := i-1, j
		if deletion < smallest {
			opType = "deletion"
			smallest = deletion
			newI, newJ = i, j-1
		}
		if match < smallest {
			opType = "match"
			newI, newJ = i-1, j-1
		}

		i, j = newI, newJ

		if opType == "insertion" {
			insertOps = append(insertOps, []int{i, j})
		} else if opType == "deletion" {
			deleteOps = append(deleteOps, []int{i, j})
		} else if opType == "match" {
			matchOps = append(matchOps, []int{i, j})
		}
	}

	// log.Println("Positions and values done")

	return positions, pathValues, insertOps, deleteOps, matchOps
}
