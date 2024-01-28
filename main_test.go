package main

import (
	"fmt"
	"testing"
)

func TestOrder(t *testing.T) {
	var x uint16 = 0x0102
	bytes := [2]byte{byte(x), byte(x >> 8)}

	// Check if the bytes are stored in little-endian order or big-endian order
	if bytes[0] == 0x02 && bytes[1] == 0x01 {
		// return binary.LittleEndian
		fmt.Println("LIttle endian")
	} else if bytes[0] == 0x01 && bytes[1] == 0x02 {
		fmt.Println("Big endian")
	} else {
		// This should not happen, but return a default value if it does
		fmt.Println("Unknown byte order, defaulting to BigEndian")
	}

}
