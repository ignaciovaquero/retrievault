package permissions

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

// StringToFileMode converts a string in the form "0644" or "644" to a
// os.FileMode
func StringToFileMode(str string) (os.FileMode, error) {
	if len(str) > 4 {
		return 0, fmt.Errorf("Invalid string %s", str)
	}
	first, err := strconv.Atoi(str[len(str)-3 : len(str)-2])
	if err != nil {
		return 0, err
	}
	second, err := strconv.Atoi(str[len(str)-2 : len(str)-1])
	if err != nil {
		return 0, err
	}
	third, err := strconv.Atoi(str[len(str)-1:])
	if err != nil {
		return 0, err
	}
	decimalPerm := float64(first)*math.Pow(8, 2) + float64(second)*math.Pow(8, 2) + float64(third)*math.Pow(8, 0)
	return os.FileMode(decimalPerm), nil
}
