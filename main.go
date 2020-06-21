package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

func decodeNote(noteToDecode string) {
	noteData, err := hex.DecodeString(noteToDecode)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(noteData) < 4 {
		fmt.Println("not enough data to decode")
		os.Exit(1)
	}
	note := ReadNote(noteData, &Module{})
	note.Details()
}

func main() {
	infoOnly := flag.Bool("info", false, "only show module info")
	playSamples := flag.Bool("samples", false, "play only the samples rather than the complete song")
	noteToDecode := flag.String("note", "", "specify a note to decode")
	start := flag.Int("s", 0, "start from the specified order (pattern list index)")
	flag.Usage = Usage
	flag.Parse()

	if *noteToDecode != "" {
		decodeNote(*noteToDecode)
		return
	}

	if len(flag.Args()) < 1 {
		fmt.Println("file name not specified")
		os.Exit(1)
	}
	fn := flag.Args()[0]
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
		Play(mod, *start)
	}

}

// Usage is our custom usage function
var Usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] [filename]\nFlags:\n", os.Args[0])
	flag.PrintDefaults()
}
