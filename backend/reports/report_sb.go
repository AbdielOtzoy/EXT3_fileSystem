package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func ReportSuperBlock(sb *structures.SuperBlock, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := fmt.Sprintf(`digraph G {
		node [shape=plaintext]
		tabla [label=<
			<table border="0" cellborder="1" cellspacing="0">
				<tr><td colspan="2" bgcolor="blue"><font color="white"><b>REPORTE SUPER BLOQUE</b></font></td></tr>
				<tr><td bgcolor="lightgray"><b>s_filesystem_type</b></td><td>%s</td></tr>
				<tr><td bgcolor="lightgray"><b>s_inodes_count</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_blocks_count</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_free_blocks_count</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_free_inodes_count</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_mtime</b></td><td>%s</td></tr>
				<tr><td bgcolor="lightgray"><b>s_umtime</b></td><td>%s</td></tr>
				<tr><td bgcolor="lightgray"><b>s_mnt_count</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_magic</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_inode_size</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_block_size</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_first_ino</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_first_blo</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_bm_inode_start</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_bm_block_start</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_inode_start</b></td><td>%d</td></tr>
				<tr><td bgcolor="lightgray"><b>s_block_start</b></td><td>%d</td></tr>
			`, sb.S_filesystem_type, sb.S_inodes_count, sb.S_blocks_count, sb.S_free_blocks_count, sb.S_free_inodes_count, time.Unix(int64(sb.S_mtime), 0).Format(time.RFC3339), time.Unix(int64(sb.S_umtime), 0).Format(time.RFC3339), sb.S_mnt_count, sb.S_magic, sb.S_inode_size, sb.S_block_size, sb.S_first_ino, sb.S_first_blo, sb.S_bm_inode_start, sb.S_bm_block_start, sb.S_inode_start, sb.S_block_start)

	dotContent += "</table>>] }"

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

	fmt.Println("Reporte SuperBloque generado con éxito")
	return nil
}

