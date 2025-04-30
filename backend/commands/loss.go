package commands

import (
	"backend/stores"
	"backend/structures"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type LOSS struct {
	id string
}

func ParseLoss(tokens []string) (string, error) {
	cmd := &LOSS{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+`)
	matches := re.FindAllString(args, -1)

	if len(matches) != len(tokens) {
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])

		switch key {
		case "-id":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.id = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	err := commandLoss(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("LOSS: Archivo %s perdido exitosamente.", cmd.id), nil
}

func commandLoss(cmd *LOSS) error {
	sb, _, partitionPath, err := stores.GetMountedPartitionSuperblock(cmd.id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Buscar archivos
	err = Loss(partitionPath, sb, cmd.id)
	if err != nil {
		return fmt.Errorf("error al perder el archivo: %w", err)
	}

	return nil
}

func Loss(path string, sb *structures.SuperBlock, id string) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %w", err)
	}

	partitionFound, _ := mbr.GetPartitionByID(id)
	if partitionFound == nil {
		return fmt.Errorf("partición con ID %s no encontrada", id)
	}

	// Abrir el archivo para lectura y escritura
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %w", err)
	}
	defer file.Close()

	// Calcular el inicio y tamaño de la partición
	partitionStart := partitionFound.Part_start
	partitionSize := partitionFound.Part_size

	// Áreas a limpiar desde el SuperBlock
	cleanAreas := []struct {
		start  int32
		length int32
		name   string
	}{
		// Limpiar el superbloque mismo (primeros 68 bytes)
		{0, 68, "superbloque"},

		// Bitmaps y tablas
		{sb.S_bm_inode_start, sb.S_inodes_count, "bitmap de inodos"},
		{sb.S_bm_block_start, sb.S_blocks_count, "bitmap de bloques"},
		{sb.S_inode_start, sb.S_inodes_count * sb.S_inode_size, "área de inodos"},
		{sb.S_block_start, sb.S_blocks_count * sb.S_block_size, "área de bloques"},
	}

	// Limpiar cada área
	for _, area := range cleanAreas {
		// Posición absoluta en el archivo = inicio de partición + posición relativa
		position := partitionStart + area.start

		// Asegurarnos de no salirnos de la partición
		if position >= partitionStart+partitionSize {
			continue
		}

		// Ajustar el length si nos pasamos del tamaño de la partición
		length := area.length
		if position+length > partitionStart+partitionSize {
			length = (partitionStart + partitionSize) - position
		}

		// Crear un buffer de ceros del tamaño adecuado
		nullBuffer := make([]byte, length)

		// Posicionar el cursor en el archivo
		_, err := file.Seek(int64(position), 0)
		if err != nil {
			return fmt.Errorf("error al posicionar el cursor para %s: %w", area.name, err)
		}

		// Escribir los bytes nulos
		_, err = file.Write(nullBuffer)
		if err != nil {
			return fmt.Errorf("error al escribir %s: %w", area.name, err)
		}

		fmt.Printf("Se ha limpiado el %s (posición: %d, tamaño: %d)\n", area.name, position, length)
	}

	// Resetear los valores del superbloque
	sb.S_inodes_count = 0
	sb.S_blocks_count = 0
	sb.S_free_inodes_count = 0
	sb.S_free_blocks_count = 0
	sb.S_mnt_count = 0
	sb.S_magic = 0

	// Reescribir el superbloque limpio
	err = sb.Serialize(path, int64(partitionFound.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	fmt.Printf("Se ha simulado correctamente la pérdida completa del sistema de archivos en la partición %s\n", id)
	return nil
}
