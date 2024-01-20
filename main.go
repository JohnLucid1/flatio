package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"

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
)

// TODO:
// TODO: implement incription based on a key :)
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

		for _, value := range files {
			file_buff := make([]byte, FLSIZE)
			file_path_buff := Package_filepath(value)
			offset := 0

			file, err := os.Open(value)
			if err != nil {
				log.Fatalln(err)
				return
			}

			for {
				bytesRead, err := file.Read(file_buff)
				if err == io.EOF {
					fmt.Println("breaking here")
					break
				} else if bytesRead == 0 {
					fmt.Println("Done sending", file.Name())
					break
				}

				testbuff := make([]byte, 2308)
				for i, v := range file_path_buff {
					testbuff[i] = v
				}

				idx := 0
				for i := 300; i < 2300; i++ {
					testbuff[i] = file_buff[idx]
					idx++
				}

				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(offset))
				offset += bytesRead
				idx = 0
				for i := 2300; i < 2308; i++ {
					testbuff[i] = b[idx]
					idx++
				}
				conn.Write(testbuff)

			}
		}
		conn.Close()

	} else { // Receiving
		listener, err := net.Listen("tcp", ip)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalln(err)
				return
			}
			defer conn.Close()

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
		}
	}
}

func handle_data(buff []byte) error {
	path_buff := buff[0:300]
	content_buff := buff[300:2300]
	offset_buff := buff[2300:2308]

	path := get_data(path_buff, 300)
	content := get_data(content_buff, 2000)
	offset := binary.LittleEndian.Uint64(offset_buff)

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

// func test_receive(cnt []byte) {
// 	path_buff := cnt[0:300]
// 	content_buff := cnt[300:2300]
// 	offset_buff := cnt[2300:2308]
// 	path := get_data(path_buff, 300)
// 	content :=  get_data(content_buff,2000)
// 	offset := binary.LittleEndian.Uint64(offset_buff)
// 	fmt.Println(path,"OFFSET: ", offset)
// 	fmt.Println(content)
// }

func get_data(data []byte, size int) string {
	content := make([]byte, size)

	idx := 0
	for i := 0; i < size; i++ {
		if data[i] == '\000' {
			break
		}
		content[i] = data[i]
		idx ++
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

func Package_filepath(path string) [FPSIZE]byte {
	var fl_path_arr [FPSIZE]byte
	for key, value := range path {
		fl_path_arr[key] = byte(value)
	}

	return fl_path_arr
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
