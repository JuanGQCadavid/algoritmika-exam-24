package utls

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// func main() {
// 	// Data to write
// 	intData := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
// 	floatData := []float64{1.1, 2.2, 3.3, 4.4, 5.5}

// 	// File name
// 	fileName := "data.txt"

// 	// Write data to file
// 	err := writeDataToFile(fileName, intData, floatData)
// 	if err != nil {
// 		fmt.Println("Error writing to file:", err)
// 		return
// 	}

// 	// Read data back from file
// 	readIntData, readFloatData, err := readDataFromFile(fileName)
// 	if err != nil {
// 		fmt.Println("Error reading from file:", err)
// 		return
// 	}

// 	// Print the read data
// 	fmt.Println("Read Integer Data:", readIntData)
// 	fmt.Println("Read Float Data:", readFloatData)
// }

func WriteDataToFile(fileName string, intData [][]int, floatData []float64) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write integer data
	for _, row := range intData {
		line := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(row)), " "), "[]")
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	// Write float data
	for _, value := range floatData {
		_, err := writer.WriteString(fmt.Sprintf("%f\n", value))
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func ReadDataFromFile(fileName string) ([][]int, []float64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var intData [][]int
	var floatData []float64

	scanner := bufio.NewScanner(file)
	readingInts := true

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ".") { // Detect float data
			readingInts = false
		}

		if readingInts {
			// Parse integers
			strValues := strings.Fields(line)
			row := make([]int, len(strValues))
			for i, str := range strValues {
				val, err := strconv.Atoi(str)
				if err != nil {
					return nil, nil, err
				}
				row[i] = val
			}
			intData = append(intData, row)
		} else {
			// Parse floats
			val, err := strconv.ParseFloat(line, 64)
			if err != nil {
				return nil, nil, err
			}
			floatData = append(floatData, val)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return intData, floatData, nil
}
