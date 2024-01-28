package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"runtime"

	// "io"
	"log"
	"math/rand"

	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	FLSIZE = 2000
	FPSIZE = 300
	PSSIZE = 100
	WHOLE  = 2400
)

// TODOOO: Rewrite the system so the first 8 bytes are the length of the filename, second 8 bytes are the length of content, and third 8 bytes is the offset
// TODO: Add a readme
// TODO: needs refactoring
func main() {
	root := ""
	sending := false
	ip := ""

	flag.StringVar(&root, "p", "./", "Set the root or filepath")                           // check if dir or not
	flag.StringVar(&ip, "i", "", "Set your ip and port adress in <127.0.0.1:8080> format") // check if dir or not
	flag.BoolVar(&sending, "s", false, "set this if you are sending data")
	flag.Parse()

	if sending {
		conn, err := net.Dial("tcp", ip)
		if err != nil {
			log.Fatalln(err)
			return
		}

		files, err := FilePathWalkDir(root)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		var endianness_indicator uint32
		err = binary.Read(conn, binary.BigEndian, &endianness_indicator)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}
		var byte_order binary.ByteOrder
		if endianness_indicator == 0x12345678 {
			byte_order = binary.BigEndian
		} else if endianness_indicator == 0x78563412 {
			byte_order = binary.LittleEndian
		}

		for _, value := range files {
			filepath_buff := make([]byte, 8)
			byte_order.PutUint64(filepath_buff, uint64(len(value)))

			file, err := os.Open(value)
			if err != nil {
				log.Fatalln(err)
				return
			}

			content_buff := make([]byte, FLSIZE)
			file_path_buff := []byte(value)
			offset := 0

			for {
				bytesRead, err := file.Read(content_buff)
				if err == io.EOF {
					fmt.Println("breaking here")
					break
				} else if bytesRead == 0 {
					fmt.Println("Done sending", file.Name())
					break
				}

				final := make([]byte, WHOLE)
				copy(final[0:len(filepath_buff)], filepath_buff)

				content_size := make([]byte, 8)
				byte_order.PutUint64(content_size, uint64(bytesRead))
				// idx := 0
				// for i := 8; i < len(content_size); i ++ {
				// 	final[i] = content_size[idx] // size of content
				// 	idx ++
				// }
				copy(final[8:8+len(content_size)], content_size)

				// idx = 0
				offset_buff := make([]byte, 8)
				byte_order.PutUint64(offset_buff, uint64(offset))
				// for i := 16; i < len(offset_buff); i ++ {
				// 	final[i] = offset_buff[idx] // size of offset
				// 	idx ++
				// }
				copy(final[16:16+len(offset_buff)], offset_buff)

				// idx = 0
				// for i := 24; i <  len(file_path_buff); i ++ {
				// 	final[i] = file_path_buff[idx]
				// 	idx++
				// }
				copy(final[24:], file_path_buff)

				// idx = 0
				// for i := len(file_path_buff); i < len(content_buff); i ++ {
				// 	final[i] = content_buff[idx]
				// 	idx++
				// }
				copy(final[len(file_path_buff):len(content_buff)], content_buff)
				test_senddata(final, byte_order)
				offset += bytesRead
				//TODO: conn.Write(final)
			}
		}
		// conn.Close()

	} else { // Receiving
		listener, err := net.Listen("tcp", ip)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer listener.Close()

		conn, err := listener.Accept()
		byte_order := get_os_indianness()
		switch byte_order {
		case binary.BigEndian:
			endianness_indicator := make([]byte, 4)
			byte_order.PutUint32(endianness_indicator, 0x12345678)
			_, err := conn.Write(endianness_indicator)
			if err != nil {
				log.Fatalln(err)
				return
			}

		case binary.LittleEndian:
			endianness_indicator := make([]byte, 4)
			byte_order.PutUint32(endianness_indicator, 0x12345678)
			_, err := conn.Write(endianness_indicator)
			if err != nil {
				log.Fatalln(err)
				return
			}
		default:
			endianness_indicator := make([]byte, 4)
			byte_order.PutUint32(endianness_indicator, 0x12345678)
			_, err := conn.Write(endianness_indicator)
			if err != nil {
				log.Fatalln(err)
				return
			}
		}

		for {
			if err != nil {
				log.Fatalln(err)
				return
			}

			buffer := make([]byte, 2308)
			_, err = conn.Read(buffer)
			if err != nil {
				log.Fatalln(err)
				return
			}

			err = handle_data(buffer)
			if err != nil {
				log.Fatalln(err)
				return
			}
			defer conn.Close()
		}
	}
}

func test_senddata(buff []byte, byteorder binary.ByteOrder) error {
	filename_length := byteorder.Uint64(buff[0:8])
	content_length := byteorder.Uint64(buff[8:16])
	offset := byteorder.Uint64(buff[16:24])
	fmt.Println("DEBUG: OFFSET: ", offset)
	filename := string(buff[24:filename_length])
	fmt.Println("DEBUG: FILENAME", filename)
	content := string(buff[filename_length:content_length])
	fmt.Println("DEBUG: CONTENT", content)
	return nil
}

func handle_data(buff []byte) error {
	path_buff := buff[0:300]
	content_buff := buff[300:2300]
	offset_buff := buff[2300:2308]

	path := get_data(path_buff, 300)
	content := get_data(content_buff, 2000)
	offset := binary.LittleEndian.Uint64(offset_buff) // todo: REPLACE this

	directory, fppath := filepath.Split(path)
	if len(directory) == 0 {
		file, err := os.OpenFile(fppath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = file.WriteAt([]byte(content), int64(offset))
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	} else {
		// TODO: parse directory and replace
		if runtime.GOOS == "darwin" {
			directory = strings.ReplaceAll("\\", directory, "/")
			path = strings.ReplaceAll("\\", path, "/")
		}

		err := os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			log.Println(err)
			return err
		}

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = file.WriteAt([]byte(content), int64(offset))
		if err != nil {
			log.Println(err)
			return err
		}
		file.Close()

		return nil
	}
}

func get_os_indianness() binary.ByteOrder {
	var x uint16 = 0x0102
	bytes := [2]byte{byte(x), byte(x >> 8)}

	// Check if the bytes are stored in little-endian order or big-endian order
	if bytes[0] == 0x02 && bytes[1] == 0x01 {
		return binary.LittleEndian
	} else if bytes[0] == 0x01 && bytes[1] == 0x02 {
		return binary.BigEndian
	} else {
		// This should not happen, but return a default value if it does
		fmt.Println("Unknown byte order, defaulting to BigEndian")
	}
	return binary.BigEndian
}

func get_data(data []byte, size int) string {
	content := make([]byte, size)

	idx := 0
	for i := 0; i < size; i++ {
		if data[i] == '\000' {
			break
		}
		content[i] = data[i]
		idx++
	}
	return string(content[:idx])
}

func gen_password(length uint8) string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, length)
	rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func FilePathWalkDir(root string) ([]string, error) {
	var skip = []string{
		".git", "yarn", ".exe",
	}

	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && !CheckContains(path, skip) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func CheckContains(input string, subject []string) bool {
	for _, value := range subject {
		if strings.Contains(input, value) {
			return true
		}
	}
	return false
}
