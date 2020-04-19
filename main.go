package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	infoOnly := flag.Bool("info", false, "only show module info")
	playSamples := flag.Bool("samples", false, "play only the samples rather than the complete song")
	flag.Usage = Usage
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
	if *infoOnly {
		return
	}
	if *playSamples {
		for i := 0; i < mod.InstrTableLen; i++ {
			if mod.Instruments[i].Len > 0 {
				fmt.Println("Playing sample", i)
				PlaySample(mod.Instruments[i])
			}
		} //*/
	} else {
		Play(mod)
	}

}

// Usage is our custom usage function
var Usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] [filename]\nFlags:\n", os.Args[0])
	flag.PrintDefaults()
}
