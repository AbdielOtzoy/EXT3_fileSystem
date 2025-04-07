package structures

import (
	"encoding/binary"
	"os"
)

// CreateBitMaps creates the Inode and Block Bitmaps in the specified file.
func (sb *SuperBlock) CreateBitMaps(path string) error {
	// Open the file for writing or create it if it doesn't exist
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Inode Bitmap
	// Move the file pointer to the start of the Inode Bitmap
	_, err = file.Seek(int64(sb.S_bm_inode_start), 0)
	if err != nil {
		return err
	}

	// Create a buffer filled with '0's to represent free inodes
	buffer := make([]byte, sb.S_free_inodes_count)
	for i := range buffer {
		buffer[i] = '0'
	}

	// Write the Inode Bitmap buffer to the file
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	// Block Bitmap
	// Move the file pointer to the start of the Block Bitmap
	_, err = file.Seek(int64(sb.S_bm_block_start), 0)
	if err != nil {
		return err
	}

	// Create a buffer filled with 'O's to represent free blocks
	buffer = make([]byte, sb.S_free_blocks_count)
	for i := range buffer {
		buffer[i] = 'O'
	}

	// Write the Block Bitmap buffer to the file
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	return nil
}

// UpdateBitmapInode updates the Inode Bitmap to mark an inode as used.
func (sb *SuperBlock) UpdateBitmapInode(path string) error {
	// Open the file for reading and writing
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move the file pointer to the position of the specific inode in the bitmap
	_, err = file.Seek(int64(sb.S_bm_inode_start)+int64(sb.S_inodes_count), 0)
	if err != nil {
		return err
	}

	// Write '1' to mark the inode as used
	_, err = file.Write([]byte{'1'})
	if err != nil {
		return err
	}

	return nil
}

// UpdateBitmapBlock updates the Block Bitmap to mark a block as used.
func (sb *SuperBlock) UpdateBitmapBlock(path string) error {
	// Open the file for reading and writing
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move the file pointer to the position of the specific block in the bitmap
	_, err = file.Seek(int64(sb.S_bm_block_start)+int64(sb.S_blocks_count), 0)
	if err != nil {
		return err
	}

	// Write 'X' to mark the block as used
	_, err = file.Write([]byte{'X'})
	if err != nil {
		return err
	}

	return nil
}