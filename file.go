package main

import (
	"encoding/binary"
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

type Effect int

const (
	Arpeggio           = iota // 0xy: x-first halfnote add, y-second
	SlideUp                   // 1xx: upspeed
	SlideDown                 // 2xx: downspeed
	Portamento                // 3xx: up/down speed
	Vibrato                   // 4xy: x-speed,   y-depth
	PortamentoVolSlide        // 5xy: x-upspeed, y-downspeed
	VibratoVolSlide           // 6xy: x-upspeed, y-downspeed
	Tremolo                   // 7xy: x-speed,   y-depth
	NotUsed8                  //
	SetSampleOffset           // 9xx: offset (23 -> 2300)
	VolSlide                  // Axy: x-upspeed, y-downspeed
	PositionJump              // Bxx: songposition
	SetVol                    // Cxx: volume, 00-40
	PatternBreak              // Dxx: break position in next patt
	Extended                  // Exy: see below...
	SetSpeed                  // Fxx: speed (00-1F) / tempo (20-FF)

	SetFilter          // E0x: 0-filter on, 1-filter off
	FineSlideUp        // E1x: value
	FineSlideDown      // E2x: value
	GlissandoControl   // E3x: 0-off, 1-on (use with tonep.)
	SetVibratoWaveform // E4x: 0-sine, 1-ramp down, 2-square
	SetLoop            // E5x: set loop point
	JumpToLoop         // E6x: jump to loop, play x times
	SetTremoloWaveform // E7x: 0-sine, 1-ramp down. 2-square
	NotUsedE8
	RetrigNote       // E9x: retrig from note + x vblanks
	FineVolSlideUp   // EAx: add x to volume
	FineVolSlideDown // EBx: subtract x from volume
	NoteCut          // ECx: cut from note + x vblanks
	NoteDelay        // EDx: delay note x vblanks
	PatternDelay     // EEx: delay pattern x notes
	InvertLoop       // EFx: speed
)

type Note = struct {
	Ins    *Instrument
	Period int
	Eff    Effect
	Pars   byte
}

type Pattern = [][]Note

// Module stores a complete MOD file
type Module = struct {
	Name          string
	Signature     [4]byte
	InstrTableLen int
	PatternCnt    int
	Instruments   [31]Instrument
	PatternTable  []int
	Patterns      [][][]Note
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

	// Patterns
	mod.Patterns = make([][][]Note, mod.PatternCnt)
	patternsOffset := 20 + mod.InstrTableLen*30 + 2 + 128 + 4
	for i := range mod.Patterns {
		mod.Patterns[i] = make([][]Note, 64)
		fmt.Printf("\n\nPattern %d:\n", i)
		for j := range mod.Patterns[i] {
			mod.Patterns[i][j] = make([]Note, 4)
			for k := range mod.Patterns[i][j] {
				noteOffset := patternsOffset + ((i*64+j)*4+k)*4
				mod.Patterns[i][j][k] = ReadNote(data, &mod, noteOffset)
			}
			fmt.Printf("%+v\n", mod.Patterns[i][j])
		}
	}

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

func ReadNote(data []byte, mod *Module, offset int) (n Note) {
	insNum := data[offset]&0xF0 | (data[offset+2]&0xF0)>>4
	n.Ins = &mod.Instruments[insNum]
	bsl := []byte{data[offset] & 0x0F, data[offset+1]}
	n.Period = (int)(binary.BigEndian.Uint16(bsl))
	effNum := data[offset+2]
	if effNum != 0xE {
		n.Eff = Effect(effNum)
		fmt.Println(n.Eff)
		n.Pars = data[offset+3]
	} else {
		effSubNum := (data[offset+3] & 0xF0) >> 4
		n.Eff = Effect(16 + effSubNum)
		fmt.Println(n.Eff)
		n.Pars = data[offset+3] & 0x0F
	}
	return
}
