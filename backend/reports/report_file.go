package reports

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"backend/structures"
	"backend/utils"
)

func ReportFile(superblock *structures.SuperBlock, diskPath string, filepath string, path string) error {
	// Crear directorios padres si no existen
	err := utils.CreateParentDirs(filepath)
	if err != nil {
		return fmt.Errorf("error al crear directorios padres: %w", err)
	}

	// Obtener nombres de archivos DOT e imagen
	dotFileName, outputImage := utils.GetFileNames(filepath)

	// Verificar si la ruta del archivo DOT es válida
	if dotFileName == "" || outputImage == "" {
		return fmt.Errorf("ruta de archivo DOT o imagen no válida")
	}

	// Obtener directorios padres y nombre del archivo
	parentsDir, destDir := utils.GetParentDirectories(path)
	fmt.Println("Generando reporte...")

	// Leer el contenido del archivo
	contentFile, err := superblock.ReadFile(diskPath, parentsDir, destDir)
	if err != nil {
		return fmt.Errorf("error al leer el archivo: %w", err)
	}

	// Eliminar caracteres nulos y espacios innecesarios
	contentFile = strings.TrimRight(contentFile, "\x00")
	contentFile = strings.TrimSpace(contentFile)

	// Escapar el contenido para HTML
	contentFile = escapeHTML(contentFile)

	// Crear contenido DOT
	dotContent := `digraph G {
        rankdir=LR; // Alineación horizontal
        node [shape=plaintext]
    `
	dotContent += fmt.Sprintf(`file [label=<
		<table border="0" cellborder="1" cellspacing="0">
			<tr><td colspan="2" bgcolor="blue"><font color="white"><b>ARCHIVO %s</b></font></td></tr>
			<tr><td bgcolor="lightgray"><b>Contenido</b></td><td>%s</td></tr>
		</table>
	>]
	`, destDir, contentFile)

	dotContent += `}`
	fmt.Println(dotContent)

	// Crear archivo DOT
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear archivo DOT: %w", err)
	}
	defer dotFile.Close()

	// Escribir contenido en el archivo DOT
	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo DOT: %w", err)
	}

	// Generar imagen usando Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	output, err := cmd.CombinedOutput() // Capturar salida estándar y de error
	if err != nil {
		return fmt.Errorf("error al generar la imagen: %w\nSalida de Graphviz:\n%s", err, string(output))
	}

	fmt.Println("Imagen del archivo generada:", outputImage)
	return nil
}