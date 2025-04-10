package reports

import (
	structures "backend/structures"
	"encoding/json"
	"fmt"
	"strings"
)

// FileSystemNode representa un nodo en el sistema de archivos (archivo o directorio)
type FileSystemNode struct {
	Name    string        `json:"name"`
	Type    int           `json:"type"`              // 0: directorio, 1: archivo
	Content []interface{} `json:"content,omitempty"` // Para directorios: []FileSystemNode, para archivos: []string
}

// GenerateFileSystemJSON genera la estructura JSON del sistema de archivos
func GenerateFileSystemJSON(superblock *structures.SuperBlock, diskPath string, path string) error {
	// Obtener la estructura raíz
	rootNode, err := generateNodeContent(superblock, diskPath, 0, make(map[int32]bool))
	if err != nil {
		return err
	}

	// Convertir a JSON
	jsonData, err := json.MarshalIndent(rootNode, "", "  ")
	if err != nil {
		return fmt.Errorf("error al generar JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// Función recursiva para generar la estructura de nodos
func generateNodeContent(superblock *structures.SuperBlock, diskPath string, inodeIndex int32, visited map[int32]bool) (*FileSystemNode, error) {
	if visited[inodeIndex] {
		return nil, nil
	}
	visited[inodeIndex] = true

	// Deserializar el inodo actual
	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(inodeIndex*superblock.S_inode_size)))
	if err != nil {
		return nil, err
	}

	node := &FileSystemNode{}

	// Determinar si es directorio (0) o archivo (1)
	if inode.I_type[0] == '0' {
		node.Type = 0
		node.Content = []interface{}{}
	} else {
		node.Type = 1
		node.Content = []interface{}{}
	}

	// Procesar bloques
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		if inode.I_type[0] == '0' { // Directorio
			folderBlock := &structures.FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
			if err != nil {
				return nil, err
			}

			// Procesar contenido del directorio
			for _, content := range folderBlock.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if content.B_inodo == -1 || name == "." || name == ".." {
					continue
				}

				// Obtener el nodo hijo
				childNode, err := generateNodeContent(superblock, diskPath, content.B_inodo, visited)
				if err != nil {
					return nil, err
				}

				if childNode != nil {
					childNode.Name = name
					node.Content = append(node.Content, *childNode)
				}
			}

		} else if inode.I_type[0] == '1' { // Archivo
			fileBlock := &structures.FileBlock{}
			err := fileBlock.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
			if err != nil {
				return nil, err
			}

			// Agregar contenido del archivo
			content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			if content != "" {
				node.Content = append(node.Content, content)
			}
		}
	}

	return node, nil
}
