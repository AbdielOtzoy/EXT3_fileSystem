package structures

import (
	utils "backend/utils"
	"fmt"
	"strings"
)

func (sb *SuperBlock) loginUserInInode(user string, password string, path string) (int32, int32, error) {
	// obtener el contenido del bloque
	content := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(content))

	var contentEndPos int
	// encontrar donde termina realmente el contenido (primer byte nulo o fin del array)
	for i, b := range content {
		if b == 0 {
			contentEndPos = i
			break
		}
	}

	// si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(content)
	}

	// obtener solo el contenido real (sin bytes nulos)
	content = string(content[:contentEndPos])
	fmt.Println("Tamaño del contenido real: ", len(content))
	// eliminar los caracteres nulos

	// separar el contenido por lineas
	lines := strings.Split(content, "\n")
	// eliminar la ultima linea si es vacia
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	fmt.Println("Contenido login: ", content)

	uid := int32(0)
	gid := int32(0)
	grupoUsuario := ""
	// iterar sobre cada linea
	for _, line := range lines {
		// separar la linea por comas
		parts := strings.Split(line, ",")

		/*
			formato del archivo:
			GID, Tipo, Grupo \n
			UID, Tipo, Grupo, Usuario, Contraseña \n
		*/

		fmt.Println("Partes: ", parts)

		// verificar si es un usuario
		if parts[1] == "U" {
			fmt.Println("Usuario: ", parts[3])
			fmt.Println("Contraseña: ", parts[4])
			if parts[3] == user && parts[4] == password {
				fmt.Println("Usuario: ", parts[3])
				fmt.Println("Contraseña: ", parts[4])
				uid = utils.StringToInt32(parts[0])
				grupoUsuario = parts[2]
			} else if parts[3] == user && parts[4] != password {
				fmt.Println("Contraseña incorrecta")
				return 0, 0, fmt.Errorf("contraseña incorrecta")
			}
		}
	}

	// iterar para buscar el grupo
	for _, line := range lines {
		// separar la linea por comas
		parts := strings.Split(line, ",")
		// verificar si es un grupo
		if parts[1] == "G" {
			fmt.Println("Grupo: ", parts[2])
			if parts[2] == grupoUsuario {
				fmt.Println("Grupo: ", parts[2])
				gid = utils.StringToInt32(parts[0])
				fmt.Println("GID: ", gid)
				return uid, gid, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("usuario o contraseña incorrectos")

}

func (sb *SuperBlock) createGroupInInode(name string, path string) error {
	useBlock := &FileBlock{}

	// Deserializar el bloque
	err := useBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	// Separar el contenido por líneas
	lines := strings.Split(content, "\n")

	// Eliminar el último elemento si está vacío
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	countGrupos := 0
	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "G" {
			fmt.Println("Grupo: ", parts[2])
			countGrupos++
			if parts[2] == name {
				fmt.Println("Grupo ya existe")
				return fmt.Errorf("grupo ya existe")
			}
		}
	}

	// Si el grupo no existe, agregarlo
	newGrupo := fmt.Sprintf("%d,G,%s\n", countGrupos+1, name)
	newContent := content + newGrupo
	fmt.Println("Nuevo grupo: ", newGrupo)
	fmt.Println("Contenido nuevo: ", newContent)
	fmt.Println("Tamaño del nuevo contenido: ", len(newContent))

	err = sb.setUsersContent(path, newContent)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) removeGroupInInode(name string, path string) error {
	userBlock := &FileBlock{}

	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	// Separar el contenido por líneas
	lines := strings.Split(content, "\n")

	// Eliminar el último elemento si está vacío
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Variable para indicar si se encontró y modificó el grupo
	grupoEncontrado := false
	// Variable para almacenar el nuevo contenido
	var newContent strings.Builder

	for i, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 2 && parts[1] == "G" && parts[2] == name {
			fmt.Println("Grupo encontrado: ", parts[2])

			if parts[0] == "0" {
				return fmt.Errorf("El grupo ya no existe porque ya fue eliminado")
			}

			// Modificar el GID a 0 para marcar como eliminado
			parts[0] = "0"
			grupoEncontrado = true

			// Reconstruir la línea con el GID modificado
			modifiedLine := strings.Join(parts, ",")

			// Añadir la línea modificada
			newContent.WriteString(modifiedLine)
		} else {
			// Añadir la línea sin modificar
			newContent.WriteString(line)
		}

		// Añadir salto de línea si no es la última línea
		if i < len(lines)-1 {
			newContent.WriteString("\n")
		} else {
			// Añadir salto de línea al final
			newContent.WriteString("\n")
		}
	}

	if !grupoEncontrado {
		fmt.Println("Grupo no encontrado")
		return fmt.Errorf("grupo no encontrado")
	}

	// Obtener el contenido modificado
	finalContent := newContent.String()
	fmt.Println("Contenido modificado: ", finalContent)
	fmt.Println("Tamaño del contenido modificado: ", len(finalContent))

	err = sb.setUsersContent(path, finalContent)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) createUserInInode(user string, pass string, grp string, path string) error {
	userBlock := &FileBlock{}

	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	lines := strings.Split(content, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	userCount := 0           // Contador de usuarios
	userGroupExists := false // Variable para verificar si el grupo existe
	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "U" {
			fmt.Println("Usuario: ", parts[3])
			userCount++
			if parts[3] == user {
				fmt.Println("Usuario ya existe")
				return fmt.Errorf("usuario ya existe")
			}

		}
		// verificar si es un grupo
		if len(parts) > 1 && parts[1] == "G" {
			fmt.Println("Grupo: ", parts[2])
			if parts[2] == grp {
				// verificar que el grupo no haya sido eliminado
				if parts[0] == "0" {
					fmt.Println("Grupo eliminado")
					return fmt.Errorf("el grupo fue eliminado")
				}
				fmt.Println("Grupo existe")
				userGroupExists = true
			}
		}
	}

	if !userGroupExists {
		fmt.Println("Grupo no existe")
		return fmt.Errorf("grupo no existe")
	}

	// agregar el nuevo usuario: userCount, U, grp, user, pass
	newUser := fmt.Sprintf("%d,U,%s,%s,%s\n", userCount+1, grp, user, pass)
	newContent := content + newUser
	fmt.Println("Nuevo usuario: ", newUser)
	fmt.Println("Contenido nuevo: ", newContent)
	fmt.Println("Tamaño del nuevo contenido: ", len(newContent))

	err = sb.setUsersContent(path, newContent)
	if err != nil {
		return err
	}

	return nil

}

