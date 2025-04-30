package commands

import (
	"backend/stores"
	"backend/structures"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type RECOVERY struct {
	id string
}

func ParseRecovery(tokens []string) (string, error) {
	cmd := &RECOVERY{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

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

	err := commandRecovery(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("recovery partition %s", cmd.id), nil
}

// commandRecovery es una función ficticia que simula la recuperación de una partición
func commandRecovery(cmd *RECOVERY) error {
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(cmd.id)
	if err != nil {
		return fmt.Errorf("error al obtener el superbloque de la partición: %w", err)
	}

	if partitionSuperblock.S_filesystem_type != 3 {
		return errors.New("sistema de archivos no es ext3")
	}

	count := 2
	// Iniciar el sistmea de archivos
	error := partitionSuperblock.CreateUsersFileExt3(partitionPath, -1)
	if error != nil {
		return fmt.Errorf("error al crear el sistema de archivos: %w", error)
	}
	for {
		// obtener el journal
		journal := &structures.Journal{}

		fmt.Println("Deserializando en:", int64(mountedPartition.Part_start)+68+114*int64(count))
		// Deserializar el journal
		err = journal.Deserialize(partitionPath, int64(mountedPartition.Part_start)+68+114*int64(count))
		if err != nil {
			return fmt.Errorf("error al deserializar el journal: %w", err)
		}
		if journal.J_count == 0 {
			break
		}
		fmt.Println("obtenido el journal")
		journal.Print()

		operation := strings.TrimRight(string(journal.J_content.I_operation[:]), "\x00")
		path := strings.TrimRight(string(journal.J_content.I_path[:]), "\x00")
		content := strings.TrimRight(string(journal.J_content.I_content[:]), "\x00")

		fmt.Println("Operación:", operation, "Ruta:", path, "Contenido:", content)

		parentDirs, destiDir := utils.GetParentDirectories(path)
		size := utils.StringToInt(content)
		if operation == "mkdir" {
			err := partitionSuperblock.CreateFolder(partitionPath, parentDirs, destiDir, 0, 0, path, -1)
			if err != nil {
				return fmt.Errorf("error al crear el directorio: %w", err)
			}
			fmt.Println("MKDIR: Directorio creado correctamente")
		} else if operation == "mkfile" {
			err := partitionSuperblock.CreateFile(partitionPath, parentDirs, destiDir, false, size, content, 0, 0, path, -1)
			if err != nil {
				return fmt.Errorf("error al crear el archivo: %w", err)
			}
			fmt.Println("MKFILE: Archivo creado correctamente")
		} else {
			fmt.Println("Operación desconocida detectada")
		}

		count++
	}

	// Serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
