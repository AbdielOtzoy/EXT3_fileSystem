package commands

import (
	reports "backend/reports"
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type REP struct {
	id           string
	path         string
	name         string
	path_file_ls string
}

func ParseRep(tokens []string) (string, error) {
	cmd := &REP{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`(?i)-id=[^\s]+|-path="[^"]+"|-path=[^\s]+|-name=[^\s]+|-path_file_ls="[^"]+"|-path_file_ls=[^\s]+`)

	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("invalid argument: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("invalid id")
			}
			cmd.id = value
		case "-path":
			if value == "" {
				return "", errors.New("invalid path")
			}
			cmd.path = value
		case "-name":
			validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "sb", "file", "ls", "tree"}
			if !contains(validNames, value) {
				return "", errors.New("nombre inválido, debe ser uno de los siguientes: mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls")
			}
			cmd.name = value
		case "-path_file_ls":
			cmd.path_file_ls = value
		default:
			return "", fmt.Errorf("invalid argument: %s", key)
		}
	}
	if cmd.id == "" || cmd.path == "" || cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -id, -path, -name")
	}

	message, err := commandRep(cmd)
	if err != nil {
		return "", fmt.Errorf("Error al generar el reporte: %v", err)
	}

	return fmt.Sprintf("Se ha generado el reporte: %s", message), nil
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func commandRep(rep *REP) (string, error) {
	mountedMbr, mountedSb, mountedDiskPath, err := stores.GetMountedPartitionRep(rep.id)
	fmt.Println("mountedDiskPath", mountedDiskPath)
	if err != nil {
		return "", fmt.Errorf("Error al obtener la partición montada: %v", err)
	}

	switch rep.name {
	case "mbr":
		err = reports.ReportMBR(mountedMbr, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("MBR report generated at %s", rep.path), nil
	case "inode":
		err = reports.ReportInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("Inode report generated at %s", rep.path), nil
	case "block":
		err = reports.ReportBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("Block report generated at %s", rep.path), nil
	case "bm_inode":
		err = reports.ReportBMInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("BM Inode report generated at %s", rep.path), nil
	case "bm_block":
		err = reports.ReportBMBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("BM Block report generated at %s", rep.path), nil
	case "file":
		err = reports.ReportFile(mountedSb, mountedDiskPath, rep.path, rep.path_file_ls)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("File report generated at %s", rep.path), nil
	case "sb":
		err = reports.ReportSuperBlock(mountedSb, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("Superblock report generated at %s", rep.path), nil
	case "disk":
		err = reports.ReportDisk(mountedMbr, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		return fmt.Sprintf("Disk report generated at %s", rep.path), nil
	case "tree":
		err = reports.ReportTree(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		return fmt.Sprintf("Tree report generated at %s", rep.path), nil
	case "ls":
		err = reports.ReportLS(mountedSb, mountedDiskPath, rep.path, rep.path_file_ls)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Sprintf("LS report generated at %s", rep.path), nil

	}

	return "", fmt.Errorf("Invalid report name: %s", rep.name)
}
