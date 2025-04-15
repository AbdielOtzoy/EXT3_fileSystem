package commands

import (
	stores "backend/stores"
	structures "backend/structures"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type FileSystemNodeWithRef struct {
	Name       string        `json:"name"`
	Type       int           `json:"type"`
	Content    []interface{} `json:"content,omitempty"`
	InodeRef   int32         `json:"inodeRef"`
	IsRef      bool          `json:"isRef,omitempty"`
	RefContent []interface{} `json:"refContent,omitempty"` // Contenido del nodo referenciado
}

func ParseGetfs(tokens []string) (string, error) {
	if len(tokens) != 0 {
		return "", errors.New("no se esperaban argumentos")
	}

	var disks []map[string]interface{}
	seenPaths := make(map[string]bool) // Mapa para trackear paths ya procesados

	for _, id := range stores.GetMountedPartitions() {
		mountedMbr, mountedSb, mountedDiskPath, err := stores.GetMountedPartitionRep(id)
		if err != nil {
			return "", fmt.Errorf("error al obtener la partición montada: %v", err)
		}

		// Verificar si ya hemos procesado este path de disco
		if _, exists := seenPaths[mountedDiskPath]; exists {
			continue // Saltar este disco, ya fue procesado
		}

		// Marcar este path como procesado
		seenPaths[mountedDiskPath] = true

		disk := map[string]interface{}{
			"diskPath":      mountedDiskPath,
			"diskSize":      mountedMbr.Mbr_size,
			"diskSignature": mountedMbr.Mbr_disk_signature,
			"diskFit":       string(mountedMbr.Mbr_disk_fit[0]),
			"partitions":    getPartitionsInfo(mountedMbr, mountedSb, mountedDiskPath, id),
		}

		disks = append(disks, disk)
	}

	jsonData, err := json.MarshalIndent(disks, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error al generar JSON: %v", err)
	}

	return string(jsonData), nil
}

func getPartitionsInfo(mbr *structures.MBR, sb *structures.SuperBlock, diskPath string, mountedID string) []map[string]interface{} {
	var partitions []map[string]interface{}

	for i, part := range mbr.Mbr_partitions {
		if part.Part_size == 0 {
			continue
		}

		partitionID := strings.TrimRight(string(part.Part_id[:]), "\x00")

		partition := map[string]interface{}{
			"partitionNumber": i + 1,
			"partitionSize":   part.Part_size,
			"partitionType":   string(part.Part_type[0]),
			"partitionFit":    string(part.Part_fit[0]),
			"partitionStart":  part.Part_start,
			"partitionName":   strings.TrimRight(string(part.Part_name[:]), "\x00"),
			"partitionID":     partitionID,
		}

		// Solo agregar filesystem si es la partición montada
		if partitionID == mountedID {
			partition["fs"] = getFileSystemStructure(sb, diskPath)
		} else {
			partition["fs"] = nil
		}

		partitions = append(partitions, partition)
	}

	return partitions
}

func getFileSystemStructure(sb *structures.SuperBlock, diskPath string) *FileSystemNodeWithRef {
	if sb == nil {
		return nil
	}

	// Creamos un mapa para almacenar los nodos ya procesados completamente
	nodesCache := make(map[int32]*FileSystemNodeWithRef)
	// Creamos un conjunto para llevar un seguimiento de inodos que están siendo procesados actualmente
	// para evitar referencias cíclicas durante la construcción inicial del árbol
	processingNodes := make(map[int32]bool)

	rootNode, err := generateNodeContent(sb, diskPath, 0, nodesCache, processingNodes)
	if err != nil {
		return nil
	}

	return rootNode
}

func generateNodeContent(superblock *structures.SuperBlock, diskPath string, inodeIndex int32,
	nodesCache map[int32]*FileSystemNodeWithRef, processingNodes map[int32]bool) (*FileSystemNodeWithRef, error) {

	// Si estamos procesando actualmente este inodo (para evitar ciclos durante la construcción)
	if processingNodes[inodeIndex] {
		// Crear un marcador temporal que será completado después
		return &FileSystemNodeWithRef{
			Name:     "processing",
			Type:     -1, // Marcador temporal
			InodeRef: inodeIndex,
			IsRef:    true,
		}, nil
	}

	// Verificar si este inodo ya fue completamente procesado
	if node, exists := nodesCache[inodeIndex]; exists {
		// Crear un nodo de referencia pero incluyendo una copia del contenido
		refNode := &FileSystemNodeWithRef{
			Name:       node.Name,
			Type:       node.Type,
			InodeRef:   inodeIndex,
			IsRef:      true,
			RefContent: make([]interface{}, 0), // Inicializar con un slice vacío
		}

		// Si el nodo original tiene contenido, copiar la referencia al contenido
		if node.Content != nil {
			refNode.RefContent = node.Content
		}

		return refNode, nil
	}

	// Marcar este inodo como "en procesamiento"
	processingNodes[inodeIndex] = true

	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(inodeIndex*superblock.S_inode_size)))
	if err != nil {
		delete(processingNodes, inodeIndex) // Limpiar el estado de procesamiento
		return nil, err
	}

	// Crear un nuevo nodo
	node := &FileSystemNodeWithRef{
		Name:     "/",
		Type:     0,
		InodeRef: inodeIndex,
		Content:  make([]interface{}, 0), // Inicializar con un slice vacío
	}

	if inode.I_type[0] == '1' {
		node.Type = 1
	}

	// Procesamiento de bloques del inodo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		if inode.I_type[0] == '0' { // Directorio
			folderBlock := &structures.FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
			if err != nil {
				delete(processingNodes, inodeIndex) // Limpiar el estado de procesamiento
				return nil, err
			}

			for _, content := range folderBlock.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if content.B_inodo == -1 || name == "." || name == ".." {
					continue
				}

				childNode, err := generateNodeContent(superblock, diskPath, content.B_inodo, nodesCache, processingNodes)
				if err != nil {
					delete(processingNodes, inodeIndex) // Limpiar el estado de procesamiento
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
				delete(processingNodes, inodeIndex) // Limpiar el estado de procesamiento
				return nil, err
			}

			content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			if content != "" {
				node.Content = append(node.Content, content)
			}
		}
	}

	// Ya no estamos procesando este inodo
	delete(processingNodes, inodeIndex)

	// Guardar el nodo completo en el caché
	nodesCache[inodeIndex] = node

	// Reemplazar los marcadores temporales de este inodo en el árbol si existen
	updateTemporaryMarkers(node, nodesCache)

	return node, nil
}

