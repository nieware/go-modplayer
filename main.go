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
	mod, _ := ReadModFile(os.Args[1])

	Play(mod)
	/*for i := 0; i < mf.InstrTableLen; i++ {
		if mf.Instruments[i].Len > 0 {
			PlaySample(mf.Instruments[i])
		}

	} //*/

}
