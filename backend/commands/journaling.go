package commands

import (
	"backend/stores"
	"backend/structures"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type JOURNAL struct {
	id string
}

func ParseJournal(tokens []string) (string, error) {
	cmd := &JOURNAL{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", errors.New("formato de parámetro inválido")
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	jsonContent, err := commandJournal(cmd)
	if err != nil {
		return "", err
	}

	return jsonContent, nil
}

func commandJournal(cmd *JOURNAL) (string, error) {
	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(cmd.id)
	if err != nil {
		return "", fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// verificar que el sistema de archivos sea ext3
	if partitionSuperblock.S_filesystem_type != 3 {
		return "", errors.New("sistema de archivos no es ext3")
	}

	// Recolectará todos los journals
	journals := []map[string]string{}

	count := 2

	for {
		// obtener el journal
		journal := &structures.Journal{}

		fmt.Println("Deserializando en:", int64(mountedPartition.Part_start)+68+114*int64(count))
		// Deserializar el journal
		err = journal.Deserialize(partitionPath, int64(mountedPartition.Part_start)+68+114*int64(count))
		if err != nil {
			return "", fmt.Errorf("error al deserializar el journal: %w", err)
		}
		if journal.J_count == 0 {
			break
		}
		fmt.Println("obtenido el journal")
		journal.Print()
		date := time.Unix(int64(journal.J_content.I_date), 0)

		operation := strings.TrimRight(string(journal.J_content.I_operation[:]), "\x00")
		path := strings.TrimRight(string(journal.J_content.I_path[:]), "\x00")
		content := strings.TrimRight(string(journal.J_content.I_content[:]), "\x00")

		// Crear un mapa para el entry actual
		journalEntry := map[string]string{
			"operation": operation,
			"date":      date.Format(time.RFC3339),
			"path":      path,
			"content":   content,
		}

		if operation == "mkdir" {
			fmt.Println("Operación mkdir detectada")
		} else if operation == "mkfile" {
			fmt.Println("Operación mkfile detectada")
		}

		// Añadir al slice de journals
		journals = append(journals, journalEntry)

		count++
	}

	// Serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return "", fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	// Convertir la estructura a JSON usando encoding/json para garantizar una sintaxis correcta
	jsonBytes, err := json.MarshalIndent(journals, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error al generar JSON: %w", err)
	}

	return string(jsonBytes), nil
}
