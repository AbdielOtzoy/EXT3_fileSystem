package commands

import (
	stores "backend/stores"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type CHOWN struct {
	path    string
	r       bool
	usuario string
}

func ParseCHOWN(tokens []string) (string, error) {
	cmd := &CHOWN{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-r|-usuario=[^\s]+`)
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
		case "-usuario":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.usuario = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	if cmd.usuario == "" {
		return "", errors.New("faltan parámetros requeridos: -usuario")
	}

	err := commandCHOWN(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("CHOWN: Cambios realizados correctamente\n"+
		"-> Path: %s\n"+
		"-> Usuario: %s\n"+
		"-> Recursivo: %t\n",
		cmd.path,
		cmd.usuario,
		cmd.r), nil
}

func commandCHOWN(cmd *CHOWN) error {
	// obtener la sesion
	username, idPartition, _, _ := stores.GetSession()
	if username == "" || idPartition == "" {
		return errors.New("no hay sesión activa")
	}
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	uid, gid, err := partitionSuperblock.GetUidGidByName(cmd.usuario, partitionPath)
	if err != nil {
		return fmt.Errorf("error al obtener el uid y gid: %w", err)
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

			err = partitionSuperblock.ChownInInode(partitionPath, 0, parentsDir2, destDir2, uid, gid)
			if err != nil {
				return fmt.Errorf("error al cambiar el dueño del archivo: %w", err)
			}

			err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
			if err != nil {
				return fmt.Errorf("error al serializar el superbloque: %w", err)
			}
		}
	}

	parentsDir, destDir := utils.GetParentDirectories(cmd.path)
	// Cambiar el dueño del archivo
	err = partitionSuperblock.Chown(partitionPath, parentsDir, destDir, uid, gid)
	if err != nil {
		return fmt.Errorf("error al cambiar el dueño del archivo: %w", err)
	}

	// Serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
