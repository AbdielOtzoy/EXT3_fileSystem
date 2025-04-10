package structures

import (
	utils "backend/utils"
	"fmt"
	"strings"
	"time"
)

// Crear users.txt en nuestro sistema de archivos
func (sb *SuperBlock) CreateUsersFileExt2(path string) error {
	// ----------- Creamos / -----------
	// Creamos el inodo raíz
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err := rootInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque del Inodo Raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Deserializar el inodo raíz
	err = rootInode.Deserialize(path, int64(sb.S_inode_start+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Actualizamos el inodo raíz
	rootInode.I_atime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = rootInode.Serialize(path, int64(sb.S_inode_start+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Deserializar el bloque de carpeta raíz
	err = rootBlock.Deserialize(path, int64(sb.S_block_start+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Actualizamos el bloque de carpeta raíz
	rootBlock.B_content[2] = FolderContent{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_block_start+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Creamos el inodo users.txt
	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	// Serializar el inodo users.txt
	err = usersInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque de users.txt
	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}
	// Copiamos el texto de usuarios en el bloque
	copy(usersBlock.B_content[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	return nil
}

func (sb *SuperBlock) createFileInodeExt2(path string, inodeIndex int32, parentsDir []string, destDir string, r bool, size int, contentFile string, uid int32, gid int32) error {
	// crear un nuevo inodo
	inode := &Inode{}
	// deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}

	// verificar que el inodo sea de tipo carpeta
	if inode.I_type[0] == '1' {
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			fmt.Println("No hay más bloques, creando uno nuevo")
			fmt.Println("i:", i)

			// 1. Crear y serializar un nuevo bloque de carpeta
			newBlock := &FolderBlock{
				B_content: [4]FolderContent{
					{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
					{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
					{B_name: [12]byte{'-'}, B_inodo: -1},
					{B_name: [12]byte{'-'}, B_inodo: -1},
				},
			}

			// actualizar contenido del bloque
			destDirByte := [12]byte{}
			copy(destDirByte[:], destDir)
			newBlock.B_content[2] = FolderContent{B_name: destDirByte, B_inodo: sb.S_inodes_count}

			newBlockPos := sb.S_blocks_count

			// Serializar el nuevo bloque
			err = newBlock.Serialize(path, int64(sb.S_block_start+(newBlockPos*sb.S_block_size)))
			if err != nil {
				return err
			}

			// Actualizar bitmap de bloques
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}

			// Actualizar superbloque (nuevo bloque)
			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size

			// 2. Crear el inodo del nuevo archivo
			folderInode := &Inode{
				I_uid:   uid,
				I_gid:   gid,
				I_size:  0,
				I_atime: float32(time.Now().Unix()),
				I_ctime: float32(time.Now().Unix()),
				I_mtime: float32(time.Now().Unix()),
				I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
				I_type:  [1]byte{'1'},
				I_perm:  [3]byte{'6', '6', '4'},
			}

			content := ""
			if contentFile != "" {
				content = contentFile
			} else {
				// llenar el contenido con una cadena de numeros del 0 al 9 cuantas veces sea el tamaño
				for i := 0; i < size; i++ {
					content += string(i%10 + '0')
				}
			}
			fmt.Println("Contenido del archivo: ", content)

			// crear un arreglo de bloques que almacene el contenido de 64 bytes
			contentBlocks := make([]string, 0)
			// dividir el contenido en bloques de 64 bytes
			for i := 0; i < len(content); i += 64 {
				if i+64 > len(content) {
					contentBlocks = append(contentBlocks, content[i:])
				} else {
					contentBlocks = append(contentBlocks, content[i:i+64])
				}
			}
			// iterar sobre cada bloque del file inode y asignar el contenido
			for i := 0; i < len(contentBlocks); i++ {
				// crear un nuevo bloque de archivo
				fileBlock := &FileBlock{
					B_content: [64]byte{},
				}
				// copiar el contenido del bloque
				copy(fileBlock.B_content[:], contentBlocks[i])

				folderInode.I_block[i] = sb.S_blocks_count

				// serializar el bloque
				err = fileBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
				if err != nil {
					return err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				// Actualizar el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

			}
			if len(contentBlocks) == 0 {
				// crear un bloque de archivo vacío
				fileBlock := &FileBlock{
					B_content: [64]byte{},
				}

				folderInode.I_block[0] = sb.S_blocks_count
				// serializar el bloque
				err = fileBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
				if err != nil {
					return err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				// Actualizar el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size
			}
			// Serializar el inodo del archivo
			err = folderInode.Serialize(path, int64(sb.S_first_ino))
			if err != nil {
				return err
			}

			// Actualizar el bitmap de inodos
			err = sb.UpdateBitmapInode(path)
			if err != nil {
				return err
			}

			// Actualizar el superbloque
			sb.S_inodes_count++
			sb.S_free_inodes_count--
			sb.S_first_ino += sb.S_inode_size
			fmt.Println("contenido del archivo para añadir al bloque: ", contentFile)

			inode.I_block[i] = newBlockPos

			// Serializar el inodo actualizado
			err = inode.Serialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
			if err != nil {
				return err
			}

			return nil
		}

		block := &FolderBlock{}
		// deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			// Si las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					break
				}

				// Obtenemos la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					err := sb.createFileInodeExt2(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, r, size, contentFile, uid, gid)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				fmt.Println("---------ESTOY  CREANDO--------")
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if content.B_inodo != -1 {
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count
				// Actualizar el bloque
				block.B_content[indexContent] = content
				// Serializar el bloque
				err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}

				// Crear el inodo del archivo
				// Crear el inodo de la carpeta
				fileInode := &Inode{
					I_uid:   uid,
					I_gid:   gid,
					I_size:  int32(size),
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'1'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				content := ""
				if contentFile != "" {
					content = contentFile
				} else {
					// llenar el contenido con una cadena de numeros del 0 al 9 cuantas veces sea el tamaño
					for i := 0; i < size; i++ {
						content += string(i%10 + '0')
					}
				}
				fmt.Println("Contenido del archivo: ", content)

				// crear un arreglo de bloques que almacene el contenido de 64 bytes
				contentBlocks := make([]string, 0)
				// dividir el contenido en bloques de 64 bytes
				for i := 0; i < len(content); i += 64 {
					if i+64 > len(content) {
						contentBlocks = append(contentBlocks, content[i:])
					} else {
						contentBlocks = append(contentBlocks, content[i:i+64])
					}
				}

				// iterar sobre cada bloque del file inode y asignar el contenido
				for i := 0; i < len(contentBlocks); i++ {
					// crear un nuevo bloque de archivo
					fileBlock := &FileBlock{
						B_content: [64]byte{},
					}
					// copiar el contenido del bloque
					copy(fileBlock.B_content[:], contentBlocks[i])

					fileInode.I_block[i] = sb.S_blocks_count

					// serializar el bloque
					err = fileBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
					if err != nil {
						return err
					}

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(path)
					if err != nil {
						return err
					}

					// Actualizar el superbloque
					sb.S_blocks_count++
					sb.S_free_blocks_count--
					sb.S_first_blo += sb.S_block_size

				}
				if len(contentBlocks) == 0 {
					// crear un bloque de archivo vacío
					fileBlock := &FileBlock{
						B_content: [64]byte{},
					}

					fileInode.I_block[0] = sb.S_blocks_count
					// serializar el bloque
					err = fileBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
					if err != nil {
						return err
					}

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(path)
					if err != nil {
						return err
					}

					// Actualizar el superbloque
					sb.S_blocks_count++
					sb.S_free_blocks_count--
					sb.S_first_blo += sb.S_block_size
				}
				// Serializar el inodo del archivo
				err = fileInode.Serialize(path, int64(sb.S_first_ino))
				if err != nil {
					return err
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return err
				}

				// Actualizar el superbloque
				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size
				fmt.Println("contenido del archivo para añadir al bloque: ", contentFile)

				return nil
			}
		}
	}
	return nil
}

func (sb *SuperBlock) folderExists(path string, inodeIndex int32, parentsDir []string, destDir string) (bool, error) {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return false, err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return false, nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return false, err
		}
		block.Print()

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					break
				}
				// Obtenemos la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return false, err
				}

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")

				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					fmt.Println("entrando a la carpeta padre")
					fileContent, err := sb.folderExists(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return false, err
					}
					return fileContent, nil
				}
			} else {

				if content.B_inodo == -1 {
					continue
				}
				fmt.Println("ya no hay más carpetas padre")
				fmt.Println("Carpeta destino: ", destDir)

				// convertir content.B_name a string
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				fmt.Println("Nombre de la carpeta: ", contentName)
				if contentName == destDir {
					fmt.Println("---------LA ENCONTRÉ-------")
					return true, nil
				}
			}
		}

	}
	return false, nil
}

func (sb *SuperBlock) readFileInInode(path string, inodeIndex int32, parentsDir []string, destDir string) (string, error) {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return "", err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return "", nil
	}

	fmt.Println("destDir: ", destDir)
	fmt.Println("Padres: ", parentsDir)

	fmt.Println("Inodo: ", inodeIndex)

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return "", err
		}

		fmt.Println("Bloque: ", blockIndex)
		block.Print()

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			if len(parentsDir) != 0 {
				fmt.Println("---------ESTOY  VISITANDO--------")

				if content.B_inodo == -1 {
					break
				}
				// Obtenemos la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return "", err
				}

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")

				fmt.Println("Nombre del contenido: ", contentName)
				fmt.Println("Nombre del archivo: ", destDir)
				fmt.Println("Nombre de la carpeta padre: ", parentDirName)

				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					fmt.Println("entrando a la carpeta padre")
					fileContent, err := sb.readFileInInode(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return "", err
					}
					if fileContent != "" {
						return fileContent, nil
					}
					return "", fmt.Errorf("no se encontró el archivo")
				}
			} else {
				fmt.Println("ya no hay más carpetas padre")
				if content.B_inodo == -1 {
					continue
				}
				// convertir destDir a 12 bytes
				destDirByte := [12]byte{}
				copy(destDirByte[:], destDir)

				if content.B_name == destDirByte {
					fmt.Println("---------LA ENCONTRÉ-------")
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					inodeFile := &Inode{}
					err := inodeFile.Deserialize(path, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
					if err != nil {
						return "", err
					}

					// Verificar si el inodo es de tipo archivo
					if inodeFile.I_type[0] == '1' {
						allContent := ""
						// Iterar sobre cada bloque del inodo (apuntadores)
						for _, blockIndex := range inodeFile.I_block {
							// Si el bloque no existe, salir
							if blockIndex == -1 {
								break
							}

							// Crear un nuevo bloque de archivo
							blockFile := &FileBlock{}

							// Deserializar el bloque
							err := blockFile.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
							if err != nil {
								return "", err
							}
							allContent += string(blockFile.B_content[:])
						}

						return allContent, nil

					}
					return "", fmt.Errorf("el inodo no es de tipo archivo")
				}
			}
		}
	}
	return "", fmt.Errorf("no se encontró el archivo")
}

