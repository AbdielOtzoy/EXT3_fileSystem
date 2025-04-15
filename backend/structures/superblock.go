package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// SuperBlock represents the superblock of a filesystem.
// It contains metadata about the filesystem, such as the number of inodes, blocks, and their sizes.
type SuperBlock struct {
	S_filesystem_type   int32   // Type of the filesystem
	S_inodes_count      int32   // Total number of inodes
	S_blocks_count      int32   // Total number of blocks
	S_free_inodes_count int32   // Number of free inodes
	S_free_blocks_count int32   // Number of free blocks
	S_mtime             float32 // Last mount time (as a Unix timestamp)
	S_umtime            float32 // Last unmount time (as a Unix timestamp)
	S_mnt_count         int32   // Number of times the filesystem has been mounted
	S_magic             int32   // Magic number to identify the filesystem
	S_inode_size        int32   // Size of each inode
	S_block_size        int32   // Size of each block
	S_first_ino         int32   // First available inode
	S_first_blo         int32   // First available block
	S_bm_inode_start    int32   // Starting position of the inode bitmap
	S_bm_block_start    int32   // Starting position of the block bitmap
	S_inode_start       int32   // Starting position of the inode table
	S_block_start       int32   // Starting position of the block table
	// Total size: 68 bytes
}

// Serialize writes the SuperBlock structure to a binary file at the specified offset.
// This function is used to persist the SuperBlock data to disk.
func (sb *SuperBlock) Serialize(path string, offset int64) error {
	// Open the file for writing or create it if it doesn't exist
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move the file pointer to the specified offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serialize the SuperBlock structure directly into the file
	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize reads the SuperBlock structure from a binary file at the specified offset.
// This function is used to load the SuperBlock data from disk into memory.
func (sb *SuperBlock) Deserialize(path string, offset int64) error {
	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move the file pointer to the specified offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Get the size of the SuperBlock structure
	sbSize := binary.Size(sb)
	if sbSize <= 0 {
		return fmt.Errorf("invalid SuperBlock size: %d", sbSize)
	}

	// Read only the number of bytes corresponding to the size of the SuperBlock structure
	buffer := make([]byte, sbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserialize the read bytes into the SuperBlock structure
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// PrintSuperBlock displays the values of the SuperBlock structure in a human-readable format.
// This function is used for debugging and visualization purposes.
func (sb *SuperBlock) Print() {
	// Convert mount time to a human-readable format
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	// Convert unmount time to a human-readable format
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

	fmt.Printf("Filesystem Type: %d\n", sb.S_filesystem_type)
	fmt.Printf("Inodes Count: %d\n", sb.S_inodes_count)
	fmt.Printf("Blocks Count: %d\n", sb.S_blocks_count)
	fmt.Printf("Free Inodes Count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("Free Blocks Count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.S_mnt_count)
	fmt.Printf("Magic: %d\n", sb.S_magic)
	fmt.Printf("Inode Size: %d\n", sb.S_inode_size)
	fmt.Printf("Block Size: %d\n", sb.S_block_size)
	fmt.Printf("First Inode: %d\n", sb.S_first_ino)
	fmt.Printf("First Block: %d\n", sb.S_first_blo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", sb.S_bm_block_start)
	fmt.Printf("Inode Start: %d\n", sb.S_inode_start)
	fmt.Printf("Block Start: %d\n", sb.S_block_start)
}

// PrintInodes displays all inodes in the filesystem.
// This function is used for debugging and visualization purposes.
func (sb *SuperBlock) PrintInodes(path string) error {
	fmt.Println("\nInodes\n----------------")
	// Iterate over each inode
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserialize the inode
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Print the inode
		fmt.Printf("\nInode %d:\n", i)
		inode.Print()
	}

	return nil
}

// PrintBlocks displays all blocks in the filesystem.
// This function is used for debugging and visualization purposes.
func (sb *SuperBlock) PrintBlocks(path string) error {
	fmt.Println("\nBlocks\n----------------")
	// Iterate over each inode
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserialize the inode
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Iterate over each block pointer in the inode
		for _, blockIndex := range inode.I_block {
			// Skip if the block doesn't exist
			if blockIndex == -1 {
				break
			}
			// Handle folder blocks
			if inode.I_type[0] == '0' {
				block := &FolderBlock{}
				// Deserialize the folder block
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}
				// Print the folder block
				fmt.Printf("\nBlock %d:\n", blockIndex)
				block.Print()
				continue
			}
			// Handle file blocks
			if inode.I_type[0] == '1' {
				block := &FileBlock{}
				// Deserialize the file block
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return err
				}
				// Print the file block
				fmt.Printf("\nBlock %d:\n", blockIndex)
				block.Print()
				continue
			}
		}
	}

	return nil
}