func (sb *SuperBlock) removeUserInInode(user string, path string) error {
	userBlock := &FileBlock{}

	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	// Encontrar dónde termina realmente el contenido (primer byte nulo o fin del array)
	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	// Si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	// Obtener solo el contenido real (sin bytes nulos)
	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	lines := strings.Split(content, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	userExists := false // Variable para verificar si el usuario existe
	// Variable para almacenar el nuevo contenido
	var newContent strings.Builder
	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "U" && parts[3] == user {
			fmt.Println("Usuario encontrado: ", parts[3])

			if parts[0] == "0" {
				return fmt.Errorf("El usuario ya no existe porque ya fue eliminado")
			}

			// Modificar el UID a 0 para marcar como eliminado
			parts[0] = "0"
			userExists = true

			// Reconstruir la línea con el UID modificado
			modifiedLine := strings.Join(parts, ",")

			// Añadir la línea modificada
			newContent.WriteString(modifiedLine)
		} else {
			// Añadir la línea sin modificar
			newContent.WriteString(line)
		}

		// Añadir salto de línea si no es la última línea
		if line != lines[len(lines)-1] {
			newContent.WriteString("\n")
		} else {
			// Añadir salto de línea al final
			newContent.WriteString("\n")
		}

	}

	if !userExists {
		fmt.Println("Usuario no encontrado")
		return fmt.Errorf("usuario no encontrado")
	}

	// Obtener el contenido modificado
	finalContent := newContent.String()
	fmt.Println("Contenido modificado: ", finalContent)
	fmt.Println("Tamaño del contenido modificado: ", len(finalContent))
	err = sb.setUsersContent(path, finalContent)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) changeGroupInInode(user string, group string, path string) error {
	userBlock := &FileBlock{}
	// Deserializar el bloque
	err := userBlock.Deserialize(path, int64(sb.S_block_start+(1*sb.S_block_size)))
	if err != nil {
		fmt.Println("Error al deserializar el bloque: ", err)
		return err
	}

	useContent := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(useContent))

	var contentEndPos int
	for i, b := range useContent {
		if b == 0 {
			contentEndPos = i
			break
		}
	}
	if contentEndPos == 0 {
		contentEndPos = len(useContent)
	}

	content := string(useContent[:contentEndPos])

	fmt.Println("Tamaño del contenido real: ", len(content))
	fmt.Println("Contenido Grup: ", content)

	lines := strings.Split(content, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	usuarioEncontrado := false
	grupoEncontrado := false
	var newContent strings.Builder

	for _, line := range lines {
		parts := strings.Split(line, ",")

		fmt.Println("Partes: ", parts)

		if len(parts) > 1 && parts[1] == "U" && parts[3] == user {
			fmt.Println("Usuario encontrado: ", parts[3])

			if parts[0] == "0" {
				fmt.Println("El usuario ya no existe porque ya fue eliminado")
				return fmt.Errorf("el usuario ya no existe porque ya fue eliminado")
			}

			// Modificar el grupo
			parts[2] = group
			usuarioEncontrado = true

			modifiedLine := strings.Join(parts, ",")

			newContent.WriteString(modifiedLine)
		} else {
			newContent.WriteString(line)
		}

		if len(parts) > 1 && parts[1] == "G" {
			// Verificar si el grupo existe o si no fue eliminado
			if parts[2] == group {
				if parts[0] == "0" {
					fmt.Println("El grupo fue eliminado")
					return fmt.Errorf("el grupo fue eliminado")
				}
				fmt.Println("Grupo encontrado: ", parts[2])
				grupoEncontrado = true
			}
		}

		if line != lines[len(lines)-1] {
			newContent.WriteString("\n")
		} else {
			newContent.WriteString("\n")
		}
	}

	if !usuarioEncontrado {
		fmt.Println("Usuario no encontrado")
		return fmt.Errorf("usuario no encontrado")
	}

	if !grupoEncontrado {
		fmt.Println("Grupo no encontrado")
		return fmt.Errorf("grupo no encontrado")
	}

	// Obtener el contenido modificado
	finalContent := newContent.String()
	fmt.Println("Contenido modificado: ", finalContent)
	fmt.Println("Tamaño del contenido modificado: ", len(finalContent))

	err = sb.setUsersContent(path, finalContent)
	if err != nil {
		return err
	}

	return nil
}

