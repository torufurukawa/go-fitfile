package fitfile

// calcCRC calculates CRC from the current crc and val and returns new CRC
// value.
func calcCRC(crc uint16, val byte) uint16 {
	table := [16]uint16{
		0x0000, 0xcc01, 0xd801, 0x1400, 0xf001, 0x3c00, 0x2800, 0xe401,
		0xa001, 0x6c00, 0x7800, 0xb401, 0x5000, 0x9c01, 0x8801, 0x4400,
	}

	var tmp uint16
	// compute checksum of lower four bits
	tmp = table[crc&0xf]
	crc = uint16((crc >> 4) & 0xfff)
	crc = uint16(crc ^ tmp ^ table[val&0xf])

	// compute checksum of upper four bits
	tmp = table[crc&0xf]
	crc = uint16((crc >> 4) & 0xfff)
	crc = uint16(crc ^ tmp ^ table[(val>>4)&0xf])

	return crc
}
