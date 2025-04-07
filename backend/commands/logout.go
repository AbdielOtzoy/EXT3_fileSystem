package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
)

type LOGOUT struct {} // no requires parameters

func ParseLogout(tokens []string) (string, error) {
	fmt.Println("LOGOUT")

	if len(tokens) != 0 {
		return "", errors.New("no se esperaban argumentos")
	}

	// check if a session is active
	username, _, _, _ := stores.GetSession()
	if username == "" {
		return "No hay sesión activa", nil
	}

	// clear the session
	stores.SetSession("", "", -1, -1)

	return "Sesión cerrada", nil

}