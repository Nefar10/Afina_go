package main

import (
	"log"
	"os"
)

var gDir string

func init() {
	var err error
	if gDir, err = os.Getwd(); err != nil {
		log.Panic("Ошибка при определении текущей директории:", err)
	}
}

func fileexists(filename string) bool {
	var err error
	if _, err = os.Stat(gDir + filename); err != nil {
		return false
	} else {
		return true
	}
}
