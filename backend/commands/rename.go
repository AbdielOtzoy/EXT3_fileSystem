package commands

import (
	"backend/stores"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type RENAME struct {
	path string
	name string
}

func ParseRename(tokens []string) (string, error) {
	cmd := &RENAME{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-name=[^\s]+`)
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
		case "-path":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.path = value
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

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	err := commandRename(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("RENAME: Archivo %s renombrado exitosamente.", cmd.path), nil
}

func commandRename(cmd *RENAME) error {
	parentDirs, destDir := utils.GetParentDirectories(cmd.path)
	username, idPartition, uid, gid := stores.GetSession()
	if username == "" || idPartition == "" {
		return errors.New("no hay sesión activa")
	}
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	err = partitionSuperblock.RenameFile(partitionPath, parentDirs, destDir, cmd.name, uid, gid)
	if err != nil {
		return fmt.Errorf("error al renombrar el archivo: %w", err)
	}

	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
