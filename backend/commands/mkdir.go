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

type MKDIR struct {
	path string 
	p    bool  
}

/*
   mkdir -p -path=/home/user/docs/usac
   mkdir -path="/home/mis documentos/archivos clases"
*/

func ParseMkdir(tokens []string) (string, error) {
	cmd := &MKDIR{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-p`)
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
		case "-p":
			cmd.p = true
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	err := commandMkdir(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKDIR: Directorio %s creado correctamente.", cmd.path), nil // Devuelve el comando MKDIR creado
}

func commandMkdir(mkdir *MKDIR) error {
	// obtener la sesión activa
	username, idPartitinUser, uid, gid := stores.GetSession()
	if username == "" {
		return errors.New("no hay sesión activa")
	}
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartitinUser)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Crear el directorio
	err = createDirectory(mkdir.path, mkdir.p, partitionSuperblock, partitionPath, mountedPartition, uid, gid)
	if err != nil {
		err = fmt.Errorf("error al crear el directorio: %w", err)
	}

	return err
}

func createDirectory(dirPath string, p bool, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition, uid int32, gid int32) error {

	parentDirs, destDir := utils.GetParentDirectories(dirPath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	if p {
		fmt.Println("Opción -p activada")
		pathSplited := utils.SplitPath(dirPath)
		for _, path := range pathSplited {
			fmt.Println("path:", path)
			// verificar que cada carpeta padre exista
			parentsDir, destDir := utils.GetParentDirectories(path)
			fmt.Printf("Directorios padres: %v\n", parentsDir)
			fmt.Printf("Directorio destino: %s\n", destDir)
			exists, err := sb.ExistsFolcer(partitionPath, parentsDir, destDir)
			if err != nil {
				return fmt.Errorf("error al verificar si la carpeta existe: %w", err)
			}

			if exists {
				fmt.Printf("La carpeta %s ya existe\n", path)
				continue
			}

			fmt.Printf("la carpeta no exisxte, creandola: %s\n", path)
			err = createDirectory(path, false, sb, partitionPath, mountedPartition, uid, gid)
			if err != nil {
				return fmt.Errorf("error al crear la carpeta: %w", err)
			}

			err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
			if err != nil {
				return fmt.Errorf("error al serializar el superbloque: %w", err)
			}
			fmt.Println("Carpeta creada correctamente")
		}
	} else {
		// verificar que los directorios padres existan
		pathSplited := utils.SplitPath(dirPath)
		for _, path := range pathSplited {
			fmt.Println("path:", path)
			// verificar que cada carpeta padre exista
			parentsDir, destDir := utils.GetParentDirectories(path)
			fmt.Printf("Directorios padres: %v\n", parentsDir)
			fmt.Printf("Directorio destino: %s\n", destDir)
			exists, _ := sb.ExistsFolcer(partitionPath, parentsDir, destDir)
			fmt.Println("exists", exists)
			if !exists {
				fmt.Printf("La carpeta %s no existe\n", path)
				return fmt.Errorf("la carpeta %s no existe", path)
			}

			if exists {
				fmt.Printf("La carpeta %s ya existe\n", path)
				continue
			}
		}
	}

	// Crear el directorio segun el path proporcionado
	err := sb.CreateFolder(partitionPath, parentDirs, destDir, uid, gid)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	// Serializar el superbloque
	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}