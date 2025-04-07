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

type MKFILE struct {
	path 	string // Ruta del archivo a crear 
	r   	bool   // indica si se deben crear las carpetas intermedias si no existen
	size 	int    // Tamaño del archivo a crear
	cont 	string // ruta de un archivo a copiar su contenido en el nuevo archivo
}

/*
   mkfile -size=15 -path=/home/user/docs/a.txt -r
   mkfile -path="/home/mis documentos/archivo 1.txt"
   mkfile -path=/home/user/docs/b.txt -r -cont=/home/Documents/b.txt
*/

func ParseMKfile(tokens []string) (string, error) {
	cmd := &MKFILE{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-size=\d+|-r|-cont=[^\s]+`)
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
		case "-size":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			cmd.size = utils.StringToInt(kv[1])
			if cmd.size <= 0 {
				return "", fmt.Errorf("el tamaño debe ser mayor a 0: %s", match)
			}
		case "-cont":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.cont = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// imprimir los parámetros
	fmt.Println("Parámetros:")
	fmt.Println("Path:", cmd.path)
	fmt.Println("Recursivo:", cmd.r)
	fmt.Println("Tamaño:", cmd.size)
	fmt.Println("Contenido:", cmd.cont)
	fmt.Println("")

	

	err := commandMkfile(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKFILE: Archivo %s creado correctamente.", cmd.path), nil // Devuelve el comando MKDIR creado
}

func commandMkfile(mkfile *MKFILE) error {
	// obtener la sesion
	username, idPartition, uid, gid := stores.GetSession()
	if username == "" || idPartition == "" {
		return errors.New("no hay sesión activa")
	}
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Crear el archivo
	err = createFile(mkfile.path, mkfile.r, mkfile.size, mkfile.cont, partitionSuperblock, partitionPath, mountedPartition, uid, gid)
	if err != nil {
		err = fmt.Errorf("error al crear el archivo: %w", err)
	}

	return err
}

func createFile(dirPath string, r bool, size int, contentPath string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition, uid int32, gid int32) error {
	fmt.Println("\nCreando directorio:", dirPath)
	fmt.Println("Recursivo:", r)
	fmt.Println("Tamaño:", size)
	fmt.Println("path del contenido:", contentPath)

	parentDirs, destDir := utils.GetParentDirectories(dirPath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)
	
	if r {
		fmt.Println("Opción -r activada")
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

			fmt.Printf("Creando la carpeta %s\n", path)
			err = createDirectory(path, false, sb, partitionPath, mountedPartition, uid, gid)
			if err != nil {
				return fmt.Errorf("error al crear la carpeta: %w", err)
			}

			// Serializar el superbloque
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

	contentFile := ""
	contentFile, err := utils.GetFileContent(contentPath)
	if err != nil {
		contentFile = ""
	}
	if contentFile != "" {
		fmt.Println("Contenido del archivo:", contentFile)
	} else {
		fmt.Println("No se encontró el archivo de contenido.")
	}

	fmt.Println("CONTENTFILE", contentFile)
	// Crear el directorio segun el path proporcionado
	err = sb.CreateFile(partitionPath, parentDirs, destDir, r, size, contentFile, uid, gid)
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