// CreateFolder crea una carpeta en el sistema de archivos
func (sb *SuperBlock) CreateFolder(path string, parentsDir []string, destDir string, uid int32, gid int32) error {
	return sb.createFolderInInodeExt2(path, 0, parentsDir, destDir, uid, gid)

}

// CreateFile crea un archivo en el sistema de archivos
func (sb *SuperBlock) CreateFile(path string, parentsDir []string, destDir string, r bool, size int, content string, uid int32, gid int32) error {
	fmt.Println("Creando archivo:", path, "contenido:", content)
	return sb.createFileInodeExt2(path, 0, parentsDir, destDir, r, size, content, uid, gid)
}

func (sb *SuperBlock) ExistsFolcer(path string, parentsDir []string, destDir string) (bool, error) {
	exists, err := sb.folderExists(path, 0, parentsDir, destDir)
	if err != nil {
		return false, err
	}

	return exists, nil

}

// ReadFile lee el contenido de un archivo en el sistema de archivos
func (sb *SuperBlock) ReadFile(path string, parentsDir []string, destDir string) (string, error) {
	content, err := sb.readFileInInode(path, 0, parentsDir, destDir)
	if err != nil {
		return "", err
	}
	fmt.Println("Contenido del archivo:", content)
	return content, nil
}

func (sb *SuperBlock) GetInode(path string, parentsDir []string, destDir string) (int32, error) {
	inode, err := sb.getInodeFromPath(path, 0, parentsDir, destDir)
	if err != nil {
		return -1, err
	}

	return inode, nil
}

func (sb *SuperBlock) LoginUser(user string, password string, path string) (int32, int32, error) {
	uid, gid, err := sb.loginUserInInode(user, password, path)
	if err != nil {
		return -1, -1, err
	}

	return uid, gid, nil
}

func (sb *SuperBlock) CreateGroup(name string, path string) error {
	err := sb.createGroupInInode(name, path)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) RemoveGroup(name string, path string) error {
	err := sb.removeGroupInInode(name, path)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) CreateUser(user string, password string, group string, path string) error {
	err := sb.createUserInInode(user, password, group, path)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) RemoveUser(user string, path string) error {
	err := sb.removeUserInInode(user, path)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) ChangeGroup(user string, group string, path string) error {
	err := sb.changeGroupInInode(user, group, path)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) Delete(path string, parentsDir []string, destDir string) error {
	return sb.RemoveInode(path, 0, parentsDir, destDir)
}

func (sb *SuperBlock) EditFile(path string, parentsDir []string, destDir string, content string, uid int32, gid int32) error {
	return sb.EditFileInInode(path, 0, parentsDir, destDir, content, uid, gid)
}

func (sb *SuperBlock) RenameFile(path string, parentsDir []string, destDir string, name string, uid int32, gid int32) error {
	return sb.RenameFileInInode(path, 0, parentsDir, destDir, name, uid, gid)
}

func (sb *SuperBlock) CopyFile(path string, parentsDir []string, destDir string, destinoParentDirs []string, destinoDir string, uid int32, gid int32) error {
	return sb.CopyFileInInode(path, 0, parentsDir, destDir, destinoParentDirs, destinoDir, uid, gid)
}
