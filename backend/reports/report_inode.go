package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func ReportInode(superblock *structures.SuperBlock, diskPath string, path string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
        rankdir=LR; // Alineación horizontal
        node [shape=plaintext]
    `

	// Iterar sobre cada inodo
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structures.Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return err
		}

		// Convertir tiempos a string
		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

		// Definir el contenido DOT para el inodo actual
		dotContent += fmt.Sprintf(`inode%d [label=<
            <table border="0" cellborder="1" cellspacing="0">
                <tr><td colspan="2" bgcolor="blue"><font color="white"><b>INODO %d</b></font></td></tr>
                <tr><td bgcolor="lightgray"><b>i_uid</b></td><td>%d</td></tr>
                <tr><td bgcolor="lightgray"><b>i_gid</b></td><td>%d</td></tr>
                <tr><td bgcolor="lightgray"><b>i_size</b></td><td>%d</td></tr>
                <tr><td bgcolor="lightgray"><b>i_atime</b></td><td>%s</td></tr>
                <tr><td bgcolor="lightgray"><b>i_ctime</b></td><td>%s</td></tr>
                <tr><td bgcolor="lightgray"><b>i_mtime</b></td><td>%s</td></tr>
                <tr><td bgcolor="lightgray"><b>i_type</b></td><td>%c</td></tr>
                <tr><td bgcolor="lightgray"><b>i_perm</b></td><td>%s</td></tr>
                <tr><td colspan="2" bgcolor="green"><b>BLOQUES DIRECTOS</b></td></tr>
            `, i, i, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))

		// Agregar los bloques directos a la tabla hasta el índice 11
		for j, block := range inode.I_block {
			if j > 11 {
				break
			}
			dotContent += fmt.Sprintf("<tr><td bgcolor='lightgreen'>%d</td><td>%d</td></tr>", j+1, block)
		}

		// Agregar los bloques indirectos a la tabla
		dotContent += fmt.Sprintf(`
                <tr><td colspan="2" bgcolor="yellow"><b>BLOQUES INDIRECTOS</b></td></tr>
                <tr><td>Simple</td><td>%d</td></tr>
                <tr><td>Doble</td><td>%d</td></tr>
                <tr><td>Triple</td><td>%d</td></tr>
            </table>>];
        `, inode.I_block[12], inode.I_block[13], inode.I_block[14])

		// Agregar enlace al siguiente inodo si no es el último
		if i < superblock.S_inodes_count-1 {
			dotContent += fmt.Sprintf("inode%d -> inode%d [color=red, penwidth=2];\n", i, i+1)
		}
	}

	dotContent += "}"

	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return err
	}
	defer dotFile.Close()

	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return err
	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("Imagen de los inodos generada:", outputImage)
	return nil
}
