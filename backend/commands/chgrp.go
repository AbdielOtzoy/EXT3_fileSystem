package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type CHGRP struct {
	user  string
	grp string
}

func ParseChgrp(tokens []string) (string, error) {
	cmd := &CHGRP{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user=[^\s]+|-grp=[^\s]+`)
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
		case "-grp":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.grp = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.user == "" || cmd.grp == "" {
		return "", errors.New("faltan parámetros requeridos: -user y -grp")
	}

	return changeGroup(cmd.user, cmd.grp), nil
}

func changeGroup(user string, grp string) string {
	// get the current session
	userName, idPartition, _, _ := stores.GetSession()
	if userName == "" {
		return "No hay sesión activa"
	}

	// check if the user is root
	if userName != "root" {
		return "No tiene permisos para ejecutar este comando"
	}

	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "Error al obtener el superbloque de la partición: %w"
	}

	err = partitionSuperblock.ChangeGroup(user, grp, partitionPath)
	if err != nil {
		return fmt.Sprintf("Error al cambiar el grupo del usuario: %s", err)
	}
	// serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Sprintf("Error al serializar el superbloque: %s", err)
	}

	return fmt.Sprintf("Grupo del usuario %s cambiado a %s correctamente", user, grp)
}

