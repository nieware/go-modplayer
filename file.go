package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strings"
)

// EffectType represents a module effect
type EffectType int

const (
	// Arpeggio 0xy: x-first halfnote add, y-second - period cycles between p, p+x, p+y each tick
	Arpeggio EffectType = iota
	// SlideUp 1xx: upspeed - period is decreased by xx each tick
	SlideUp
	// SlideDown 2xx: downspeed - period is increased by xx each tick
	SlideDown
	// Portamento 3xx: up/down speed
	Portamento
	// Vibrato 4xy: x-speed, y-depth
	Vibrato
	// PortamentoVolSlide 5xy: x-upspeed, y-downspeed
	PortamentoVolSlide
	// VibratoVolSlide 6xy: x-upspeed, y-downspeed
	VibratoVolSlide
	// Tremolo 7xy: x-speed,   y-depth
	Tremolo
	// NotUsed8 value 8, unused
	NotUsed8
	// SetSampleOffset 9xx: offset (23 -> 2300)
	SetSampleOffset
	// VolSlide Axy: x-upspeed, y-downspeed
	VolSlide
	// PositionJump Bxx: songposition
	PositionJump
	// SetVol  Cxx: volume, 00-40
	SetVol
	// PatternBreak Dxx: break position in next patt
	PatternBreak
	// Extended Exy: see below...
	Extended
	// SetSpeed  Fxx: speed (00-1F) / tempo (20-FF)
	SetSpeed

	// SetFilter E0x: 0-filter on, 1-filter off
	SetFilter
	// FineSlideUp E1x: value
	FineSlideUp
	// FineSlideDown E2x: value
	FineSlideDown
	// GlissandoControl E3x: 0-off, 1-on (use with tonep.)
	GlissandoControl
	// SetVibratoWaveform E4x: 0-sine, 1-ramp down, 2-square
	SetVibratoWaveform
	// SetLoop E5x: set loop point
	SetLoop
	// JumpToLoop E6x: jump to loop, play x times
	JumpToLoop
	// SetTremoloWaveform E7x: 0-sine, 1-ramp down. 2-square
	SetTremoloWaveform
	// NotUsedE8 unused extended command
	NotUsedE8
	// RetrigNote E9x: retrig from note + x vblanks
	RetrigNote
	// FineVolSlideUp EAx: add x to volume
	FineVolSlideUp
	// FineVolSlideDown EBx: subtract x from volume
	FineVolSlideDown
	// NoteCut ECx: cut from note + x vblanks
	NoteCut
	// NoteDelay EDx: delay note x vblanks
	NoteDelay
	// PatternDelay EEx: delay pattern x notes
	PatternDelay
	// InvertLoop EFx: speed
	InvertLoop
)

//go:generate stringer -type=EffectType

// Effect is an effect/command (encoded as part of a note, but may affect the whole song)
type Effect struct {
	EffType EffectType
	EffCode uint16
}

// Par returns the parameter byte in its entirety
func (e Effect) Par() byte {
	return byte(e.EffCode & 0xFF)
}

// ParX returns the first nibble of the parameter byte
func (e Effect) ParX() byte {
	return byte(e.EffCode & 0xF0 >> 4)
}

// ParY returns the second nibble of the parameter byte
func (e Effect) ParY() byte {
	return byte(e.EffCode & 0x0F)
}

// Instrument represents an instrument used in a MOD file, including the sample data
type Instrument struct {
	Num      int
	Name     string
	Len      int
	Finetune int
	Volume   int
	RepStart int
	RepLen   int
	Offset   int
	Sample   []int8
}

// Note is an individual note, containing an Instrument, a Period and an Effect (with parameters)
type Note struct {
	InsNum int
	Ins    *Instrument
	Period int
	Effect
}

// Pattern is a 2-dimensional slice of Notes (lines x channels)
type Pattern [][]Note

// Module stores a complete MOD file
type Module struct {
	FileName      string
	Name          string
	Signature     [4]byte
	InstrTableLen int
	PatternCnt    int
	Instruments   [32]Instrument
	PatternTable  []int
	Patterns      [][][]Note
}

// Info prints information on the module file
func (m Module) Info() {
	fmt.Println("FileName:", m.FileName)
	fmt.Println("Name:", m.Name)
	fmt.Printf("Signature: %#v %s\n", m.Signature, string(m.Signature[0:4]))
	fmt.Println("Patterns (used):", len(m.Patterns))
	fmt.Println("Pattern sequence:", m.PatternTable)
	fmt.Println("Instruments:")
	for idx, ins := range m.Instruments {
		if ins.Len == 0 {
			continue
		}
		fmt.Printf("    %d %s : Offs %x, Len %x, RepS %x, RepL %x; Finetune %d, Vol %d\n",
			idx, ins.Name, ins.Offset, ins.Len, ins.RepStart, ins.RepLen, ins.Finetune, ins.Volume)
	}

	EffStats := make([]int, 32)
	for _, pattern := range m.Patterns {
		for _, line := range pattern {
			for _, note := range line {
				if note.EffType == Arpeggio && note.Par() == 0 {
					// Arpeggio effect (0) only counts if it has params
					continue
				}
				EffStats[note.EffType]++
			}
		}
	}
	fmt.Print("Effect counts: ")
	for eff, cnt := range EffStats {
		if cnt == 0 {
			continue
		}
		fmt.Printf("%v: %d; ", EffectType(eff), cnt)
	}
	fmt.Println()
	fmt.Println()
}