// Función auxiliar para actualizar los marcadores temporales
func updateTemporaryMarkers(completeNode *FileSystemNodeWithRef, nodesCache map[int32]*FileSystemNodeWithRef) {
	for _, nodePtr := range nodesCache {
		// Si es un directorio, buscar en su contenido los marcadores temporales
		if nodePtr.Type == 1 && nodePtr.Content != nil {
			for i := range nodePtr.Content {
				// Convertir el elemento de la interfaz a FileSystemNodeWithRef
				if contentNode, ok := nodePtr.Content[i].(FileSystemNodeWithRef); ok {
					// Si es un marcador temporal para el inodo que acabamos de completar
					if contentNode.Type == -1 && contentNode.InodeRef == completeNode.InodeRef {
						// Reemplazar con una referencia completa
						nodePtr.Content[i] = FileSystemNodeWithRef{
							Name:       contentNode.Name, // Mantener el nombre asignado
							Type:       completeNode.Type,
							InodeRef:   completeNode.InodeRef,
							IsRef:      true,
							RefContent: completeNode.Content,
						}
					}
				}
			}
		}
	}
}

// Función para obtener un clon del contenido de un nodo
func cloneContent(content []interface{}) []interface{} {
	if content == nil {
		return nil
	}

	result := make([]interface{}, len(content))
	for i, item := range content {
		switch v := item.(type) {
		case FileSystemNodeWithRef:
			// Clonar el nodo
			nodeCopy := v
			// Si el nodo tiene contenido, clonarlo recursivamente
			if v.Content != nil {
				nodeCopy.Content = cloneContent(v.Content)
			}
			result[i] = nodeCopy
		default:
			// Para otros tipos (como strings), copiar directamente
			result[i] = v
		}
	}

	return result
}
