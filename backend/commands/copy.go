package commands

import (
	"backend/stores"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type COPY struct {
	path    string
	destino string
}

func ParseCopy(tokens []string) (string, error) {
	cmd := &COPY{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-destino=[^\s]+`)
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
		case "-destino":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.destino = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.destino == "" {
		return "", errors.New("faltan parámetros requeridos: -destino")
	}

	err := commandCopy(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("COPY: Archivo %s copiado exitosamente.", cmd.path), nil
}

func commandCopy(cmd *COPY) error {
	parentDirs, destDir := utils.GetParentDirectories(cmd.path)
	destinoParentDirs, destinoDir := utils.GetParentDirectories(cmd.destino)
	username, idPartition, uid, gid := stores.GetSession()
	if username == "" || idPartition == "" {
		return errors.New("no hay sesión activa")
	}
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	err = partitionSuperblock.CopyFile(partitionPath, parentDirs, destDir, destinoParentDirs, destinoDir, uid, gid)
	if err != nil {
		return fmt.Errorf("error al copiar el archivo: %w", err)
	}

	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
