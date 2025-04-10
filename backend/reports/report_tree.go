package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Recursive function to generate dot content for an inode and its blocks
func generateInodeTreeContent(superblock *structures.SuperBlock, diskPath string, inodeIndex int32, visited map[int32]bool) (string, string, error) {
	if visited[inodeIndex] {
		return "", "", nil
	}
	visited[inodeIndex] = true

	// Deserialize the current inode
	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+(inodeIndex*superblock.S_inode_size)))
	if err != nil {
		return "", "", err
	}

	// Convertir tiempos a string
	atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
	ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
	mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

	// Generate dot content for this inode with all information
	nodeContent := fmt.Sprintf(`inode%d [label=<
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
			<tr><td colspan="2" bgcolor="lightgreen"><b>Bloques Directos</b></td></tr>
	`, inodeIndex, inodeIndex, inode.I_uid, inode.I_gid, inode.I_size,
		atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))

	// Add direct blocks (0-11)
	for i := 0; i < 12; i++ {
		nodeContent += fmt.Sprintf("<tr><td>%d</td><td>%d</td></tr>", i, inode.I_block[i])
	}

	// Add indirect blocks (12-14)
	nodeContent += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="lightyellow"><b>Bloques Indirectos</b></td></tr>
			<tr><td>Simple</td><td>%d</td></tr>
			<tr><td>Doble</td><td>%d</td></tr>
			<tr><td>Triple</td><td>%d</td></tr>
		</table>>];
	`, inode.I_block[12], inode.I_block[13], inode.I_block[14])

	var connections string
	var blockNodes string

	// Process blocks
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		if inode.I_type[0] == '0' { // Folder
			folderBlock := &structures.FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
			if err != nil {
				return "", "", err
			}

			// Generate folder block node with all content
			blockNode := fmt.Sprintf(`block%d [label=<
				<table border="0" cellborder="1" cellspacing="0">
					<tr><td colspan="4" bgcolor="lightgreen"><b>Bloque Carpeta %d</b></td></tr>
					<tr>
						<td bgcolor="gray"><b>#</b></td>
						<td bgcolor="gray"><b>Nombre</b></td>
						<td bgcolor="gray"><b>Inodo</b></td>
						<td bgcolor="gray"><b>Tipo</b></td>
					</tr>
				`, blockIndex, blockIndex)

			// Add folder content (excluding parent references)
			for i, content := range folderBlock.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if name == "" || content.B_inodo == -1 {
					continue
				}
				// Skip parent directory references ("..")
				if name == ".." {
					continue
				}

				entryType := "Archivo"
				if content.B_inodo != -1 {
					// Check if the pointed inode is a directory
					childInode := &structures.Inode{}
					err := childInode.Deserialize(diskPath, int64(superblock.S_inode_start+(content.B_inodo*superblock.S_inode_size)))
					if err == nil && childInode.I_type[0] == '0' {
						entryType = "Carpeta"
					}
				}

				blockNode += fmt.Sprintf(`
					<tr>
						<td>%d</td>
						<td>%s</td>
						<td>%d</td>
						<td>%s</td>
					</tr>`, i+1, escape(name), content.B_inodo, entryType)
			}
			blockNode += "</table>>];\n"
			blockNodes += blockNode

			// Generate connections from block to inodes (excluding parent references)
			for _, content := range folderBlock.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if content.B_inodo == -1 || name == ".." {
					continue
				}
				connections += fmt.Sprintf("block%d -> inode%d [color=black];\n", blockIndex, content.B_inodo)
			}

		} else if inode.I_type[0] == '1' { // File
			fileBlock := &structures.FileBlock{}
			err := fileBlock.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
			if err != nil {
				return "", "", err
			}

			// Generate file block node
			content := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			blockNodes += fmt.Sprintf(`block%d [label=<
				<table border="0" cellborder="1" cellspacing="0">
					<tr><td bgcolor="lightblue"><b>Bloque Archivo %d</b></td></tr>
					<tr><td>%s</td></tr>
				</table>>];
			`, blockIndex, blockIndex, escapeHTML(content))
		}

		// Connect inode to its block (all in black)
		connections += fmt.Sprintf("inode%d -> block%d [color=black];\n", inodeIndex, blockIndex)
	}

	// Process child inodes recursively (for folder inodes)
	if inode.I_type[0] == '0' {
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				continue
			}

			folderBlock := &structures.FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(superblock.S_block_start+(blockIndex*superblock.S_block_size)))
			if err != nil {
				return "", "", err
			}

			for _, content := range folderBlock.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if content.B_inodo == -1 || content.B_inodo == inodeIndex || name == ".." {
					continue
				}

				childNodes, childConnections, err := generateInodeTreeContent(superblock, diskPath, content.B_inodo, visited)
				if err != nil {
					return "", "", err
				}
				nodeContent += childNodes
				blockNodes += childConnections
			}
		}
	}

	return nodeContent + blockNodes, connections, nil
}

func ReportTree(superblock *structures.SuperBlock, diskPath string, path string) error {
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	dotContent := `digraph G {
		rankdir=TB; // Top to bottom layout
		node [shape=plaintext]
	`

	// Track visited inodes to prevent infinite recursion
	visited := make(map[int32]bool)

	// Generate all nodes first, then connections
	nodes, connections, err := generateInodeTreeContent(superblock, diskPath, 0, visited)
	if err != nil {
		return err
	}

	dotContent += nodes + connections
	dotContent += "}"

	// Write DOT file
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return err
	}
	defer dotFile.Close()

	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return err
	}

	// Generate image using Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error generating image: %w", err)
	}

	fmt.Println("Filesystem tree image generated:", outputImage)
	return nil
}

// Helper function to escape HTML special characters
func escape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	).Replace(s)
}
