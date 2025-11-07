package main

import (
	"fmt"
	"os"
)

func rootDir() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("get working dir failed:", err)
		return "."
	}
	return wd
}
