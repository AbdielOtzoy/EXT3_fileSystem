package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type LOGIN struct {
	user string
	pass string
	id  string
}

func ParseLogin(tokens []string) (string, error) {
	cmd := &LOGIN{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user=[^\s]+|-pass=[^\s]+|-id=[^\s]+`)
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

	if cmd.user == "" {
		return "", errors.New("faltan parámetros requeridos: -user")
	}

	if cmd.pass == "" {
		return "", errors.New("faltan parámetros requeridos: -pass")
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	err := commandLogin(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Login exitoso para el usuario %s en la partición %s", cmd.user, cmd.id), nil
}

func commandLogin(cmd *LOGIN) error {
	userLogged, _, _, _ := stores.GetSession();
	if userLogged != "" {
		return errors.New("ya hay un usuario logueado")
	}

	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(cmd.id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	uid, gid, err := partitionSuperblock.LoginUser(cmd.user, cmd.pass, partitionPath)

	if err != nil {
		return fmt.Errorf("error al loguear el usuario: %w", err)
	}

	// si no hay error, loguear el usuario
	stores.SetSession(cmd.user, cmd.id, uid, gid)

	return nil
}