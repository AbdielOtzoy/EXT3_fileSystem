package commands

import (
	stores "backend/stores"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type FIND struct {
	path string
	name string
}

func ParseFIND(tokens []string) (string, error) {
	cmd := &FIND{} // create the mkdisk command

	args := strings.Join(tokens, " ") // join the tokens to get the arguments
	re := regexp.MustCompile(`-path=[^\s]+`)
	matches := re.FindAllString(args, -1) // find all the matches

	if len(matches) != len(tokens) {
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2) // split the match in key and value
		if len(kv) != 2 {
			return "", fmt.Errorf("parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-path":
			if value == "" {
				return "", errors.New("falta el path")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return "", errors.New("falta el nombre")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("falta el path")
	}

	if cmd.name == "" {
		return "", errors.New("falta el nombre")
	}

	err := commandFind(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Se han encontrado los archivos: %s", cmd.name), nil
}

func commandFind(cmd *FIND) error {
	// obtener la sesion
	username, idPartition, uid, gid := stores.GetSession()
	if username == "" || idPartition == "" || uid == 0 || gid == 0 {
		return errors.New("no hay sesión activa")
	}

	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	parentsDir, destDir := utils.GetParentDirectories(cmd.path)

	// Buscar archivos
	files, err := partitionSuperblock.Find(partitionPath, parentsDir, destDir, cmd.name)
	if err != nil {
		return fmt.Errorf("error al buscar archivos: %w", err)
	}

	if len(files) == 0 {
		return errors.New("no se encontraron archivos")
	}

	return nil
}
