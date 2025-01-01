package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// LoadImage loads an image from a file.
func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// ExtractVerticalLine extracts the pixel intensities of a specific column from an image.
func ExtractVerticalLine(img image.Image, col int) []float64 {
	bounds := img.Bounds()
	height := bounds.Max.Y
	line := make([]float64, height)

	for y := 0; y < height; y++ {
		r, g, b, _ := img.At(col, y).RGBA()
		gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
		line[y] = gray / 65535.0 // Normalize to [0, 1]
	}
	return line
}

// CosineSimilarity calculates the cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// FindTopSimilarLines compares a target line to all other vertical lines in the dataset.
func FindTopSimilarLines(target []float64, images []string, col int) []struct {
	ImagePath  string
	Similarity float64
} {
	results := []struct {
		ImagePath  string
		Similarity float64
	}{}

	for _, path := range images {
		img, err := LoadImage(path)
		if err != nil {
			fmt.Printf("Error loading image %s: %v\n", path, err)
			continue
		}
		line := ExtractVerticalLine(img, col)
		similarity := CosineSimilarity(target, line)
		results = append(results, struct {
			ImagePath  string
			Similarity float64
		}{path, similarity})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > 10 {
		results = results[:10]
	}
	return results
}

func main() {
	// Specify the directory containing images and the target image.
	imageDir := "./images"
	targetImagePath := "./target.jpg"
	targetColumn := 50 // Column index to compare.

	// Load the target image and extract the vertical line.
	targetImg, err := LoadImage(targetImagePath)
	if err != nil {
		fmt.Printf("Error loading target image: %v\n", err)
		return
	}
	targetLine := ExtractVerticalLine(targetImg, targetColumn)

	// Load all images from the directory.
	var imagePaths []string
	filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".jpg" {
			imagePaths = append(imagePaths, path)
		}
		return nil
	})

	// Find top-10 similar lines using cosine similarity.
	topSimilar := FindTopSimilarLines(targetLine, imagePaths, targetColumn)
	fmt.Println("Top-10 Most Similar Lines (Cosine Similarity):")
	for _, result := range topSimilar {
		fmt.Printf("Image: %s, Similarity: %.4f\n", result.ImagePath, result.Similarity)
	}
}
