package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func ReportLS(sb *structures.SuperBlock, diskPath string, path string, path_file_ls string) error {
	fmt.Println("Generando Reporte LS")
	
	// 1. Verificar y crear directorios padres
	err := utils.CreateParentDirs(path)
	if err != nil {
		return fmt.Errorf("error al crear directorios padres: %v", err)
	}

	// 2. Obtener nombres de archivos
	dotFileName, outputImage := utils.GetFileNames(path)
	
	// Asegurar que la extensión sea .dot
	if !strings.HasSuffix(dotFileName, ".dot") {
		dotFileName += ".dot"
	}

	// 3. Obtener directorios padres y destino
	parentDirs, destDir := utils.GetParentDirectories(path_file_ls)
	fmt.Printf("Buscando en: %v -> %s\n", parentDirs, destDir)

	// 4. Obtener el inodo del directorio
	indexInode, err := sb.GetInode(diskPath, parentDirs, destDir)
	if err != nil {
		return fmt.Errorf("error al obtener inodo: %v", err)
	}

	// 5. Leer el inodo del directorio
	inode := &structures.Inode{}
	err = inode.Deserialize(diskPath, int64(sb.S_inode_start+(indexInode*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo: %v", err)
	}

	// 6. Generar contenido DOT
	dotContent := generateDotContent(sb, diskPath, inode)

	// 7. Escribir archivo DOT
	if err := writeDotFile(dotFileName, dotContent); err != nil {
		return err
	}

	// 8. Generar imagen con Graphviz
	if err := generateGraphvizImage(dotFileName, outputImage); err != nil {
		return fmt.Errorf("error al generar imagen: %v. Asegúrese que Graphviz está instalado y en el PATH", err)
	}

	fmt.Printf("Reporte LS generado exitosamente: %s\n", outputImage)
	return nil
}

// generateDotContent genera el contenido DOT para el reporte LS
func generateDotContent(sb *structures.SuperBlock, diskPath string, inode *structures.Inode) string {
	var builder strings.Builder

	builder.WriteString(`digraph G {
		node [shape=plaintext]
		tabla [label=<
			<table border="0" cellborder="1" cellspacing="0">
				<tr><td colspan="5" bgcolor="blue"><font color="white"><b>REPORTE LS</b></font></td></tr>
				<tr><td bgcolor="lightgray"><b>Nombre</b></td><td bgcolor="lightgray"><b>UID</b></td><td bgcolor="lightgray"><b>GID</b></td><td bgcolor="lightgray"><b>Size</b></td><td bgcolor="lightgray"><b>Tipo</b></td><td bgcolor="lightgray"><b>Fecha</b></td><td bgcolor="lightgray"><b>Hora</b></td><td bgcolor="lightgray"><b>Permisos</b></td></tr>
	`)

	// Procesar bloques del inodo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}

		block := &structures.FolderBlock{}
		if err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) 
		err != nil {
			continue
		}

		// Procesar contenidos del bloque (empezando desde 2 para omitir . y ..)
		for _, content := range block.B_content[2:] {
			if content.B_inodo == -1 {
				continue
			}

			// Obtener información del inodo
			inodeContent := &structures.Inode{}
			if err := inodeContent.Deserialize(diskPath, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
			err != nil {
				continue
			}

			// Limpiar nombre y determinar tipo
			name := strings.TrimRight(string(content.B_name[:]), "\x00")
			fileType := "Archivo"
			if inodeContent.I_type[0] == '0' {
				fileType = "Carpeta"
			}
			ctime := time.Unix(int64(inode.I_ctime), 0)

			date, time := utils.FormatDate(ctime.Format(time.RFC3339))

			// Agregar fila a la tabla
			builder.WriteString(fmt.Sprintf(`
				<tr>
					<td>%s</td>
					<td>%d</td>
					<td>%d</td>
					<td>%d</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s</td>
				</tr>
			`, name, inodeContent.I_uid, inodeContent.I_gid, inodeContent.I_size, fileType, date, time, inodeContent.I_perm))
		}
	}

	builder.WriteString("</table>>] }")
	return builder.String()
}

// writeDotFile escribe el contenido DOT en un archivo
func writeDotFile(filename, content string) error {
	// Asegurar que el directorio existe
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("error al crear directorio para archivo DOT: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error al crear archivo DOT: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("error al escribir en archivo DOT: %v", err)
	}

	return nil
}

// generateGraphvizImage genera la imagen a partir del archivo DOT
func generateGraphvizImage(dotFile, outputImage string) error {
	// Verificar si Graphviz está instalado
	if _, err := exec.LookPath("dot"); err != nil {
		return fmt.Errorf("graphviz no está instalado o no está en el PATH: %v", err)
	}

	cmd := exec.Command("dot", "-Tpng", dotFile, "-o", outputImage)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error al ejecutar dot: %s - %v", string(output), err)
	}

	return nil
}