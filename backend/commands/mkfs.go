package commands

import (
	stores	"backend/stores"
	structures "backend/structures"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

type MKFS struct {
	id  string
	typ string
}

func ParseMkfs(tokens []string) (string, error) {
	cmd := &MKFS{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-type=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-type":
			if value != "full" {
				return "", errors.New("el tipo debe ser full")
			}
			cmd.typ = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	if cmd.typ == "" {
		cmd.typ = "full"
	}

	err := commandMkfs(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return fmt.Sprintf("Se ha formateado la partición con id: %s", cmd.id), nil
}

func commandMkfs(mkfs *MKFS) error {
	mountedPartition, partitionPath, err := stores.GetMountedPartition(mkfs.id)
	if err != nil {
		return err
	}

	fmt.Println("\nPatición montada:")
	mountedPartition.PrintPartition()

	n := calculateN(mountedPartition)

	fmt.Println("\nValor de n:", n)

	superBlock := createSuperBlock(mountedPartition, n)

	fmt.Println("\nSuperBlock:")
	superBlock.Print()

	err = superBlock.CreateBitMaps(partitionPath)
	if err != nil {
		return err
	}

	err = superBlock.CreateUsersFile(partitionPath)
	if err != nil {
		return err
	}

	fmt.Println("\nSuperBlock actualizado:")
	superBlock.Print()

	err = superBlock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}

	return nil
}

func calculateN(partition *structures.Partition) int32 {
	numerator := int(partition.Part_size) - binary.Size(structures.SuperBlock{})
	denominator := 4 + binary.Size(structures.Inode{}) + 3*binary.Size(structures.FileBlock{})
	n := math.Floor(float64(numerator) / float64(denominator))

	return int32(n)
}

func createSuperBlock(partition *structures.Partition, n int32) *structures.SuperBlock {
	bm_inode_start := partition.Part_start + int32(binary.Size(structures.SuperBlock{}))
	bm_block_start := bm_inode_start + n
	inode_start := bm_block_start + (3 * n)
	block_start := inode_start + (int32(binary.Size(structures.Inode{})) * n)

	superBlock := &structures.SuperBlock{
		S_filesystem_type:   2,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float32(time.Now().Unix()),
		S_umtime:            float32(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_size:        int32(binary.Size(structures.Inode{})),
		S_block_size:        int32(binary.Size(structures.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
	}
	return superBlock
}