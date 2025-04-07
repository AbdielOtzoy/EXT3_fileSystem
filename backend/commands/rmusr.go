package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type RMUSR struct {
	user  string
}

func ParseRmuser(tokens []string) (string, error) {
	cmd := &RMUSR{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user=[^\s]+`)
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
		case "-user":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.user = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.user == "" {
		return "", errors.New("faltan parámetros requeridos: -user")
	}

	return removeUser(cmd.user), nil
}

func removeUser(user string) string {
	// get the current session
	userName, idPartition, _, _  := stores.GetSession()
	if userName == "" {
		return "No hay sesión activa"
	}

	// verificar que sea el usuario root
	if userName != "root" {
		return "No tiene permisos para eliminar usuarios"
	}

	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "Error al obtener el superbloque de la partición: %w"
	}

	err = partitionSuperblock.RemoveUser(user, partitionPath)
	if err != nil {
		return fmt.Sprintf("Error al eliminar el usuario: %s", err)
	}

	// serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Sprintf("Error al serializar el superbloque: %s", err)
	}

	return fmt.Sprintf("Usuario %s eliminado correctamente", user)

}

