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