// getInodeFromPath busca y devuelve el inodo correspondiente a una ruta específica
func (sb *SuperBlock) getInodeFromPath(path string, inodeIndex int32, parentsDir []string, targetName string) (int32, error) {
	// Deserializar el inodo actual
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return -1, err
	}

	// Si es un archivo y no hay más padres, verificar si es el target
	if inode.I_type[0] == '1' && len(parentsDir) == 0 {
		return inodeIndex, nil
	}

	// Iterar sobre cada bloque del inodo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}

		// Si es un directorio, buscar en los bloques de carpeta
		if inode.I_type[0] == '0' {
			block := &FolderBlock{}
			err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return -1, err
			}

			// Buscar en los contenidos del bloque (empezando desde 2 para saltar . y ..)
			for _, content := range block.B_content[2:] {
				if content.B_inodo == -1 {
					continue
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				if len(parentsDir) > 0 {
					// Si hay padres por recorrer, buscar el siguiente nivel
					parentDir := strings.Trim(parentsDir[0], "\x00 ")
					if strings.EqualFold(contentName, parentDir) {
						// Llamada recursiva con el siguiente nivel
						return sb.getInodeFromPath(path, content.B_inodo, parentsDir[1:], targetName)
					}
				} else {
					// Si no hay más padres, verificar si es el target
					target := strings.Trim(targetName, "\x00 ")
					if strings.EqualFold(contentName, target) {
						return content.B_inodo, nil
					}
				}
			}
		}
	}

	return -1, fmt.Errorf("no se encontró el inodo para '%s'", targetName)
}