// ReadModFile reads the full MOD file given by fn and loads the data into the relevant objects
func ReadModFile(fn string) (mod Module, err error) {
	mod.FileName = fn
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return
	}

	// Module Name
	mod.Name = strings.Trim(string(data[0:20]), " \t\n\v\f\r\x00")

	// Signature (also tells us the number of instruments)
	copy(mod.Signature[0:4], data[1080:1084])
	// These are the default parameters for "original" SoundTracker modules (without signature)
	mod.InstrTableLen = 31
	signatureLen := 4
	for _, c := range mod.Signature {
		// if the signature is not an ASCII string, we have an old module with 15 instruments
		// FIXME: set the parameters depending on the various known signatures
		if c < 32 {
			mod.InstrTableLen = 15
			signatureLen = 0 // in old modules without "M.K." (or similar) signature, there is no space for it either. Duh...
		}
	}

	// Pattern Table (have to read this first because this tells us the number of patterns)
	patternTableOffset := 20 + mod.InstrTableLen*30 + 2
	patternTableLen := int(data[20+mod.InstrTableLen*30 /*+1*/]) // 20+31*30
	if patternTableLen > 128 {
		patternTableLen = 128 // some MOD files (e.g. BeatWave.mod) have patternTableLen > 128, which is illegal!
	}
	mod.PatternTable = make([]int, patternTableLen)
	for i := 0; i < patternTableLen; i++ {
		mod.PatternTable[i] = int(data[patternTableOffset+i])
		if mod.PatternTable[i] > mod.PatternCnt {
			mod.PatternCnt = mod.PatternTable[i] + 1
		}
	}
	//fmt.Printf("offs %x, cnt %d, tableLen %d, %+v\n", patternTableOffset, mod.PatternCnt, patternTableLen, mod.PatternTable)

	// Instruments
	// We read the samples from the end of the file - this assumes that there is no additional data at the end of the file.
	// Getting the sample offset from the previous data is unreliable because there may be patterns which are not in the pattern table.
	mod.Instruments[0] = Instrument{Num: 0, Name: "NOP"}
	sampleOffset := len(data)
	for i := mod.InstrTableLen; i > 0; i-- {
		instrOffset := 20 + (i-1)*30
		mod.Instruments[i], err = ReadInstrument(data[instrOffset : instrOffset+30])
		mod.Instruments[i].Num = i
		if mod.Instruments[i].Len == 0 {
			continue
		}
		sampleOffset -= mod.Instruments[i].Len
		mod.Instruments[i].Offset = sampleOffset
		mod.Instruments[i].Sample = make([]int8, mod.Instruments[i].Len)
		//copy(ins.Sample, sampleData[0:ins.Len]) -- doesn't work with byte -> int8; FIXME: faster version?!
		for j := range mod.Instruments[i].Sample {
			mod.Instruments[i].Sample[j] = int8(data[sampleOffset+j])
		}

	}

	// Patterns
	mod.Patterns = make([][][]Note, mod.PatternCnt)
	patternsOffset := 20 + mod.InstrTableLen*30 + 2 + 128 + signatureLen
	//fmt.Printf("PatternsOffset %x:\n", patternsOffset)
	for i := range mod.Patterns {
		mod.Patterns[i] = make([][]Note, 64)
		//fmt.Printf("\n\nPattern %d:\n", i)
		for j := range mod.Patterns[i] {
			mod.Patterns[i][j] = make([]Note, 4)
			for k := range mod.Patterns[i][j] {
				noteOffset := patternsOffset + ((i*64+j)*4+k)*4
				mod.Patterns[i][j][k] = ReadNote(data[noteOffset:noteOffset+4], &mod)
			}
			//fmt.Println(mod.Patterns[i][j][0], mod.Patterns[i][j][1], mod.Patterns[i][j][2], mod.Patterns[i][j][3])
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

// ReadInstrument constructs an instrument from the given instrData slice
func ReadInstrument(instrData []byte) (ins Instrument, err error) {
	ins.Name = strings.Trim(string(instrData[0:22]), " \t\n\v\f\r\x00")

	ins.Len = int(instrData[22])<<9 | int(instrData[23])<<1

	// FIXME this is actually a signed nibble, but we are currently treating it as unsigned
	ins.Finetune = int(instrData[24] & 0x0F)

	ins.Volume = int(instrData[25])

	ins.RepStart = int(instrData[26])<<9 | int(instrData[27])<<1
	if instrData[29] > 1 {
		ins.RepLen = int(instrData[28])<<9 | int(instrData[29])<<1
	}
	if ins.Len == 0 {
		return
	}

	return
}

func (n Note) String() string {
	s := ""
	if n.Period == 0 {
		if n.InsNum == 0 && n.EffCode == 0 {
			return "p---i--e---"
		}
		s += "p---"
	} else {
		s += fmt.Sprintf("p%03d", n.Period)
	}
	s += fmt.Sprintf("i%02xe%03x", n.InsNum, n.EffCode)
	return s
}

// Details prints detailed info about the given note
func (n Note) Details() {
	fmt.Println("Ins", n.InsNum)
	fmt.Println("Period", n.Period)
	fmt.Println("Effect", n.Effect)
}

// ReadNote constructs a Note from the given noteData slice
func ReadNote(noteData []byte, mod *Module) (n Note) {
	n.InsNum = int(noteData[0]&0xF0 | (noteData[2]&0xF0)>>4)
	if len(mod.Instruments) > n.InsNum {
		n.Ins = &mod.Instruments[n.InsNum]
	}

	bsl := []byte{noteData[0] & 0x0F, noteData[1]}
	n.Period = (int)(binary.BigEndian.Uint16(bsl))

	effNum := noteData[2] & 0x0F
	effPar := noteData[3]
	n.EffCode = uint16(effNum)<<8 | uint16(effPar)
	if effNum != 0xE {
		n.EffType = EffectType(effNum)
	} else {
		effSubNum := (noteData[3] & 0xF0) >> 4
		n.EffType = EffectType(16 + effSubNum)
	}

	return
}
