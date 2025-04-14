package commands

import (
	stores "backend/stores"
	structures "backend/structures"
	utils "backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type REMOVE struct {
	path string
}

func ParseRemove(tokens []string) (string, error) {
	cmd := &REMOVE{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+`)
	matches := re.FindAllString(args, -1)

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
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	err := commandRemove(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("REMOVE: Directorio %s eliminado correctamente.", cmd.path), nil // Devuelve el comando REMOVE creado
}

func commandRemove(cmd *REMOVE) error {
	// obtener la sesión activa
	username, idPartitinUser, _, _ := stores.GetSession()
	if username == "" {
		return errors.New("no hay sesión activa")
	}
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartitinUser)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Crear el directorio
	err = remove(cmd.path, partitionSuperblock, partitionPath, mountedPartition)
	if err != nil {
		err = fmt.Errorf("error al crear el directorio: %w", err)
	}

	return err
}

func remove(path string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition) error {
	parentDirs, destDir := utils.GetParentDirectories(path)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)
	// elimina el archivo o carpeta
	err := sb.Delete(partitionPath, parentDirs, destDir)
	if err != nil {
		return fmt.Errorf("error al eliminar el archivo o carpeta: %w", err)
	}

	// Serializar el superbloque
	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
