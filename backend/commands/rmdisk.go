package commands

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"fmt"
)

type RMDISK struct {
	Path string
}

func ParseRmdisk(tokens []string) (string, error) {
	cmd := &RMDISK{} // create the rmdisk command

	args := strings.Join(tokens, " ") // join the tokens to get the arguments

	// regular expression to get the path
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)

	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		// dibide the argument in the key and value
		kv := strings.SplitN(match, "=", 3)
		if len(kv) != 2 {
			return "", errors.New("invalid argument")
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
			case "-path":
				cmd.Path = value
			default:
				return "", errors.New("invalid argument")
		}
	}

	if cmd.Path == "" {
		return "", errors.New("missing path")
	}

	err := commandRmdisk(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Disk %s removed successfully", cmd.Path), nil
}

func commandRmdisk(cmd *RMDISK) error {
	path := cmd.Path

	if !filepath.IsAbs(path) {
		return errors.New("invalid path")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("disk not found")
	}

	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	return nil
}