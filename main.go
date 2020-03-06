package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("file name not specified")
		return
	}
	ReadModFile(os.Args[1])

}
