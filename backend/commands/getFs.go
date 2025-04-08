package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
)

// returns the file system structure in json
func ParseGetfs(tokens []string) (string, error) {
	fmt.Println("GETFS")
	// no arguments are needed
	if len(tokens) != 0 {
		return "", errors.New("no se esperaban argumentos")
	}

	//jsonContent := "["
	for _, id := range stores.GetMountedPartitions() {
		fmt.Println("id: ", id)
		fmt.Println(stores.MountedPartitions[id])
	}

	return "", nil
}
