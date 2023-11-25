package main

import (
	"os"
)

func init() {

}

func fileexists(filename string) bool {
	var err error
	if _, err = os.Stat(filename); err != nil {
		return false
	} else {
		return true
	}
}
