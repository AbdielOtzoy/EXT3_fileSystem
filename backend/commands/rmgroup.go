package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type RMGROUP struct {
	name string
}

func ParseRmgroup(tokens []string) (string, error) {
	cmd := &RMGROUP{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	if len(matches) != len(tokens) {
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])

		switch key {
		case "-name":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	return removeGroup(cmd.name), nil
}

func removeGroup(name string) string {
	// get the current session
	userName, idPartition, _, _  := stores.GetSession()
	if userName == "" {
		return "No hay sesión activa"
	}

	// confirmar que el usuario es el usuario root
	if userName != "root" {
		return "No tiene permisos para crear grupos"
	}

	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "Error al obtener el superbloque de la partición: %w"
	}

	err = partitionSuperblock.RemoveGroup(name, partitionPath)

	if err != nil {
		return fmt.Sprintf("Error al eliminar el grupo: %s", err)
	}

	// serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Sprintf("Error al serializar el superbloque: %s", err)
	}

	return fmt.Sprintf("Grupo %s eliminado correctamente", name)
}