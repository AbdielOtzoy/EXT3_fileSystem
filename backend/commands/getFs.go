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
	Name     string        `json:"name"`
	Type     int           `json:"type"`
	Content  []interface{} `json:"content,omitempty"`
	InodeRef int32         `json:"inodeRef"`
	IsRef    bool          `json:"isRef,omitempty"`
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

	// Creamos un mapa para almacenar los nodos ya generados
	nodesCache := make(map[int32]*FileSystemNodeWithRef)

	rootNode, err := generateNodeContent(sb, diskPath, 0, nodesCache)
	if err != nil {
		return nil
	}

	return rootNode
}

func generateNodeContent(superblock *structures.SuperBlock, diskPath string, inodeIndex int32, nodesCache map[int32]*FileSystemNodeWithRef) (*FileSystemNodeWithRef, error) {
	// Verificar si este inodo ya fue procesado
	if node, exists := nodesCache[inodeIndex]; exists {
		// Crear una referencia al nodo ya existente
		return &FileSystemNodeWithRef{
			Name:     node.Name,
			Type:     node.Type,
			InodeRef: inodeIndex,
			IsRef:    true,
		}, nil
	}

	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(inodeIndex*superblock.S_inode_size)))
	if err != nil {
		return nil, err
	}

	// Crear un nuevo nodo
	node := &FileSystemNodeWithRef{
		Name:     "/",
		Type:     0,
		InodeRef: inodeIndex,
	}

	// Agregar el nodo al cache antes de procesar su contenido para manejar referencias circulares
	nodesCache[inodeIndex] = node

	if inode.I_type[0] == '1' {
		node.Type = 1
		node.Content = []interface{}{}
	}

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

			for _, content := range folderBlock.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if content.B_inodo == -1 || name == "." || name == ".." {
					continue
				}

				childNode, err := generateNodeContent(superblock, diskPath, content.B_inodo, nodesCache)
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

			content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			if content != "" {
				node.Content = append(node.Content, content)
			}
		}
	}

	return node, nil
}
