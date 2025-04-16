package commands

import (
	stores "backend/stores"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type CHMOD struct {
	path string
	ugo  string
	r    bool
}

func ParseCHMOD(tokens []string) (string, error) {
	cmd := &CHMOD{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-r|-ugo=[^\s]+`)
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
		case "-r":
			cmd.r = true
		case "-ugo":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.ugo = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.ugo == "" {
		return "", errors.New("faltan parámetros requeridos: -ugo")
	}

	err := commandCHMOD(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("CHMOD: Cambios realizados correctamente\n"+
		"-> Path: %s\n"+
		"-> UGO: %s\n"+
		"-> Recursivo: %t\n",
		cmd.path,
		cmd.ugo,
		cmd.r), nil
}

func commandCHMOD(cmd *CHMOD) error {
	// obtener la sesion
	username, idPartition, uid, gid := stores.GetSession()
	if username == "" || idPartition == "" || uid == 0 || gid == 0 {
		return errors.New("no hay sesión activa")
	}

	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	if cmd.r {
		fmt.Println("Opción -r activada")
		pathSplited := utils.SplitPath(cmd.path)
		for _, path := range pathSplited {
			fmt.Println("path:", path)
			parentsDir2, destDir2 := utils.GetParentDirectories(path)
			exists, err := partitionSuperblock.ExistsFolcer(partitionPath, parentsDir2, destDir2)
			if err != nil {
				return fmt.Errorf("error al verificar si la carpeta existe: %w", err)
			}
			if !exists {
				fmt.Printf("La carpeta %s no existe\n", path)
				return fmt.Errorf("la carpeta %s no existe", path)
			}

			err = partitionSuperblock.ChmodInInode(partitionPath, 0, parentsDir2, destDir2, cmd.ugo, uid, gid)
			if err != nil {
				return fmt.Errorf("error al cambiar los permisos: %w", err)
			}

			err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
			if err != nil {
				return fmt.Errorf("error al serializar el superbloque: %w", err)
			}
		}
	}

	parentsDir, destDir := utils.GetParentDirectories(cmd.path)

	// Cambiar los permisos
	err = partitionSuperblock.Chmod(partitionPath, parentsDir, destDir, cmd.ugo, uid, gid)
	if err != nil {
		return fmt.Errorf("error al cambiar los permisos: %w", err)
	}

	// Serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
