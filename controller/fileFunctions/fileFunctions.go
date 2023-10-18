package fileFunctions

import (
	"io"
	"log"
	"os"
)

func GetFileContent(fileHandle *os.File) string {
	b, err := io.ReadAll(fileHandle)
	if err != nil {
		log.Println(err)
	}

	return string(b)
}

func GetFilehandle(filename string) *os.File {
	fileHandle, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println(err)
	}

	return fileHandle
}

func WriteFile(newContent string, fileHandle *os.File) {
	oldContent := GetFileContent(fileHandle)

	if newContent != oldContent {
		log.Println("File changed")
		if err := fileHandle.Truncate(0); err != nil {
			log.Println(err.Error())
		}
		if _, err := fileHandle.Seek(0, 0); err != nil {
			log.Println(err.Error())
		}
		if _, err := fileHandle.Write([]byte(newContent)); err != nil {
			log.Println(err.Error())
		}
	} else {
		log.Println("File not changed")
	}
}