func (sb *SuperBlock) loginUserInInode(user string, password string, path string) (int32, int32, error) {
	// obtener el contenido del bloque
	content := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(content))

	var contentEndPos int
	// encontrar donde termina realmente el contenido (primer byte nulo o fin del array)
	for i, b := range content {
		if b == 0 {
			contentEndPos = i
			break
		}
	}

	// si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(content)
	}

	// obtener solo el contenido real (sin bytes nulos)
	content = string(content[:contentEndPos])
	fmt.Println("Tamaño del contenido real: ", len(content))
	// eliminar los caracteres nulos

	// separar el contenido por lineas
	lines := strings.Split(content, "\n")
	// eliminar la ultima linea si es vacia
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	fmt.Println("Contenido login: ", content)

	uid := int32(0)
	gid := int32(0)
	grupoUsuario := ""
	// iterar sobre cada linea
	for _, line := range lines {
		// separar la linea por comas
		parts := strings.Split(line, ",")

		/*
			formato del archivo:
			GID, Tipo, Grupo \n
			UID, Tipo, Grupo, Usuario, Contraseña \n
		*/

		fmt.Println("Partes: ", parts)

		// verificar si es un usuario
		if parts[1] == "U" {
			fmt.Println("Usuario: ", parts[3])
			fmt.Println("Contraseña: ", parts[4])
			if parts[3] == user && parts[4] == password {
				fmt.Println("Usuario: ", parts[3])
				fmt.Println("Contraseña: ", parts[4])
				uid = utils.StringToInt32(parts[0])
				grupoUsuario = parts[2]
			} else if parts[3] == user && parts[4] != password {
				fmt.Println("Contraseña incorrecta")
				return 0, 0, fmt.Errorf("contraseña incorrecta")
			}
		}
	}

	// iterar para buscar el grupo
	for _, line := range lines {
		// separar la linea por comas
		parts := strings.Split(line, ",")
		// verificar si es un grupo
		if parts[1] == "G" {
			fmt.Println("Grupo: ", parts[2])
			if parts[2] == grupoUsuario {
				fmt.Println("Grupo: ", parts[2])
				gid = utils.StringToInt32(parts[0])
				fmt.Println("GID: ", gid)
				return uid, gid, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("usuario o contraseña incorrectos")

}

func (sb *SuperBlock) createGroupInInode(name string, path string) error {
	useBlock := &FileBlock{}

	// Deserializar el bloque
	err := useBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	// Separar el contenido por líneas
	lines := strings.Split(content, "\n")

	// Eliminar el último elemento si está vacío
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	countGrupos := 0
	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "G" {
			fmt.Println("Grupo: ", parts[2])
			countGrupos++
			if parts[2] == name {
				fmt.Println("Grupo ya existe")
				return fmt.Errorf("grupo ya existe")
			}
		}
	}

	// Si el grupo no existe, agregarlo
	newGrupo := fmt.Sprintf("%d,G,%s\n", countGrupos+1, name)
	newContent := content + newGrupo
	fmt.Println("Nuevo grupo: ", newGrupo)
	fmt.Println("Contenido nuevo: ", newContent)
	fmt.Println("Tamaño del nuevo contenido: ", len(newContent))

	err = sb.setUsersContent(path, newContent)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) removeGroupInInode(name string, path string) error {
	userBlock := &FileBlock{}

	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	// Separar el contenido por líneas
	lines := strings.Split(content, "\n")

	// Eliminar el último elemento si está vacío
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Variable para indicar si se encontró y modificó el grupo
	grupoEncontrado := false
	// Variable para almacenar el nuevo contenido
	var newContent strings.Builder

	for i, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 2 && parts[1] == "G" && parts[2] == name {
			fmt.Println("Grupo encontrado: ", parts[2])

			if parts[0] == "0" {
				return fmt.Errorf("El grupo ya no existe porque ya fue eliminado")
			}

			// Modificar el GID a 0 para marcar como eliminado
			parts[0] = "0"
			grupoEncontrado = true

			// Reconstruir la línea con el GID modificado
			modifiedLine := strings.Join(parts, ",")

			// Añadir la línea modificada
			newContent.WriteString(modifiedLine)
		} else {
			// Añadir la línea sin modificar
			newContent.WriteString(line)
		}

		// Añadir salto de línea si no es la última línea
		if i < len(lines)-1 {
			newContent.WriteString("\n")
		} else {
			// Añadir salto de línea al final
			newContent.WriteString("\n")
		}
	}

	if !grupoEncontrado {
		fmt.Println("Grupo no encontrado")
		return fmt.Errorf("grupo no encontrado")
	}

	// Obtener el contenido modificado
	finalContent := newContent.String()
	fmt.Println("Contenido modificado: ", finalContent)
	fmt.Println("Tamaño del contenido modificado: ", len(finalContent))

	err = sb.setUsersContent(path, finalContent)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) createUserInInode(user string, pass string, grp string, path string) error {
	userBlock := &FileBlock{}

	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	lines := strings.Split(content, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	userCount := 0           // Contador de usuarios
	userGroupExists := false // Variable para verificar si el grupo existe
	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "U" {
			fmt.Println("Usuario: ", parts[3])
			userCount++
			if parts[3] == user {
				fmt.Println("Usuario ya existe")
				return fmt.Errorf("usuario ya existe")
			}

		}
		// verificar si es un grupo
		if len(parts) > 1 && parts[1] == "G" {
			fmt.Println("Grupo: ", parts[2])
			if parts[2] == grp {
				// verificar que el grupo no haya sido eliminado
				if parts[0] == "0" {
					fmt.Println("Grupo eliminado")
					return fmt.Errorf("el grupo fue eliminado")
				}
				fmt.Println("Grupo existe")
				userGroupExists = true
			}
		}
	}

	if !userGroupExists {
		fmt.Println("Grupo no existe")
		return fmt.Errorf("grupo no existe")
	}

	// agregar el nuevo usuario: userCount, U, grp, user, pass
	newUser := fmt.Sprintf("%d,U,%s,%s,%s\n", userCount+1, grp, user, pass)
	newContent := content + newUser
	fmt.Println("Nuevo usuario: ", newUser)
	fmt.Println("Contenido nuevo: ", newContent)
	fmt.Println("Tamaño del nuevo contenido: ", len(newContent))

	err = sb.setUsersContent(path, newContent)
	if err != nil {
		return err
	}

	return nil

}

