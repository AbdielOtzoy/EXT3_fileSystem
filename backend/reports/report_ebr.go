package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ReportEBR(ebr *structures.EBR, path string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)	
	
	dotContent := fmt.Sprintf(`digraph G {
		node [shape=plaintext]
		tabla [label=<
			<table border="0" cellborder="1" cellspacing="0">
				<tr><td colspan="2" bgcolor="blue"><font color="white"><b>REPORTE EBR</b></font></td></tr>
				<tr><td bgcolor="lightgray"><b>ebr_part_mount</b></td><td>%c</td></tr>
				<tr><td bgcolor="lightgray"><b>ebr_part_fit</b></td><td>%c</td></tr>
				<tr><td bgcolor="lightgray"><b>ebr_part_start</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>ebr_part_size</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>ebr_part_next</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>ebr_part_name</b></td><td>%s</td></tr>
			</table>>] }`, ebr.Ebr_part_mount[0], ebr.Ebr_part_fit[0], ebr.Ebr_part_start, ebr.Ebr_part_size, ebr.Ebr_part_next, strings.TrimRight(string(ebr.Ebr_part_name[:]), "\x00"))
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}

	defer file.Close()
	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando dot: %v", err)
	}

	fmt.Printf("Reporte EBR generado: %s\n", outputImage)
	return nil



}