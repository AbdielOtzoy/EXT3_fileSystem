package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type MBR struct {
	Mbr_size			int32
	Mbr_creation_date	float32
	Mbr_disk_signature	int32
	Mbr_disk_fit		[1]byte
	Mbr_partitions		[4]Partition
}

// serializes the MBR struct to a byte array
func(mbr *MBR) SerializeMBR(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// deserializes a byte array to a MBR struct
func (mbr *MBR) DeserializeMBR(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	// get the size of the MBR struct
	mbrSize := binary.Size(mbr)
	if mbrSize <= 0 {
		return fmt.Errorf("invalid size of MBR struct: %d", mbrSize)
	}

	// read the number of bytes of the MBR struct
	buffer := make([]byte, mbrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserialize the MBR struct
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// Get the first free partition in the MBR
func (mbr *MBR) GetFreePartition() (*Partition, int, int) {
	// calculate the offset of the first partition
	offset := binary.Size(mbr)

	// iterate over the partitions to find the first free partition
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		// if the start of the partition is -1 then it is free
		if mbr.Mbr_partitions[i].Part_start == -1 {
			return &mbr.Mbr_partitions[i], offset, i
		} else {
			offset += binary.Size(mbr.Mbr_partitions[i])
		}
	}

	return nil, -1, -1
}

// Get the partition with the given name
func (mbr *MBR) GetPartitionByName(name string) (*Partition, int) {
	// iterate over the partitions to find the partition with the given name
	for i, partition := range mbr.Mbr_partitions {
		partitionName := strings.Trim(string(partition.Part_name[:]), "\x00")
		inputName := strings.Trim(name, "\x00")
		fmt.Println("partitionName:", partitionName)
		fmt.Println("inputName:", inputName)
		fmt.Println("equal:", strings.EqualFold(partitionName, inputName))

		if strings.EqualFold(partitionName, inputName) {
			return &partition, i+1
		}
	}

	return nil, -1
}

func (mbr *MBR) GetPartitionByID(id string) (*Partition, error) {
	for _, partition := range mbr.Mbr_partitions {
		partID := strings.Trim(string(partition.Part_id[:]), "\x00")
		if strings.EqualFold(partID, id) {
			return &partition, nil
		}
	}

	return nil, fmt.Errorf("partition with ID %s not found", id)
}

// GetExtendedPartition
func (mbr *MBR) GetExtendedPartition() (*Partition, int) {
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_type[0] == 'E' {
			return &mbr.Mbr_partitions[i], i
		}
	}
	return nil, -1
}

// Get free space in the MBR
func (mbr *MBR) GetFreeSpace() int32 {
	var totalUsedSpace int32
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_start != -1 {
			totalUsedSpace += partition.Part_size
		}
	}
	return mbr.Mbr_size - totalUsedSpace
}
// print the values of the MBR struct
func (mbr *MBR) PrintMBR() {
	creationTime := time.Unix(int64(mbr.Mbr_creation_date), 0) // convert to time.Time

	diskFit := rune(mbr.Mbr_disk_fit[0]) // convert to rune

	fmt.Println("MBR")
	fmt.Printf("Size: %d\n", mbr.Mbr_size)
	fmt.Printf("Creation Date: %s\n", creationTime)
	fmt.Printf("Disk Signature: %d\n", mbr.Mbr_disk_signature)
	fmt.Printf("Disk Fit: %c\n", diskFit)
}

// print the partition of the MBR struct
func (mbr *MBR) PrintPartitions() {
	for i, partition := range mbr.Mbr_partitions {
		partStatus := rune(partition.Part_status[0])
		partType := rune(partition.Part_type[0])
		partFit := rune(partition.Part_fit[0])

		partName := string(partition.Part_name[:])
		partID := string(partition.Part_id[:])

		fmt.Printf("Partition %d:\n", i+1)
		fmt.Printf("  Status: %c\n", partStatus)
		fmt.Printf("  Type: %c\n", partType)
		fmt.Printf("  Fit: %c\n", partFit)
		fmt.Printf("  Start: %d\n", partition.Part_start)
		fmt.Printf("  Size: %d\n", partition.Part_size)
		fmt.Printf("  Name: %s\n", partName)
		fmt.Printf("  Correlative: %d\n", partition.Part_correlative)
		fmt.Printf("  ID: %s\n", partID)
	}

}