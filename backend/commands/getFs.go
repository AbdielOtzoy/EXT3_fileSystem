package commands

import (
	stores "backend/stores"
	structures "backend/structures"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type FileSystemNode struct {
	Name    string        `json:"name"`
	Type    int           `json:"type"`
	Content []interface{} `json:"content,omitempty"`
}

func ParseGetfs(tokens []string) (string, error) {
	if len(tokens) != 0 {
		return "", errors.New("no se esperaban argumentos")
	}

	var disks []map[string]interface{}

	for _, id := range stores.GetMountedPartitions() {
		mountedMbr, mountedSb, mountedDiskPath, err := stores.GetMountedPartitionRep(id)
		if err != nil {
			return "", fmt.Errorf("error al obtener la partición montada: %v", err)
		}

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

func getFileSystemStructure(sb *structures.SuperBlock, diskPath string) *FileSystemNode {
	if sb == nil {
		return nil
	}

	rootNode, err := generateNodeContent(sb, diskPath, 0, make(map[int32]bool))
	if err != nil {
		return nil
	}

	return rootNode
}

func generateNodeContent(superblock *structures.SuperBlock, diskPath string, inodeIndex int32, visited map[int32]bool) (*FileSystemNode, error) {
	if visited[inodeIndex] {
		return nil, nil
	}
	visited[inodeIndex] = true

	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(inodeIndex*superblock.S_inode_size)))
	if err != nil {
		return nil, err
	}

	node := &FileSystemNode{
		Name: "/",
		Type: 0,
	}

	if inode.I_type[0] == '1' {
		node.Type = 1
		node.Content = []interface{}{}
	}

	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		if inode.I_type[0] == '0' {
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

				childNode, err := generateNodeContent(superblock, diskPath, content.B_inodo, visited)
				if err != nil {
					return nil, err
				}

				if childNode != nil {
					childNode.Name = name
					node.Content = append(node.Content, *childNode)
				}
			}

		} else if inode.I_type[0] == '1' {
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
