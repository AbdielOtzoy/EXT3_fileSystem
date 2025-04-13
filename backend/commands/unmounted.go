package commands

import (
	"backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type UNMOUNTED struct {
	id string
}

func ParseUnmounted(tokens []string) (string, error) {
	cmd := &UNMOUNTED{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	err := commandUnmount(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("unmounting partition %s", cmd.id), nil
}

func commandUnmount(unmounted *UNMOUNTED) error {
	fmt.Println("unmounting partition", unmounted.id)
	if stores.MountedPartitions[string(unmounted.id)] == "" {
		return errors.New("partition not mounted")
	}

	delete(stores.MountedPartitions, string(unmounted.id))
	return nil
}
