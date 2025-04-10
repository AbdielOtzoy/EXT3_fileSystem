package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// Inode represents a filesystem inode, which stores metadata about a file or directory.
// It includes information such as ownership, size, timestamps, block pointers, type, and permissions.
type Inode struct {
	I_uid   int32     // User ID of the owner
	I_gid   int32     // Group ID of the owner
	I_size  int32     // Size of the file in bytes
	I_atime float32   // Last access time (as a Unix timestamp)
	I_ctime float32   // Creation time (as a Unix timestamp)
	I_mtime float32   // Last modification time (as a Unix timestamp)
	I_block [15]int32 // Pointers to data blocks (15 blocks)
	I_type  [1]byte   // Type of the inode (e.g., file, directory)
	I_perm  [3]byte   // Permissions (e.g., read, write, execute)
	// Total size: 88 bytes
}

// Serialize writes the Inode structure to a binary file at the specified offset.
// This function is used to persist the Inode data to disk.
func (inode *Inode) Serialize(path string, offset int64) error {
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

	// Serialize the Inode structure directly into the file
	err = binary.Write(file, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize reads the Inode structure from a binary file at the specified offset.
// This function is used to load the Inode data from disk into memory.
func (inode *Inode) Deserialize(path string, offset int64) error {
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

	// Get the size of the Inode structure
	inodeSize := binary.Size(inode)
	if inodeSize <= 0 {
		return fmt.Errorf("invalid Inode size: %d", inodeSize)
	}

	// Read only the number of bytes corresponding to the size of the Inode structure
	buffer := make([]byte, inodeSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserialize the read bytes into the Inode structure
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil
}

// Print displays the attributes of the Inode in a human-readable format.
// This function is used for debugging and visualization purposes.
func (inode *Inode) Print() {
	// Convert Unix timestamps to human-readable format
	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	// Print all attributes of the Inode
	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
}
