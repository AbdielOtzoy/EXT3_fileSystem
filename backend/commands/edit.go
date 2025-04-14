package commands

import (
	stores "backend/stores"
	"backend/structures"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type EDIT struct {
	path      string // path del archivo a editar
	contenido string // path del archivo con el contenido
}

func ParseEdit(tokens []string) (string, error) {
	cmd := &EDIT{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-contenido=[^\s]+`)
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
		case "-contenido":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.contenido = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.contenido == "" {
		return "", errors.New("faltan parámetros requeridos: -cont")
	}

	err := commandEdit(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("EDIT: Archivo %s editado correctamente.", cmd.path), nil // Devuelve el comando EDIT creado
}

func commandEdit(cmd *EDIT) error {
	// obtener la sesion
	username, idPartition, uid, gid := stores.GetSession()
	if username == "" || idPartition == "" {
		return errors.New("no hay sesión activa")
	}
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Editar el archivo
	err = editFile(cmd.path, cmd.contenido, partitionSuperblock, partitionPath, mountedPartition, uid, gid)
	if err != nil {
		return fmt.Errorf("error al editar el archivo: %w", err)
	}

	// Serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}

func editFile(path, content string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition, uid int32, gid int32) error {
	contentFile, err := utils.GetFileContent(content)
	parentDirs, destDir := utils.GetParentDirectories(path)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)
	if err != nil {
		return fmt.Errorf("error al obtener el contenido del archivo: %w", err)
	}

	if contentFile == "" {
		return fmt.Errorf("el archivo %s no existe", content)
	}

	err = sb.EditFile(partitionPath, parentDirs, destDir, contentFile, uid, gid)
	if err != nil {
		return fmt.Errorf("error al editar el archivo: %w", err)
	}

	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
