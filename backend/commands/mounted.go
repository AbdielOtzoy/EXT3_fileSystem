package commands

import (
	"errors"
	"backend/stores"
	"fmt"
	"strings"
)

// no fields are needed for this command
type MOUNTED struct {}

func ParseMounted(tokens []string) (string, error) {
	fmt.Println("MOUNTED")
	// no arguments are needed
	if len(tokens) != 0 {
		return "", errors.New("no se esperaban argumentos")
	}

	mountedPartitions := stores.GetMountedPartitions()

	if len(mountedPartitions) == 0 {
		return "No hay particiones montadas", nil
	}

	return "Particiones montadas:\n" + strings.Join(mountedPartitions, "\n"), nil
}

