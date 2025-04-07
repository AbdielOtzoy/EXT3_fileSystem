package stores

import (
	structures "backend/structures"
	"errors"
)

const Carnet string = "50" // 202300350

var (
	MountedPartitions map[string]string = make(map[string]string)
	userNameLogged string = ""
	idMountedPartition string = ""
	userid int32 = -1
	groupid int32 = -1

)

func SetSession(user string, idPartition string, uid int32, gid int32) {
	userNameLogged = user
	idMountedPartition = idPartition
	userid = uid
	groupid = gid
}

func GetSession() (string, string, int32, int32) {	
	return userNameLogged, idMountedPartition, userid, groupid
}


func GetMountedPartition(id string) (*structures.Partition, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, "", errors.New("la partición no está montada")
	}

	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, "", err
	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, "", err
	}

	return partition, path, nil
}

func GetMountedPartitionRep(id string) (*structures.MBR, *structures.SuperBlock, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la partición no está montada")
	}

	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	var sb structures.SuperBlock

	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &mbr, &sb, path, nil
}

// print the mounted partitions id
func PrintMountedPartitions() {
	for id, path := range MountedPartitions {
		println("id:", id, "path:", path)
	}
}

// return string[] with the mounted partitions
func GetMountedPartitions() []string {
	var mountedPartitions []string
	for id := range MountedPartitions {
		mountedPartitions = append(mountedPartitions, id)
	}
	return mountedPartitions
}

// GetMountedPartitionSuperblock obtiene el SuperBlock de la partición montada con el id especificado
func GetMountedPartitionSuperblock(id string) (*structures.SuperBlock, *structures.Partition, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la partición no está montada")
	}

	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	var sb structures.SuperBlock

	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &sb, partition, path, nil
}