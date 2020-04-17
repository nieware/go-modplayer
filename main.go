package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	infoOnly := flag.Bool("info", false, "only show module info")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("file name not specified")
		return
	}
	fn := flag.Args()[0]
	fmt.Println(fn)
	mod, err := ReadModFile(fn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mod.Info()
	if !*infoOnly {
		Play(mod)
	}

	/*for i := 0; i < mf.InstrTableLen; i++ {
		if mf.Instruments[i].Len > 0 {
			PlaySample(mf.Instruments[i])
		}

	} //*/

}
