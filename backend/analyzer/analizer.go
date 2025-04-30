package analyzer

import (
	commands "backend/commands"
	"errors"
	"fmt"
	"strings"
)

func Analyzer(input string) (string, error) {

	tokens := strings.Fields(input)

	if len(tokens) == 0 {
		return "", errors.New("no se proporcionó ningún comando")
	}

	command := strings.ToLower(tokens[0])

	switch command {
	case "mkdisk":
		return commands.ParseMkdisk(tokens[1:])
	case "rmdisk":
		return commands.ParseRmdisk(tokens[1:])
	case "fdisk":
		return commands.ParseFdisk(tokens[1:])
	case "mount":
		return commands.ParseMount(tokens[1:])
	case "unmount":
		return commands.ParseUnmounted(tokens[1:])
	case "mkfs":
		return commands.ParseMkfs(tokens[1:])
	case "rep":
		return commands.ParseRep(tokens[1:])
	case "mounted":
		return commands.ParseMounted(tokens[1:])
	case "mkdir":
		return commands.ParseMkdir(tokens[1:])
	case "mkfile":
		return commands.ParseMKfile(tokens[1:])
	case "cat":
		return commands.ParseCat(tokens[1:])
	case "login":
		return commands.ParseLogin(tokens[1:])
	case "logout":
		return commands.ParseLogout(tokens[1:])
	case "mkgrp":
		return commands.ParseMkgroup(tokens[1:])
	case "rmgrp":
		return commands.ParseRmgroup(tokens[1:])
	case "chgrp":
		return commands.ParseChgrp(tokens[1:])
	case "mkusr":
		return commands.ParseMkuser(tokens[1:])
	case "rmusr":
		return commands.ParseRmuser(tokens[1:])
	case "getfs":
		return commands.ParseGetfs(tokens[1:])
	case "remove":
		return commands.ParseRemove(tokens[1:])
	case "edit":
		return commands.ParseEdit(tokens[1:])
	case "rename":
		return commands.ParseRename(tokens[1:])
	case "copy":
		return commands.ParseCopy(tokens[1:])
	case "move":
		return commands.ParseMove(tokens[1:])
	case "chown":
		return commands.ParseCHOWN(tokens[1:])
	case "chmod":
		return commands.ParseCHMOD(tokens[1:])
	case "find":
		return commands.ParseFIND(tokens[1:])
	case "journaling":
		return commands.ParseJournal(tokens[1:])
	case "loss":
		return commands.ParseLoss(tokens[1:])
	case "recovery":
		return commands.ParseRecovery(tokens[1:])

	default:
		return "", fmt.Errorf("comando desconocido: %s", tokens[0])
	}
}
