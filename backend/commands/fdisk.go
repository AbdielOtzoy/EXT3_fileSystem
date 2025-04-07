package commands

import (
	structures "backend/structures"
	utils "backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)


type FDISK struct {
	Size 	int
	unit 	string
	path 	string
	type_ 	string
	fit 	string
	name 	string

}

func ParseFdisk(tokens []string) (string, error) {
	cmd := &FDISK{} // create the mkdisk command

	args := strings.Join(tokens, " ") // join the tokens to get the arguments

	re := regexp.MustCompile(`-size=\d+|-unit=[kKmMbB]|-fit=[bBfF]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)

	matches := re.FindAllString(args, -1) // find all the matches

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2) // split the match in key and value
		if len(kv) != 2 {
			return "", fmt.Errorf("invalid argument: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
			case "-size":
				size, err := strconv.Atoi(value) // convert the value to integer
				if err != nil || size <= 0 {
					return "", errors.New("invalid size")
				}
				cmd.Size = size
			case "-unit":
				if value != "K" && value != "M" && value != "B" && value != "k" && value != "m" && value != "b" {
					// K = Kilobytes, M = Megabytes, B = Bytes
					return "", errors.New("invalid unit")
				}
				cmd.unit = strings.ToUpper(value)
			case "-fit":
				value = strings.ToUpper(value)
				if value != "BF" && value != "FF" && value != "WF" {
					return "", errors.New("invalid fit")
				}
				cmd.fit = value
			case "-path":
				if value == "" {
					return "", errors.New("invalid path")
				}
				cmd.path = value
			case "-type":
				value = strings.ToUpper(value)
				if value != "P" && value != "E" && value != "L" {
					return "", errors.New("invalid type")
				}
				cmd.type_ = value
			case "-name":
				if value == "" {
					return "", errors.New("invalid name")
				}
				cmd.name = value
			default:
				return "", fmt.Errorf("invalid argument: %s", key)

		}
	}

	if cmd.Size == 0 {
		return "", errors.New("size is required")
	}

	if cmd.path == "" {
		return "", errors.New("path is required")
	}

	if cmd.unit == "" {
		cmd.unit = "M"
	}

	if cmd.fit == "" {
		cmd.fit = "FF"
	}

	if cmd.type_ == "" {
		cmd.type_ = "P"
	}

	if cmd.name == "" {
		return "", errors.New("name is required")
	}

	err := commandFdisk(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Partition %s created successfully", cmd.name), nil
}

func commandFdisk(fdisk *FDISK) error {
	sizeBytes, err := utils.ConvertToBytes(fdisk.Size, fdisk.unit) // convert the size to bytes
	fmt.Println("size:", fdisk.Size)
	fmt.Println("unit:", fdisk.unit)
	fmt.Println("Size in bytes:", sizeBytes)
	if err != nil {
		fmt.Println("error converting size to bytes:", err)
		return err
	}

	if fdisk.type_ == "P" {
		err = createPrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("error creating primary partition:", err)
			return err
		}
	} else if fdisk.type_ == "E" {
		fmt.Println("extended partition")	
		err = createExtendedPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("error creating extended partition:", err)
			return err
		} 
	} else if fdisk.type_ == "L" {
		fmt.Println("logical partition")
		 err = createLogicalPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("error creating logical partition:", err)
			return err
		} 
	}

	return nil
}

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR // create the MBR struct

	err := mbr.DeserializeMBR(fdisk.path) // read the MBR
	if err != nil {
		fmt.Println("error reading MBR:", err)
		return err
	}

	// check if is there is space for the partition
	freeSpace := mbr.GetFreeSpace()
	if freeSpace < int32(sizeBytes) {
		fmt.Println("not enough space in the MBR")
		return errors.New("not enough space in the disk")
	}
	fmt.Println("MBR before creating primary partition")
	mbr.PrintMBR()

	// Get the first free partition
	availablePartition, startPartition, index := mbr.GetFreePartition()
	if availablePartition == nil {
		fmt.Println("no available partition")
		return errors.New("no available partition")
	}

	fmt.Println("Available partition")
	availablePartition.PrintPartition()

	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.type_, fdisk.fit, fdisk.name)

	fmt.Println("MBR after creating primary partition")
	availablePartition.PrintPartition()	

	if availablePartition != nil {
		mbr.Mbr_partitions[index] = *availablePartition // update the partition
	}

	// Print the partitions
	fmt.Println("\nParticiones del MBR:")
	mbr.PrintPartitions()

	err = mbr.SerializeMBR(fdisk.path) // serialize the MBR
	if err != nil {
		fmt.Println("error serializing MBR", err)
	}

	return nil
}

func createExtendedPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		fmt.Println("error reading MBR:", err)
	}

	// Check if there is an extended partition
	extendedPartition, _ := mbr.GetExtendedPartition()
	if extendedPartition != nil {
		fmt.Println("extended partition already exists")
		return errors.New("extended partition already exists")
	}
	// Check if there is space for the extended partition
	availablePartition, _, _ := mbr.GetFreePartition()
	if availablePartition == nil {
		fmt.Println("no available partition")
		return errors.New("no available partition")
	}

	fmt.Println("MBR before creating extended partition")
	mbr.PrintMBR()

	availablePartition, startPartition, index := mbr.GetFreePartition()
	if availablePartition == nil {
		fmt.Println("no available partition")
		return errors.New("no available partition")
	}

	fmt.Println("Available partition")
	availablePartition.PrintPartition()

	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.type_, fdisk.fit, fdisk.name)

	fmt.Println("MBR after creating extended partition")
	availablePartition.PrintPartition()

	if availablePartition != nil {
		mbr.Mbr_partitions[index] = *availablePartition
	}

	ebr := &structures.EBR{
		Ebr_part_mount: [1]byte{'N'},
		Ebr_part_fit:   [1]byte{fdisk.fit[0]},
		Ebr_part_start: int32(startPartition),
		Ebr_part_size:  -1,
		Ebr_part_next:  -1,
		Ebr_part_name:  [16]byte{},
	}
	copy(ebr.Ebr_part_name[:], fdisk.name)
	fmt.Println("EBR has been created")
	ebr.PrintEBR()

	err = ebr.SerializeEBR(fdisk.path, int64(startPartition))
	if err != nil {
		fmt.Println("error serializing EBR:", err)
	}

	fmt.Println("\nParticiones del MBR:")	
	mbr.PrintPartitions()

	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		fmt.Println("error serializing MBR", err)
	}


	if err != nil {
		fmt.Println("error creating EBR:", err)
	}

	return nil
}

func createLogicalPartition(fdisk *FDISK, logicalSize int) error {
	// 1. Read the MBR
	var mbr structures.MBR

	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		fmt.Println("error reading MBR:", err)
		return err
	}

	fmt.Println("MBR before creating logical partition")
	mbr.PrintMBR()

	// 2. Get the extended partition, if not exists return an error
	extendedPartition, _ := mbr.GetExtendedPartition()

	if extendedPartition == nil {
		fmt.Println("no extended partition")
		return errors.New("no extended partition")
	}

	fmt.Println("Extended partition")
	extendedPartition.PrintPartition()

	// Get the first ebr in the extendened partition, the first ebr is in the extendedPartition.Part_start + 32
	ebr := &structures.EBR{}
	err = ebr.DeserializeEBR(fdisk.path, int64(extendedPartition.Part_start))
	if err != nil {
		fmt.Println("error reading EBR:", err)
	}

	fmt.Println("First EBR")
	ebr.PrintEBR()

	lastEBR, _, err := extendedPartition.GetLastEBR(ebr, fdisk.path)
	fmt.Println("Last EBR")
	lastEBR.PrintEBR()

	// Calculate the start of the new EBR
	newEbrStart := lastEBR.Ebr_part_start + structures.EBRSize + int32(logicalSize)
	fmt.Println("New EBR start:", newEbrStart)

	// Ensure that the new EBR does not overlap with other partitions or EBRs
	if newEbrStart > extendedPartition.Part_start+extendedPartition.Part_size {
		fmt.Println("not enough space in the extended partition")
		return errors.New("not enough space in the extended partition")
	}

	// Create the new EBR
	newEbr := &structures.EBR{
		Ebr_part_mount: [1]byte{'N'},
		Ebr_part_fit:   [1]byte{fdisk.fit[0]},
		Ebr_part_start: newEbrStart,
		Ebr_part_size:  int32(logicalSize),
		Ebr_part_next:  -1,
		Ebr_part_name:  [16]byte{},
	}

	copy(newEbr.Ebr_part_name[:], fdisk.name)

	fmt.Println("New EBR")
	newEbr.PrintEBR()

	// Update the last EBR
	lastEBR.Ebr_part_next = newEbr.Ebr_part_start
	lastEBR.Ebr_part_size = int32(logicalSize)
	fmt.Println("Last EBR after update")
	lastEBR.PrintEBR()

	// Serialize the new EBR
	err = newEbr.SerializeEBR(fdisk.path, int64(newEbr.Ebr_part_start))
	if err != nil {
		fmt.Println("error serializing new EBR:", err)
		return err
	}

	// Serialize the last EBR
	err = lastEBR.SerializeEBR(fdisk.path, int64(lastEBR.Ebr_part_start))
	if err != nil {
		fmt.Println("error serializing last EBR:", err)
		return err
	}

	

	// Serialize the MBR
	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		fmt.Println("error serializing MBR", err)
		return err
	}

	fmt.Println("Logical partition created successfully")
	return nil
}