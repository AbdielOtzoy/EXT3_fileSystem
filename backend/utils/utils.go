package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// convert a size and unit to bytes
func ConvertToBytes(size int, unit string) (int, error) {
	switch unit {
	case "B":
		return size, nil
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	case "G":
		return size * 1024 * 1024 * 1024, nil
	case "T":
		return size * 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, errors.New("invalid unit")
	}
}



// converting a string to an int
func StringToInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}

// alphabet array
var alphabet =	[]string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// map to store the assignment of letters to the different paths
var pathToLetter = make(map[string]string)

var nextLetterIndex = 0

// get the letter assigned to a path
func GetLetter(path string) (string, error) {
	// Assign a letter to the path if it does not have one
	if _, exists := pathToLetter[path]; !exists {
		if nextLetterIndex < len(alphabet) {
			pathToLetter[path] = alphabet[nextLetterIndex]
			nextLetterIndex++
		} else {
			fmt.Println("Error: there are no more letters available to assign")
			return "", errors.New("no more letters available")
		}
	}

	return pathToLetter[path], nil
}

// createParentDirs crea las carpetas padre si no existen
func CreateParentDirs(path string) error {
	dir := filepath.Dir(path)
	// os.MkdirAll no sobrescribe las carpetas existentes, solo crea las que no existen
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error al crear las carpetas padre: %v", err)
	}
	return nil
}

// getFileNames obtiene el nombre del archivo .dot y el nombre de la imagen de salida
func GetFileNames(path string) (string, string) {
	dir := filepath.Dir(path)
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	dotFileName := filepath.Join(dir, baseName+".dot")
	outputImage := path
	return dotFileName, outputImage
}

func GetParentDirectories(path string) ([]string, string) {
	path = filepath.Clean(path)

	components := strings.Split(path, string(filepath.Separator))

	var parentDirs []string

	for i := 1; i < len(components)-1; i++ {
		parentDirs = append(parentDirs, components[i])
	}
	destDir := components[len(components)-1]

	return parentDirs, destDir
}

func First[T any](slice []T) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, errors.New("el slice está vacío")
	}
	return slice[0], nil
}

func RemoveElement[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice 
	}
	return append(slice[:index], slice[index+1:]...)
}

// splitStringIntoChunks divide una cadena en partes de tamaño chunkSize y las almacena en una lista
func SplitStringIntoChunks(s string) []string {
	var chunks []string
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func GetFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error al leer el archivo: %w", err)
	}
	contentStr := string(content)
	return contentStr, nil

}

func SplitPath(path string) []string {
	components := strings.Split(path, "/")

	var result []string

	currentPath := ""
	for _, component := range components {
		if component == "" {
			continue
		}
		currentPath += "/" + component
		result = append(result, currentPath)
	}
	// retornar pero sin el ultimo elemento ya que solo se verifica que existan las carpetas padre
	return result[:len(result)-1]
}
 // Función que recibe la ruta original y el newPath, y devuelve la nueva ruta
func ReplacePath(originalPath, newPath string) string {
	// Obtener el directorio de la ruta original
	dir := filepath.Dir(originalPath)
	
	// Combinar el directorio con el newPath
	newFullPath := filepath.Join(dir, newPath)
	
	return newFullPath
}

// formatDate
// 2025-03-26T09:19:28-06:00 to 2025-03-26 09:19:28
func FormatDate(date string) (string, string) {
	//devuelve la fecha y la hora
	parts := strings.Split(date, "T")
	if len(parts) != 2 {
		return "", ""
	}

	datePart := parts[0]
	timePart := strings.Split(parts[1], "-")[0]
	return datePart, timePart
}

func StringToInt32(s string) int32 {
	var result int32
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int32(char-'0')
		} else {
			return 0
		}
	}
	return result
}