// funcion para obtener el contenido de users.txt
func (sb *SuperBlock) getUsersContent(path string) string {
	// obtener el inodo para users.txt, este siempre será el inodo 1
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(1*sb.S_inode_size)))
	if err != nil {
		return ""
	}
	fmt.Println("Inodo de users.txt")
	inode.Print()

	// iterar sobre cada bloque del inodo y obtener el contenido
	contentBlocks := "" // para almacenar el contenido de los bloques
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		// crear un nuevo bloque de archivo
		block := &FileBlock{}
		// deserializar el bloque
		err = block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return ""
		}

		fmt.Println("Bloque de archivo: ", blockIndex)
		block.Print()
		// convertir el bloque a string
		blockContent := string(block.B_content[:])
		fmt.Println("Contenido del bloque: ", blockContent)
		// agregar el contenido del bloque al contenido total
		contentBlocks += blockContent
		fmt.Println("Contenido total: ", contentBlocks)

	}

	// returnar el contenido
	return contentBlocks
}

// funcion para setear el contenido de users.txt
func (sb *SuperBlock) setUsersContent(path string, content string) error {
	fmt.Println("Proximo bloque disponible: ", sb.S_blocks_count)
	// obtener el inodo para users.txt, este siempre será el inodo 1
	// dividir el contenido en bloques de 64 bytes - contentDivide
	// iterar sobre cada bloque del inodo y settear el contenido
	// si len(contentDivide) > bloques utilizados, entonces se necesita crear nuevo bloque para alamacenar el contenido
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(1*sb.S_inode_size)))
	if err != nil {
		return err
	}

	fmt.Println("Inodo de users.txt")
	inode.Print()
	// obtener el contenido de los bloques del inodo

	// crear un arreglo de bloques que almacene el contenido de 64 bytes
	contentBlocks := make([]string, 0)
	for i := 0; i < len(content); i += 64 {
		if i+64 > len(content) {
			contentBlocks = append(contentBlocks, content[i:])
		} else {
			contentBlocks = append(contentBlocks, content[i:i+64])
		}
	}

	fmt.Println("Contenido de los bloques: ", contentBlocks)

	// iterar sobre cada bloque del inodo y settear el contenido
	for i, blockIndex := range inode.I_block {
		fmt.Println("Bloque: ", blockIndex)
		// a cada bloque del inodo se le asigna el contenido de 64 bytes
		if blockIndex == -1 {
			if len(contentBlocks) > i {
				// todavia hay contenido por asignar pero el bloque no existe
				// actualizar el inodo para que apunte al nuevo bloque
				inode.I_block[i] = sb.S_blocks_count
				// serializar el inodo
				err = inode.Serialize(path, int64(sb.S_inode_start+(1*sb.S_inode_size)))
				if err != nil {
					return err
				}

				newBlock := &FileBlock{
					B_content: [64]byte{},
				}
				copy(newBlock.B_content[:], contentBlocks[i])
				fmt.Println("Bloque de archivo: ", blockIndex)
				// serializar el bloque
				err = newBlock.Serialize(path, int64(sb.S_block_start+(sb.S_blocks_count*sb.S_block_size)))
				if err != nil {
					return err
				}

				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				// Actualizamos el superbloque
				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size
				break

			}
			break
		}

		// obtener el bloque
		block := &FileBlock{}
		// deserializar el bloque
		err = block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return err
		}

		// limpiar el bloque
		for j := range block.B_content {
			block.B_content[j] = 0
		}

		contentBytes := []byte(contentBlocks[i])
		// Copiar byte por byte el nuevo contenido
		for i := 0; i < len(contentBytes); i++ {
			block.B_content[i] = contentBytes[i]
		}
		fmt.Println("Bloque después de actualizar: ", string(block.B_content[:]))
		// serializar el bloque
		err = block.Serialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return err
		}
	}

	return nil
}