func (sb *SuperBlock) removeUserInInode(user string, path string) error {
	userBlock := &FileBlock{}

	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	lines := strings.Split(content, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	userExists := false // Variable para verificar si el usuario existe
	// Variable para almacenar el nuevo contenido
	var newContent strings.Builder
	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "U" && parts[3] == user {
			fmt.Println("Usuario encontrado: ", parts[3])

			if parts[0] == "0" {
				return fmt.Errorf("El usuario ya no existe porque ya fue eliminado")
			}

			// Modificar el UID a 0 para marcar como eliminado
			parts[0] = "0"
			userExists = true

			// Reconstruir la línea con el UID modificado
			modifiedLine := strings.Join(parts, ",")

			// Añadir la línea modificada
			newContent.WriteString(modifiedLine)
		} else {
			// Añadir la línea sin modificar
			newContent.WriteString(line)
		}

		// Añadir salto de línea si no es la última línea
		if line != lines[len(lines)-1] {
			newContent.WriteString("\n")
		} else {
			// Añadir salto de línea al final
			newContent.WriteString("\n")
		}

	}

	if !userExists {
		fmt.Println("Usuario no encontrado")
		return fmt.Errorf("usuario no encontrado")
	}

	// Obtener el contenido modificado
	finalContent := newContent.String()
	fmt.Println("Contenido modificado: ", finalContent)
	fmt.Println("Tamaño del contenido modificado: ", len(finalContent))
	err = sb.setUsersContent(path, finalContent)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) changeGroupInInode(user string, group string, path string) error {
	userBlock := &FileBlock{}
	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		fmt.Println("Error al deserializar el bloque: ", err)
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	lines := strings.Split(content, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	usuarioEncontrado := false
	grupoEncontrado := false
	var newContent strings.Builder

	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "U" && parts[3] == user {
			fmt.Println("Usuario encontrado: ", parts[3])

			if parts[0] == "0" {
				fmt.Println("El usuario ya no existe porque ya fue eliminado")
				return fmt.Errorf("el usuario ya no existe porque ya fue eliminado")
			}

			// Modificar el grupo
			parts[2] = group
			usuarioEncontrado = true

			modifiedLine := strings.Join(parts, ",")

			newContent.WriteString(modifiedLine)
		} else {
			newContent.WriteString(line)
		}

		if len(parts) > 1 && parts[1] == "G" {
			// Verificar si el grupo existe o si no fue eliminado
			if parts[2] == group {
				if parts[0] == "0" {
					fmt.Println("El grupo fue eliminado")
					return fmt.Errorf("el grupo fue eliminado")
				}
				fmt.Println("Grupo encontrado: ", parts[2])
				grupoEncontrado = true
			}
		}

		if line != lines[len(lines)-1] {
			newContent.WriteString("\n")
		} else {
			newContent.WriteString("\n")
		}
	}

	if !usuarioEncontrado {
		fmt.Println("Usuario no encontrado")
		return fmt.Errorf("usuario no encontrado")
	}

	if !grupoEncontrado {
		fmt.Println("Grupo no encontrado")
		return fmt.Errorf("grupo no encontrado")
	}

	// Obtener el contenido modificado
	finalContent := newContent.String()
	fmt.Println("Contenido modificado: ", finalContent)
	fmt.Println("Tamaño del contenido modificado: ", len(finalContent))

	err = sb.setUsersContent(path, finalContent)
	if err != nil {
		return err
	}

	return nil
}

