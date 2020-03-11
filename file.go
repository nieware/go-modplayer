package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Instrument represents an instrument used in a MOD file, including the sample data
type Instrument = struct {
	Name     string
	Len      int
	Finetune int
	Volume   int
	RepStart int
	RepLen   int
	Sample   []byte
}

// Module stores a complete MOD file
type Module = struct {
	Name          string
	Signature     [4]byte
	InstrTableLen int
	PatternCnt    int
	Instruments   [31]Instrument
	PatternTable  []int
}

// ReadModFile reads the full MOD file given by fn and loads the data into the relevant objects
func ReadModFile(fn string) (mod Module, err error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Module Name
	mod.Name = strings.Trim(string(data[0:20]), " \t\n\v\f\r\x00")

	// Signature (also tells us the number of instruments)
	copy(mod.Signature[0:4], data[1080:1084])
	fmt.Printf("%#v %s\n", mod.Signature, string(mod.Signature[0:4]))
	mod.InstrTableLen = 31
	for _, c := range mod.Signature {
		// if the signature is not an ASCII string, we have an old module with 15 instruments
		if c < 32 {
			mod.InstrTableLen = 15
		}
	}

	// Pattern Table (have to read this first because this tells us the number of patterns)
	patternTableOffset := 20 + mod.InstrTableLen*30 + 2
	patternTableLen := int(data[20+mod.InstrTableLen*30+1])
	mod.PatternTable = make([]int, patternTableLen)
	for i := 0; i < patternTableLen; i++ {
		mod.PatternTable[i] = int(data[patternTableOffset+i])
		if mod.PatternTable[i] > mod.PatternCnt {
			mod.PatternCnt = mod.PatternTable[i] + 1
		}
	}
	fmt.Printf("%+v\n", mod)

	// Instruments
	sampleOffset := 20 + mod.InstrTableLen*30 + 2 + 128 + 4 + mod.PatternCnt*1024
	for i := 0; i < mod.InstrTableLen; i++ {
		mod.Instruments[i], err = ReadInstrument(data, 20+i*30, sampleOffset)
		sampleOffset += mod.Instruments[i].Len
	}

	// TODO Patterns

	return
}

/*
22        Sample's name, padded with null bytes. If a name begins with a
          '#', it is assumed not to be an instrument name, and is
          probably a message.
2         Sample length in words (1 word = 2 bytes). The first word of
          the sample is overwritten by the tracker, so a length of 1
          still means an empty sample. See below for sample format.
1         Lowest four bits represent a signed nibble (-8..7) which is
          the finetune value for the sample. Each finetune step changes
          the note 1/8th of a semitone. Implemented by switching to a
          different table of period-values for each finetune value.
1         Volume of sample. Legal values are 0..64. Volume is the linear
          difference between sound intensities. 64 is full volume, and
          the change in decibels can be calculated with 20*log10(Vol/64)
2         Start of sample repeat offset in words. Once the sample has
          been played all of the way through, it will loop if the repeat
          length is greater than one. It repeats by jumping to this
          position in the sample and playing for the repeat length, then
          jumping back to this position, and playing for the repeat
          length, etc.
2         Length of sample repeat in words. Only loop if greater than 1.
*/

// ReadInstrument reads an instrument from the MOD file data, including the sample data.
// The offset of the instrument data and the sampleOffset have to be passed as a parameter.
func ReadInstrument(data []byte, offset int, sampleOffset int) (ins Instrument, err error) {
	ins.Name = strings.Trim(string(data[offset:offset+22]), " \t\n\v\f\r\x00")

	//fmt.Printf("%x %x", data[offset+22], data[offset+23])
	ins.Len = int(data[offset+22])<<9 | int(data[offset+23])<<1

	//TODO ins.Finetune - signed nibble. sounds interesting...

	ins.Volume = int(data[offset+25])

	ins.RepStart = int(data[offset+26])<<9 | int(data[offset+27])<<1
	if data[offset+29] > 1 {
		ins.RepLen = int(data[offset+28])<<9 | int(data[offset+29])<<1
	}
	fmt.Printf("%s : Len %d, Vol %d, RepS %d, RepL %d\n", ins.Name, ins.Len, ins.Volume, ins.RepStart, ins.RepLen)

	ins.Sample = make([]byte, ins.Len)
	copy(ins.Sample, data[sampleOffset:sampleOffset+ins.Len])

	return
}
