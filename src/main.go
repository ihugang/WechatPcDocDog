package main

import (
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// 取不含后缀的文件名部门
func getFilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

// 还原图片文件
func decodeImage(fileName string) (bool, error) {
	fmt.Println("Decoding image :", fileName)

	file, err := os.Open(fileName)
	defer file.Close()

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	bytes := make([]byte, 2)
	signBytes := make([]byte, 2)

	_, err = file.Read(bytes)
	if err != nil {
		fmt.Println("读取文件头失败：", err)
		return false, err
	}

	ext := ".jpg"
	signBytes[0] = bytes[0] ^ 0xff
	signBytes[1] = bytes[1] ^ 0xd8
	if signBytes[0] != signBytes[1] {
		fmt.Println("文件头不是JPEG文件")

		ext = ".png"
		signBytes[0] = bytes[0] ^ 0x89
		signBytes[1] = bytes[1] ^ 0x50
		if signBytes[0] != signBytes[1] {
			fmt.Println("文件头不是PNG文件")

			ext = ".gif"
			signBytes[0] = bytes[0] ^ 0x47
			signBytes[1] = bytes[1] ^ 0x49

			if signBytes[0] != signBytes[1] {
				fmt.Println("文件头不是GIF文件")

				ext = ".tif"
				signBytes[0] = bytes[0] ^ 0x49
				signBytes[1] = bytes[1] ^ 0x49

				if signBytes[0] != signBytes[1] {
					fmt.Println("文件头不是JPEG文件")

					ext = ".bmp"
					signBytes[0] = bytes[0] ^ 0x42
					signBytes[1] = bytes[1] ^ 0x4d

					if signBytes[0] != signBytes[1] {
						return false, fmt.Errorf("文件头不是Image文件")
					}
				}

			}

		}
	}

	fmt.Println("Sign Byte:", hex.EncodeToString(signBytes)+"  "+ext)

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Println("文件偏移失败：", err)
		return false, err
	}

	folder := path.Dir(fileName)
	fileName = getFilenameWithoutExtension(path.Base(fileName)) + ext
	newFileName := path.Join(folder, fileName)

	fileW, errW := os.Create(newFileName)
	defer fileW.Close()

	if errW != nil {
		fmt.Println(errW)
		return false, errW
	}

	buffer := make([]byte, 1024)
	for {
		len, err := file.Read(buffer)
		if err == io.EOF || len < 0 {
			break
		}
		//fmt.Println("Read Bytes:", hex.EncodeToString(buffer))
		for i := 0; i < len; i++ {
			buffer[i] = buffer[i] ^ signBytes[1]
		}
		//fmt.Println("XOR Bytes:", hex.EncodeToString(buffer))
		fileW.Write(buffer[:len])
	}
	fileW.Close()
	file.Close()

	err = insertData(db, path.Base(fileName))
	if err != nil {
		fmt.Println("记录到数据库失败:", err)
	}

	fmt.Println("Image decoded successfully")
	return true, nil

}

var root = flag.String("folder", "", "File to decode")

func WalkDir(filePath string) {
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, v := range files {
			if v.IsDir() {
				fmt.Println("processing dir:", v.Name())
				WalkDir(path.Join(filePath, v.Name()))
			} else {
				decodeImage(path.Join(filePath, v.Name()))
			}
		}
	}
}

var db *sql.DB
var err error

func main() {
	fmt.Println("Decoding Wechat Backup images...")

	flag.Parse()

	db, err = initDb()

	if err != nil {
		fmt.Println(err)
		return
	}

	if *root == "" {
		folder, _ := os.Getwd()
		WalkDir(folder)
	} else {
		WalkDir(*root)
	}

}
