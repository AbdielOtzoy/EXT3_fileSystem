package commands

import (
	structures "backend/structures"
	"backend/stores"
	"backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type MOUNT struct {
	path string
	name string
}

func ParseMount(tokens []string) (string, error) {
	cmd := &MOUNT{} // create the mkdisk command

	args := strings.Join(tokens, " ") // join the tokens to get the arguments

	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)

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
			case "-path":
				if value == "" {
					return "", errors.New("invalid path")
				}
				cmd.path = value
			case "-name":
				if value == "" {
					return "", errors.New("invalid name")
				}
				cmd.name = value
			default:
				return "", fmt.Errorf("invalid argument: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("missing path")
	}

	if cmd.name == "" {
		return "", errors.New("missing name")
	}

	err := commandMount(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("mounting partition %s in %s", cmd.name, cmd.path), nil
}

func commandMount(mount *MOUNT) error {
	var mbr structures.MBR
	
	err := mbr.DeserializeMBR(mount.path)
	if err != nil {
		fmt.Println("error deserializing mbr: ", err)
		return err
	}
	fmt.Println("name: ", mount.name)
	fmt.Println("path: ", mount.path)

	partition, indexPartition := mbr.GetPartitionByName(mount.name)
	if indexPartition == -1 {
		fmt.Println("partition not found")
		return errors.New("partition not found")
	}
	fmt.Println("indexPartition: ", indexPartition)	

	if stores.MountedPartitions[string(partition.Part_id[:])] != "" {
		fmt.Println("partition already mounted")
		return errors.New("partition already mounted")
	}

	if partition == nil {
		fmt.Println("partition not found")
		return errors.New("partition not found")
	}

	fmt.Println("Partition Available:")
	partition.PrintPartition()

	// Generate id for the partition
	idPartition, err := generatePartitionID(mount, indexPartition)
	if err != nil {
		fmt.Println("error generating partition id: ", err)
		return err 
	}

	stores.MountedPartitions[idPartition] = mount.path  // mount the partition

	partition.MountPartition(indexPartition, idPartition) // mount the partition

	fmt.Println("partition mounted successfully")
	partition.PrintPartition()

	// Save the changes in the mbr
	mbr.Mbr_partitions[indexPartition-1] = *partition

	err = mbr.SerializeMBR(mount.path)
	if err != nil {
		fmt.Println("error serializing mbr: ", err)
		return err
	}

	return nil
}

func generatePartitionID(mount *MOUNT, indexPartition int) (string, error) {
	letter, err := utils.GetLetter(mount.path)
	if err != nil {
		fmt.Println("error getting letter: ", err)
		return "", err
	}

	idPartition := fmt.Sprintf("%s%d%s", stores.Carnet, indexPartition, letter)

	return idPartition, nil
}