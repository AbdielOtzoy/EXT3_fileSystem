package commands
import (
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type MKUSR struct {
	user string
	pass string
	group string
}

func ParseMkuser(tokens []string) (string, error) {
	cmd := &MKUSR{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user=[^\s]+|-pass=[^\s]+|-grp=[^\s]+`)
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
		case "-user":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.user = value
		case "-pass":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.pass = value
		case "-grp":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.group = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// el user, pass y group son obligatorios y deben tener como maximo 10 caracteres
	if cmd.user == "" || cmd.pass == "" || cmd.group == "" {
		return "", errors.New("faltan parámetros requeridos: -user, -pass, -group")
	}
	if len(cmd.user) > 10 || len(cmd.pass) > 10 || len(cmd.group) > 10 {
		return "", errors.New("los parámetros -user, -pass y -group deben tener como máximo 10 caracteres")
	}

	return createUser(cmd.user, cmd.pass, cmd.group), nil
}

func createUser(user string, pass string, group string) string {
	// get the current session
	userName, idPartition, _, _ := stores.GetSession()
	if userName == "" {
		return "No hay sesión activa"
	}

	// veriricar que sea el usuario root
	if userName != "root" {
		return "No tienes permisos para crear un usuario"
	}

	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return "Error al obtener el superbloque de la partición: %w"
	}

	err = partitionSuperblock.CreateUser(user, pass, group, partitionPath)
	if err != nil {
		return fmt.Sprintf("Error al crear el usuario: %s", err)
	}

	// serializar el superbloque
	err = partitionSuperblock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Sprintf("Error al serializar el superbloque: %s", err)
	}

	return fmt.Sprintf("Usuario %s creado correctamente", user)
}