package structures

import "fmt"

type Partition struct {
	Part_status      [1]byte // N Available, 0 created, 1 mounted
	Part_type        [1]byte
	Part_fit         [1]byte
	Part_start       int32
	Part_size        int32
	Part_name        [16]byte
	Part_correlative int32
	Part_id          [4]byte
}

func (p *Partition) CreatePartition(partStart, partSize int, partType, partFit, partName string) {

	p.Part_status[0] = '0' // 0 = Created, -1 = No created
	p.Part_start = int32(partStart)
	p.Part_size = int32(partSize)

	if len(partType) > 0 {
		p.Part_type[0] = partType[0]
	}

	if len(partFit) > 0 {
		p.Part_fit[0] = partFit[0]
	}

	copy(p.Part_name[:], partName) // copy the name to the partition
}

// Mount a partition by id
func (p *Partition) MountPartition(correlative int, id string) error {
	p.Part_status[0] = '1' // 0 = Created, 1 = Mounted
	p.Part_correlative = int32(correlative)
	copy(p.Part_id[:], id)

	return nil
}

// GetLastEBR returns the last EBR in the linked list
func (p *Partition) GetLastEBR(ebr *EBR, path string) (*EBR, int32, error) {
	// get the last EBR in the linked list
	// if the partition is not extended, return nil
	if p.Part_type[0] != 'E' {
		return nil, -1, fmt.Errorf("partition is not extended")
	}
	// recorrer el EBR hasta encontrar el ultimo

	next := ebr.Ebr_part_next
	for next != -1 {
		// deserialize the EBR
		err := ebr.DeserializeEBR(path, int64(next))
		if err != nil {
			return nil, -1, err
		}
		next = ebr.Ebr_part_next
	}

	fmt.Println("Last EBR")
	ebr.PrintEBR()

	return ebr, next, nil
}

func (p *Partition) DeletePartition() {
	p.Part_status[0] = 'N' // 0 = Created, -1 = No created
	p.Part_start = -1
	p.Part_size = -1
	p.Part_name = [16]byte{'N'}
	p.Part_correlative = -1
	p.Part_id = [4]byte{'N'}
	p.Part_type = [1]byte{'N'}
	p.Part_fit = [1]byte{'N'}
}

func (p *Partition) PrintPartition() {
	partStatus := rune(p.Part_status[0])
	partType := rune(p.Part_type[0])
	partFit := rune(p.Part_fit[0])

	partName := string(p.Part_name[:])
	partID := string(p.Part_id[:])

	fmt.Println("Partition")
	fmt.Printf("  Status: %c\n", partStatus)
	fmt.Printf("  Type: %c\n", partType)
	fmt.Printf("  Fit: %c\n", partFit)
	fmt.Printf("  Start: %d\n", p.Part_start)
	fmt.Printf("  Size: %d\n", p.Part_size)
	fmt.Printf("  Name: %s\n", partName)
	fmt.Printf("  Correlative: %d\n", p.Part_correlative)
	fmt.Printf("  ID: %s\n", partID)
}
