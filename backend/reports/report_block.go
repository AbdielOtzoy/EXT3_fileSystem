package reports

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"backend/structures"
	"backend/utils"
)

// Función para escapar caracteres especiales en HTML
func escapeHTML(content string) string {
	content = strings.ReplaceAll(content, "&", "&amp;")
	content = strings.ReplaceAll(content, "<", "&lt;")
	content = strings.ReplaceAll(content, ">", "&gt;")
	content = strings.ReplaceAll(content, "\"", "&quot;")
	content = strings.ReplaceAll(content, "'", "&apos;")
	return content
}

func ReportBlock(superblock *structures.SuperBlock, diskPath string, path string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	fmt.Println("Generating report...")

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
		rankdir=LR; // Alineación horizontal
		node [shape=plaintext]
	`

	// Variable para almacenar los nombres de los bloques y conectarlos
	var blockNames []string

	// Iterar sobre cada inodo
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structures.Inode{}

		err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return err
		}

		inode.Print()

		// Iterar sobre cada bloque del inodo
		for i, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}

			// Nombre del bloque
			blockName := fmt.Sprintf("block%d", blockIndex)
			fmt.Println("Block name:", blockName)
			blockNames = append(blockNames, blockName)

			// Manejar bloques de carpeta
			if inode.I_type[0] == '0' {
				// Manejar bloques de apuntadores
				if i == 12 || i == 13 || i == 14 {
					fmt.Println("Pointer block")
					block := &structures.PointerBlock{}
					err := block.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
					if err != nil {
						return err
					}

					fmt.Println("Block pointers:")
					block.Print()

					// Agregar la información del bloque
					dotContent += fmt.Sprintf("%s [label=<\n", blockName)
					dotContent += "<table border='0' cellborder='1' cellspacing='0'>\n"
					dotContent += fmt.Sprintf("<tr><td bgcolor='lightblue'><b>Block %d</b></td></tr>\n", blockIndex)
					
					for j, pointer := range block.P_pointers {
						dotContent += fmt.Sprintf("<tr><td>%d: %d</td></tr>\n", j+1, pointer)
					}

					dotContent += "</table>>];\n"

					// graficar los bloques a los que apunta el bloque de apuntadores
					for _, pointer := range block.P_pointers {
						if pointer == -1 {
							break
						}
						blockName := fmt.Sprintf("block%d", pointer)
						dotContent += fmt.Sprintf("%s -> %s;\n", blockName, blockName)
						blockNames = append(blockNames, blockName)
						block := &structures.FolderBlock{}
						err := block.Deserialize(diskPath, int64(superblock.S_block_start+(pointer*superblock.S_block_size)))
						if err != nil {
							return err
						}
						dotContent += fmt.Sprintf("%s [label=<\n", blockName)
						dotContent += "<table border='0' cellborder='1' cellspacing='0'>\n"
						dotContent += fmt.Sprintf("<tr><td bgcolor='lightblue'><b>Block %d</b></td></tr>\n", pointer)
						for j, content := range block.B_content {
							name := strings.TrimRight(string(content.B_name[:]), "\x00")
							dotContent += fmt.Sprintf("<tr><td>%d: %s</td></tr>\n", j+1, escapeHTML(name))
						}
						dotContent += "</table>>];\n"
						dotContent += fmt.Sprintf("%s -> %s;\n", blockName, blockName)
					}
					continue
				}
				fmt.Println("Folder block")
				block := &structures.FolderBlock{}

				err := block.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
				if err != nil {
					return err
				}

				// Agregar la información del bloque
				dotContent += fmt.Sprintf("%s [label=<\n", blockName)
				dotContent += "<table border='0' cellborder='1' cellspacing='0'>\n"
				dotContent += fmt.Sprintf("<tr><td bgcolor='lightblue'><b>Block %d</b></td></tr>\n", blockIndex)

				for j, content := range block.B_content {
					name := strings.TrimRight(string(content.B_name[:]), "\x00")
					dotContent += fmt.Sprintf("<tr><td>%d: %s</td></tr>\n", j+1, escapeHTML(name))
				}

				dotContent += "</table>>];\n"
				continue
			}

			// Manejar bloques de archivo
			if inode.I_type[0] == '1' {
				fmt.Println("File block")
				block := &structures.FileBlock{}

				err := block.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
				if err != nil {
					return err
				}

				block.Print()

				// Agregar la información del bloque
				dotContent += fmt.Sprintf("%s [label=<\n", blockName)
				dotContent += "<table border='0' cellborder='1' cellspacing='0'>\n"
				dotContent += fmt.Sprintf("<tr><td bgcolor='lightblue'><b>Block %d</b></td></tr>\n", blockIndex)

				content := strings.TrimRight(string(block.B_content[:]), "\x00")
				dotContent += fmt.Sprintf("<tr><td>%s</td></tr>\n", escapeHTML(content))

				dotContent += "</table>>];\n"
				continue
			}
		}
	}

	// Conectar los bloques con flechas
	for i := 0; i < len(blockNames)-1; i++ {
		dotContent += fmt.Sprintf("%s -> %s;\n", blockNames[i], blockNames[i+1])
	}

	fmt.Println("Generating image...")

	dotContent += "}"

	// Escribir el archivo DOT
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return err
	}
	defer dotFile.Close()

	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return err
	}

	// Generar la imagen usando Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al generar la imagen: %w", err)
	}

	fmt.Println("Image generated:", outputImage)
	return nil
}