// funcion para obtener el contenido de users.txt
func (sb *SuperBlock) getUsersContent(path string) string {
	// obtener el inodo para users.txt, este siempre será el inodo 1
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(1*sb.S_inode_size)))
	if err != nil {
		return ""
	}
	fmt.Println("Inodo de users.txt")
	inode.Print()

	// iterar sobre cada bloque del inodo y obtener el contenido
	contentBlocks := "" // para almacenar el contenido de los bloques
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		// crear un nuevo bloque de archivo
		block := &FileBlock{}
		// deserializar el bloque
		err = block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return ""
		}

		fmt.Println("Bloque de archivo: ", blockIndex)
		block.Print()
		// convertir el bloque a string
		blockContent := string(block.B_content[:])
		fmt.Println("Contenido del bloque: ", blockContent)
		// agregar el contenido del bloque al contenido total
		contentBlocks += blockContent
		fmt.Println("Contenido total: ", contentBlocks)

	}

	// returnar el contenido
	return contentBlocks
}

// funcion para setear el contenido de users.txt
func (sb *SuperBlock) setUsersContent(path string, content string) error {
	fmt.Println("Proximo bloque disponible: ", sb.S_blocks_count)
	// obtener el inodo para users.txt, este siempre será el inodo 1
	// dividir el contenido en bloques de 64 bytes - contentDivide
	// iterar sobre cada bloque del inodo y settear el contenido
	// si len(contentDivide) > bloques utilizados, entonces se necesita crear nuevo bloque para alamacenar el contenido
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(1*sb.S_inode_size)))
	if err != nil {
		return err
	}

	fmt.Println("Inodo de users.txt")
	inode.Print()
	// obtener el contenido de los bloques del inodo

	// crear un arreglo de bloques que almacene el contenido de 64 bytes
	contentBlocks := make([]string, 0)
	for i := 0; i < len(content); i += 64 {
		if i+64 > len(content) {
			contentBlocks = append(contentBlocks, content[i:])
		} else {
			contentBlocks = append(contentBlocks, content[i:i+64])
		}
	}

	fmt.Println("Contenido de los bloques: ", contentBlocks)

	// iterar sobre cada bloque del inodo y settear el contenido
	for i, blockIndex := range inode.I_block {
		fmt.Println("Bloque: ", blockIndex)
		// a cada bloque del inodo se le asigna el contenido de 64 bytes
		if blockIndex == -1 {
			if len(contentBlocks) > i {
				// todavia hay contenido por asignar pero el bloque no existe
				// actualizar el inodo para que apunte al nuevo bloque
				inode.I_block[i] = sb.S_blocks_count
				// serializar el inodo
				err = inode.Serialize(path, int64(sb.S_inode_start+(1*sb.S_inode_size)))
				if err != nil {
					return err
				}

				newBlock := &FileBlock{
					B_content: [64]byte{},
				}
				copy(newBlock.B_content[:], contentBlocks[i])
				fmt.Println("Bloque de archivo: ", blockIndex)
				// serializar el bloque
				err = newBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
				if err != nil {
					return err
				}

				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				// Actualizamos el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size
				break

			}
			break
		}

		// obtener el bloque
		block := &FileBlock{}
		// deserializar el bloque
		err = block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return err
		}

		// limpiar el bloque
		for j := range block.B_content {
			block.B_content[j] = 0
		}

		contentBytes := []byte(contentBlocks[i])
		// Copiar byte por byte el nuevo contenido
		for i := 0; i < len(contentBytes); i++ {
			block.B_content[i] = contentBytes[i]
		}
		fmt.Println("Bloque después de actualizar: ", string(block.B_content[:]))
		// serializar el bloque
		err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return err
		}
	}

	return nil
}
