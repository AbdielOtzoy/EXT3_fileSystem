package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
)

func ReportDisk(mbr *structures.MBR, diskPath string, path string) error {
	fmt.Println("Reportando Disco")
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	storageUsed := 0
	diskSize := mbr.Mbr_size

	// Crear el archivo DOT
	dotFileName, outputImage := utils.GetFileNames(path)
	dotContent := `digraph G {
		rankdir=LR; // Orientación horizontal
		node [shape=plaintext]
		disco [label=<
			<table border="1" cellborder="1" cellspacing="0">
				<tr><td colspan="4" bgcolor="blue"><font color="white"><b>DISCO</b></font></td></tr>
				<tr><td bgcolor="lightgray"><b>MBR</b></td><td bgcolor="lightgray"><b>Particiones</b></td><td bgcolor="lightgray"><b>Espacio Libre</b></td></tr>
	`

	// Recorrer las particiones del MBR
	for i, part := range mbr.Mbr_partitions {
		partType := rune(part.Part_type[0])
		partSize := part.Part_size

		if partType == 'P' || partType == 'E' {
			storageUsed += int(partSize)
			partitionName := ""
			if partType == 'P' {
				partitionName = "Primaria"
			} else if partType == 'E' {
				partitionName = "Extendida"
			}

			// Agregar la partición al archivo DOT
			dotContent += fmt.Sprintf(`
				<tr>
					<td colspan="4" bgcolor="yellow"><b>PARTICIÓN %d (%s)</b></td>
				</tr>
				<tr>
					<td>%s</td>
					<td>%d bytes</td>
					<td>%.2f%%</td>
				</tr>
			`, i+1, partitionName, partitionName, partSize, (float32(partSize)/float32(diskSize))*100)

			// Si es una partición extendida, recorrer los EBRs
			if partType == 'E' {
				ebr := structures.EBR{}
				err := ebr.DeserializeEBR(diskPath, int64(part.Part_start))
				if err != nil {
					return fmt.Errorf("error al deserializar el EBR: %v", err)
				}

				storageUsedExtended := int(ebr.Ebr_part_size)
				dotContent += fmt.Sprintf(`
					<tr>
						<td colspan="4" bgcolor="orange"><b>EBR INICIAL</b></td>
					</tr>
					<tr>
						<td>EBR</td>
						<td>%d bytes</td>
						<td>%.2f%%</td>
					</tr>
				`, ebr.Ebr_part_size, (float32(ebr.Ebr_part_size)/float32(diskSize)*100))

				ebrNext := ebr.Ebr_part_next
				for ebrNext != -1 {
					err := ebr.DeserializeEBR(diskPath, int64(ebrNext))
					if err != nil {
						return fmt.Errorf("error al deserializar el EBR: %v", err)
					}

					storageUsedExtended += int(ebr.Ebr_part_size)
					dotContent += fmt.Sprintf(`
						<tr>
							<td colspan="4" bgcolor="orange"><b>EBR SIGUIENTE</b></td>
						</tr>
						<tr>
							<td>EBR</td>
							<td>%d bytes</td>
							<td>%.2f%%</td>
						</tr>
					`, ebr.Ebr_part_size, (float32(ebr.Ebr_part_size)/float32(diskSize)*100))

					ebrNext = ebr.Ebr_part_next
				}

				// Espacio libre en la partición extendida
				freeSpaceExtended := partSize - int32(storageUsedExtended)
				dotContent += fmt.Sprintf(`
					<tr>
						<td colspan="4" bgcolor="lightgreen"><b>ESPACIO LIBRE (Extendida)</b></td>
					</tr>
					<tr>
						<td>Libre</td>
						<td>%d bytes</td>
						<td>%.2f%%</td>
					</tr>
				`, freeSpaceExtended, (float32(freeSpaceExtended)/float32(diskSize))*100)
			}
		}
	}

	// Espacio libre en el disco
	freeSpace := diskSize - int32(storageUsed)
	dotContent += fmt.Sprintf(`
		<tr>
			<td colspan="4" bgcolor="lightgreen"><b>ESPACIO LIBRE (Disco)</b></td>
		</tr>
		<tr>
			<td>Libre</td>
			<td>%d bytes</td>
			<td>%.2f%%</td>
		</tr>
	`, freeSpace, (float32(freeSpace)/float32(diskSize)*100))

	dotContent += `</table>>]; }`

	// Crear el archivo DOT
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	// Generar la imagen usando Graphviz
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)
	}

	fmt.Println("Imagen del disco generada:", outputImage)
	return nil
}