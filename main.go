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
	chans := flag.String("S", "", "play only specified channels")
	flag.Usage = Usage
	flag.Parse()

	if *noteToDecode != "" {
		decodeNote(*noteToDecode)
		return
	}

	if len(flag.Args()) < 1 {
		exitOnError(fmt.Errorf("file name not specified"), 1)
	}
	fn := flag.Args()[0]
	mod, err := ReadModFile(fn)
	exitOnError(err, 1)

	mod.Info()
	if *infoOnly {
		return
	}
	if *playSamples {
		for i := 0; i < mod.InstrTableLen; i++ {
			if mod.Instruments[i].Len > 0 {
				fmt.Println("Playing sample", i)
				err := PlaySample(mod.Instruments[i])
				exitOnError(err, 2)
			}
		} //*/
	} else {
		err := Play(mod, *start, *chans)
		exitOnError(err, 2)
	}

}

// exitOnError terminates the application printing err and returning code if err is set
func exitOnError(err error, code int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(code)
	}
}

// Usage is our custom usage function
var Usage = func() {
	_, err := fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] [filename]\nFlags:\n", os.Args[0])
	if err != nil {
		return
	}
	flag.PrintDefaults()
}
