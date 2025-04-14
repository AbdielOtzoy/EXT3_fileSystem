package structures

import (
	utils "backend/utils"
	"fmt"
	"strings"
)

// RemoveInode elimina un inodo
func (sb *SuperBlock) RemoveInode(path string, inodeIndex int32, parentsDir []string, destDir string) error {
	// deserializar el inodo
	inode := &Inode{}
	// deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar el inodo: %w", err)
	}

	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return fmt.Errorf("el inodo %d es de tipo carpeta", inodeIndex)
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		fmt.Println("Bloque actual:", blockIndex)
		if blockIndex == -1 {
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}
		block.Print()

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]
			fmt.Println("Contenido del bloque:", string(content.B_name[:]))
			fmt.Println("Inodo del contenido:", content.B_inodo)

			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					return fmt.Errorf("el inodo %d no tiene contenido", blockIndex)
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
					fmt.Println("entrando a la carpeta padre")
					err := sb.RemoveInode(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return err
					}
					return nil
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
					// borrar la referencia del inodo en el bloque
					block.B_content[indexContent] = FolderContent{B_name: [12]byte{'-'}, B_inodo: -1}
					// serializar el bloque
					err := block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
					if err != nil {
						return fmt.Errorf("error al serializar el bloque: %w", err)
					}
					return nil
				}
			}
		}

	}

	return nil
}

// EditFileInInode edita el contenido de un archivo en el sistema de archivos
func (sb *SuperBlock) EditFileInInode(path string, inodeIndex int32, parentsDir []string, destDir string, contentFile string, uid int32, gid int32) error {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return fmt.Errorf("el inodo %d es de tipo carpeta", inodeIndex)
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		fmt.Println("Bloque: ", blockIndex)
		block.Print()

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					break
				}

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
					fmt.Println("entrando a la carpeta padre")
					err := sb.EditFileInInode(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, contentFile, uid, gid)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				if content.B_inodo == -1 {
					continue
				}
				destDirByte := [12]byte{}
				copy(destDirByte[:], destDir)

				if content.B_name == destDirByte {
					fmt.Println("---------LA ENCONTRÉ-------")
					inodeFile := &Inode{}
					err := inodeFile.Deserialize(path, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
					if err != nil {
						return err
					}

					if inodeFile.I_type[0] == '1' {
						// modificar el contenido del archivo
						contentBlocks := make([]string, 0)
						for i := 0; i < len(contentFile); i += 64 {
							if i+64 > len(contentFile) {
								contentBlocks = append(contentBlocks, contentFile[i:])
							} else {
								contentBlocks = append(contentBlocks, contentFile[i:i+64])
							}
						}

						// iterar sobre cada bloque del inodo y settear el contenido
						for i, blockIndex := range inodeFile.I_block {
							// Si ya no hay más contenido para escribir, podemos terminar
							if i >= len(contentBlocks) {
								break
							}

							if blockIndex == -1 {
								// todavía hay contenido por asignar pero el bloque no existe
								// actualizar el inodo para que apunte al nuevo bloque
								inodeFile.I_block[i] = sb.S_blocks_count
								// serializar el inodo
								err := inodeFile.Serialize(path, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
								if err != nil {
									return err
								}

								newBlock := &FileBlock{
									B_content: [64]byte{},
								}
								// copiar el contenido del bloque
								copy(newBlock.B_content[:], contentBlocks[i])
								// serializar el bloque
								err = newBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
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
								continue // Ya hemos procesado este bloque, continuamos con el siguiente
							}

							// obtener el bloque
							block := &FileBlock{}
							err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
							if err != nil {
								return err
							}

							// limpiar el bloque
							for j := range block.B_content {
								block.B_content[j] = 0
							}

							contentBytes := []byte(contentBlocks[i])
							// Copiar byte por byte el nuevo contenido
							for j := 0; j < len(contentBytes); j++ {
								block.B_content[j] = contentBytes[j]
							}

							// serializar el bloque
							err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
							if err != nil {
								return err
							}
						}
						return nil
					}
					return fmt.Errorf("el inodo no es de tipo archivo")
				}
			}
		}
	}

	return fmt.Errorf("no se encontró el archivo")
}

// RenameFileInInode cambia el nombre de un archivo o carpeta
func (sb *SuperBlock) RenameFileInInode(path string, inodeIndex int32, parentsDir []string, destDir string, newName string, uid int32, gid int32) error {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return fmt.Errorf("el inodo %d es de tipo carpeta", inodeIndex)
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		fmt.Println("Bloque: ", blockIndex)
		block.Print()

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					break
				}

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
					fmt.Println("entrando a la carpeta padre")
					err := sb.RenameFileInInode(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, newName, uid, gid)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				if content.B_inodo == -1 {
					continue
				}
				destDirByte := [12]byte{}
				copy(destDirByte[:], destDir)

				if content.B_name == destDirByte {
					fmt.Println("---------LA ENCONTRÉ-------")
					// Cambiar el nombre del archivo o carpeta en el bloque
					block.B_content[indexContent].B_name = [12]byte{}
					copy(block.B_content[indexContent].B_name[:], newName)
					// Serializar el bloque
					err := block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
					if err != nil {
						return err
					}
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no se encontró el archivo")
}
