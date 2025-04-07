package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const EBRSize = 32

type EBR struct {
	Ebr_part_mount [1]byte // if the partition is mounted
	Ebr_part_fit  [1]byte // fit type: B, F, W
	Ebr_part_start int32 // starting byte of the partition
	Ebr_part_size int32 // size of the partition in bytes
	Ebr_part_next int32 // next EBR in the linked list
	Ebr_part_name [16]byte // name of the partition
}

func(ebr *EBR) SerializeEBR(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	// Move the file pointer to the specified offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serialize the ebr structure directly into the file
	err = binary.Write(file, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}

	return nil
}

func(ebr *EBR) DeserializeEBR(path string, offset int64) error {
	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move the file pointer to the specified offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Get the size of the ebr structure
	ebrSize := binary.Size(ebr)
	if ebrSize <= 0 {
		return fmt.Errorf("invalid ebr size: %d", ebrSize)
	}

	// Read only the number of bytes corresponding to the size of the Inode structure
	buffer := make([]byte, ebrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserialize the read bytes into the Inode structure
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}

	return nil
}


func(ebr *EBR) PrintEBR() {
	partMount := rune(ebr.Ebr_part_mount[0])
	partFit := rune(ebr.Ebr_part_fit[0])
	partName := string(ebr.Ebr_part_name[:])

	fmt.Println("EBR")
	fmt.Printf("  Mount: %c\n", partMount)
	fmt.Printf("  Fit: %c\n", partFit)
	fmt.Printf("  Start: %d\n", ebr.Ebr_part_start)
	fmt.Printf("  Size: %d\n", ebr.Ebr_part_size)
	fmt.Printf("  Next: %d\n", ebr.Ebr_part_next)
	fmt.Printf("  Name: %s\n", partName)
}

func(ebr *EBR) GetFreeEBR() (*EBR, int) {
	offset := binary.Size(ebr)

	if ebr.Ebr_part_start == -1 {
		return ebr, offset
	}

	return nil, -1
}

func(ebr *EBR) GetEBRByName(name string) (*EBR, int) {
	if string(ebr.Ebr_part_name[:]) == name {
		return ebr, 0
	}

	return nil, -1
}

func(ebr *EBR) GetEBRByStart(start int32) (*EBR, int) {
	if ebr.Ebr_part_start == start {
		return ebr, 0
	}

	return nil, -1
}

func(ebr *EBR) GetEBRByNext(next int32) (*EBR, int) {
	if ebr.Ebr_part_next == next {
		return ebr, 0
	}

	return nil, -1
}