// función para obtener el uid y gid de un usuario por el nombre
func (sb *SuperBlock) GetUidGidByNameInInode(name string, path string) (int32, int32, error) {
	// obtener el contenido del bloque
	content := sb.getUsersContent(path)

	fmt.Println("Tamaño del bloque: ", len(content))

	var contentEndPos int
	// encontrar donde termina realmente el contenido (primer byte nulo o fin del array)
	for i, b := range content {
		if b == 0 {
			contentEndPos = i
			break
		}
	}

	// si no hay bytes nulos, usar todo el array
	if contentEndPos == 0 {
		contentEndPos = len(content)
	}

	// obtener solo el contenido real (sin bytes nulos)
	content = string(content[:contentEndPos])
	fmt.Println("Tamaño del contenido real: ", len(content))
	// eliminar los caracteres nulos

	// separar el contenido por lineas
	lines := strings.Split(content, "\n")
	// eliminar la ultima linea si es vacia
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	fmt.Println("Contenido login: ", content)

	uid := int32(0)
	gid := int32(0)
	grupoUsuario := ""
	// iterar sobre cada linea
	for _, line := range lines {
		// separar la linea por comas
		parts := strings.Split(line, ",")

		/*
			formato del archivo:
			GID, Tipo, Grupo \n
			UID, Tipo, Grupo, Usuario, Contraseña \n
		*/

		fmt.Println("Partes: ", parts)

		// verificar si es un usuario
		if parts[1] == "U" {
			fmt.Println("Usuario: ", parts[3])
			fmt.Println("Contraseña: ", parts[4])
			if parts[3] == name {
				fmt.Println("Usuario: ", parts[3])
				fmt.Println("Contraseña: ", parts[4])
				uid = utils.StringToInt32(parts[0])
				grupoUsuario = parts[2]
			}
		}
	}

	// iterar para buscar el grupo
	for _, line := range lines {
		// separar la linea por comas
		parts := strings.Split(line, ",")
		// verificar si es un grupo
		if parts[1] == "G" {
			fmt.Println("Grupo: ", parts[2])
			if parts[2] == grupoUsuario {
				fmt.Println("Grupo: ", parts[2])
				gid = utils.StringToInt32(parts[0])
				fmt.Println("GID: ", gid)
				return uid, gid, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("usuario o contraseña incorrectos")
}
