package structures

import (
	utils "backend/utils"
	"fmt"
	"strings"
	"time"
)

// createFolderInode crea una carpeta en un inodo específico
func (sb *SuperBlock) createFolderInode(path string, inodeIndex int32, parentsDir []string, destDir string, uid int32, gid int32, folderPath string, journalStart int64) error {
	// Crear un nuevo inodo
	inode := &Inode{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for i, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if i > 14 {
			fmt.Println("Ya no hay más espacio en el inodo")
			break
		}
		if blockIndex == -1 {
			if blockIndex == -1 {
				fmt.Println("Ya no hay más bloques, creando uno nuevo")
				fmt.Println("i:", i)

				// 1. Crear y serializar el nuevo bloque de carpeta
				newBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// Actualizar contenido del bloque
				destDirByte := [12]byte{}
				copy(destDirByte[:], destDir)
				newBlock.B_content[2] = FolderContent{B_name: destDirByte, B_inodo: sb.S_inodes_count}

				// Guardar posición del nuevo bloque
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

				if sb.S_filesystem_type == 3 {
					// Jornaling
					// Crear el journal
					bytePath := [32]byte{}
					copy(bytePath[:], folderPath)
					journal := &Journal{
						J_count: sb.S_inodes_count,
						J_content: Information{
							I_operation: [10]byte{'m', 'k', 'd', 'i', 'r'},
							I_path:      bytePath,
							I_content:   [64]byte{},
							I_date:      float32(time.Now().Unix()),
						},
					}
					fmt.Println("Journal:")
					journal.Print()
					// Serializar el journal
					err = journal.Serialize(path, journalStart)
					if err != nil {
						return err
					}
				}

				// 2. Crear el inodo de la nueva carpeta
				folderInode := &Inode{
					I_uid:   uid,
					I_gid:   gid,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Serializar el nuevo inodo
				err = folderInode.Serialize(path, int64(sb.S_first_ino))
				if err != nil {
					return err
				}

				// Actualizar bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return err
				}

				// Actualizar superbloque (nuevo inodo)
				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size

				// 3. Crear bloque para el nuevo inodo
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: sb.S_inodes_count - 1}, // Apunta a sí mismo
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},       // Apunta al padre
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// Serializar el bloque de la nueva carpeta
				err = folderBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
				if err != nil {
					return err
				}

				// Actualizar bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				// 4. Manejar apuntadores indirectos si es necesario
				if i >= 12 { // Bloques indirectos
					var pointerBlock *PointerBlock
					pointerPos := sb.S_blocks_count

					switch i {
					case 12: // Indirecto simple
						fmt.Println("Creando bloque de apuntadores indirecto simple")
						pointerBlock = &PointerBlock{
							P_pointers: [16]int32{
								newBlockPos, // Primer bloque que creamos
								-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
							},
						}

					case 13: // Indirecto doble
						fmt.Println("Creando bloque de apuntadores indirecto doble")
						// Implementar lógica similar para doble indirecto
						return fmt.Errorf("apuntadores indirectos dobles no implementados aún")

					case 14: // Indirecto triple
						fmt.Println("Creando bloque de apuntadores indirecto triple")
						// Implementar lógica similar para triple indirecto
						return fmt.Errorf("apuntadores indirectos triples no implementados aún")
					}

					// Serializar el bloque de apuntadores
					err = pointerBlock.Serialize(path, int64(sb.S_block_start+(pointerPos*sb.S_block_size)))
					if err != nil {
						return err
					}

					// Actualizar bitmap de bloques
					err = sb.UpdateBitmapBlock(path)
					if err != nil {
						return err
					}

					// Actualizar superbloque (bloque de apuntadores)
					sb.S_blocks_count++
					sb.S_free_blocks_count--
					sb.S_first_blo += sb.S_block_size

					// Actualizar inodo con referencia al bloque de apuntadores
					inode.I_block[i] = pointerPos
				} else {
					// Para bloques directos, simplemente actualizar la referencia
					inode.I_block[i] = newBlockPos
				}

				// Serializar el inodo actualizado
				err = inode.Serialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
				if err != nil {
					return err
				}

				return nil
			}
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				fmt.Println("---------ESTOY  VISITANDO--------")

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
					//fmt.Println("---------LA ENCONTRÉ-------")
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					err := sb.createFolderInode(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir, uid, gid, folderPath, journalStart)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				fmt.Println("---------ESTOY  CREANDO--------")

				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if content.B_inodo != -1 {
					fmt.Println("Ya no hay espacio en el bloque")
					continue
				}

				if i == 12 || i == 13 || i == 14 {
					fmt.Println("estoy en un bloque indirecto")
					pointerBlock := &PointerBlock{}
					err := pointerBlock.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
					if err != nil {
						return err
					}
					for j := 0; j < len(pointerBlock.P_pointers); j++ {
						if pointerBlock.P_pointers[j] != -1 {
							folderBlock := &FolderBlock{}
							err := folderBlock.Deserialize(path, int64(sb.S_block_start+(pointerBlock.P_pointers[j]*sb.S_block_size)))
							if err != nil {
								return err
							}

							for indexContentpb := 2; indexContentpb < len(folderBlock.B_content); indexContentpb++ {
								contentpb := folderBlock.B_content[indexContentpb]

								if contentpb.B_inodo != -1 {
									fmt.Println("Ya no hay espacio en el bloque")
									continue
								}

								if contentpb.B_inodo == -1 {
									// añadir el contenido al bloque
									copy(contentpb.B_name[:], destDir)
									contentpb.B_inodo = sb.S_inodes_count

									// Actualizar el bloque
									folderBlock.B_content[indexContentpb] = contentpb
									// Serializar el bloque
									err = folderBlock.Serialize(path, int64(sb.S_block_start+(pointerBlock.P_pointers[j]*sb.S_block_size)))
									if err != nil {
										return err
									}

									// crear el inodo de la carpeta
									folderInode := &Inode{
										I_uid:   uid,
										I_gid:   gid,
										I_size:  0,
										I_atime: float32(time.Now().Unix()),
										I_ctime: float32(time.Now().Unix()),
										I_mtime: float32(time.Now().Unix()),
										I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
										I_type:  [1]byte{'0'},
										I_perm:  [3]byte{'6', '6', '4'},
									}

									// Serializar el inodo de la carpeta
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

									// Crear el bloque de la carpeta
									newFolderBlock := &FolderBlock{
										B_content: [4]FolderContent{
											{B_name: [12]byte{'.'}, B_inodo: contentpb.B_inodo},
											{B_name: [12]byte{'.', '.'}, B_inodo: content.B_inodo},
											{B_name: [12]byte{'-'}, B_inodo: -1},
											{B_name: [12]byte{'-'}, B_inodo: -1},
										},
									}
									// Serializar el bloque de la carpeta
									err = newFolderBlock.Serialize(path, int64(sb.S_first_blo))
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

									return nil
								}
							}
						} else {
							// se necesita crear otro bloque para el bloque de apuntadores
							newBlock := &FolderBlock{
								B_content: [4]FolderContent{
									{B_name: [12]byte{'.'}, B_inodo: inodeIndex},
									{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
									{B_name: [12]byte{'-'}, B_inodo: -1},
									{B_name: [12]byte{'-'}, B_inodo: -1},
								},
							}

							destDirByte := [12]byte{}
							copy(destDirByte[:], destDir)
							newBlock.B_content[2] = FolderContent{B_name: destDirByte, B_inodo: sb.S_inodes_count}
							// Guardar posición del nuevo bloque
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

							// 2. Crear el inodo de la nueva carpeta
							folderInode := &Inode{
								I_uid:   uid,
								I_gid:   gid,
								I_size:  0,
								I_atime: float32(time.Now().Unix()),
								I_ctime: float32(time.Now().Unix()),
								I_mtime: float32(time.Now().Unix()),
								I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
								I_type:  [1]byte{'0'},
								I_perm:  [3]byte{'6', '6', '4'},
							}

							// Serializar el nuevo inodo
							err = folderInode.Serialize(path, int64(sb.S_first_ino))
							if err != nil {
								return err
							}

							// Actualizar bitmap de inodos
							err = sb.UpdateBitmapInode(path)
							if err != nil {
								return err
							}

							// Actualizar superbloque (nuevo inodo)
							sb.S_inodes_count++
							sb.S_free_inodes_count--
							sb.S_first_ino += sb.S_inode_size

							// 3. Crear bloque para el nuevo inodo
							folderBlock := &FolderBlock{
								B_content: [4]FolderContent{
									{B_name: [12]byte{'.'}, B_inodo: sb.S_inodes_count - 1}, // Apunta a sí mismo
									{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},       // Apunta al padre
									{B_name: [12]byte{'-'}, B_inodo: -1},
									{B_name: [12]byte{'-'}, B_inodo: -1},
								},
							}

							// Serializar el bloque de la nueva carpeta
							err = folderBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
							if err != nil {
								return err
							}

							// Actualizar bitmap de bloques
							err = sb.UpdateBitmapBlock(path)
							if err != nil {
								return err
							}
							sb.S_blocks_count++
							sb.S_free_blocks_count--
							sb.S_first_blo += sb.S_block_size

							// actualizar el bloque de apuntadores
							pointerBlock.P_pointers[j] = newBlockPos
							// Serializar el bloque de apuntadores
							err = pointerBlock.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
							if err != nil {
								return err
							}
							fmt.Println("Actualizando el bloque de apuntadores")
							pointerBlock.Print()
							return nil
						}
					}
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
				fmt.Println("Tipoe de sistema de archivos:", sb.S_filesystem_type)
				fmt.Println(sb.S_filesystem_type == 3)
				if sb.S_filesystem_type == 3 {
					// Jornaling
					// Crear el journal
					bytePath := [32]byte{}
					copy(bytePath[:], folderPath)
					journal := &Journal{
						J_count: sb.S_inodes_count,
						J_content: Information{
							I_operation: [10]byte{'m', 'k', 'd', 'i', 'r'},
							I_path:      bytePath,
							I_content:   [64]byte{},
							I_date:      float32(time.Now().Unix()),
						},
					}
					fmt.Println("Journal:")
					journal.Print()

					// Serializar el journal
					err = journal.Serialize(path, journalStart)
					if err != nil {
						return err
					}
				}
				// Crear el inodo de la carpeta
				folderInode := &Inode{
					I_uid:   uid,
					I_gid:   gid,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Serializar el inodo de la carpeta
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

				// Crear el bloque de la carpeta
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				// Serializar el bloque de la carpeta
				err = folderBlock.Serialize(path, int64(sb.S_first_blo))
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

				return nil
			}
		}

	}
	return nil
}
