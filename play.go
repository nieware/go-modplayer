package main

import (
	"fmt"
	"io"
	"math"

	"github.com/hajimehoshi/oto"
)

// 1/214 .. 16574.27
// 1/p   .. y
// y = ((1/p)*16574.27) / (1/214) = (16574.27/p) * 214 = 16574.27 * 214 / p

// p = 428 --> y = 3546894.6 / 428 = 8287.13
// step = samplerate/y = (samplerate * p) / 3546894.6

var ctx *oto.Context

const (
	sampleRate      = 48000
	channelNum      = 2
	bitDepthInBytes = 2
	bufferSize      = 4096
)

// Speed holds all the parameters which affect the speed of playing a MOD file
type Speed struct {
	Tempo int // play speed part 1: number of ticks per pattern line (default 6)
	BPM   int // play speed part 2: so-called "beats per minute", but actually freq = curBPM * 0,4 Hz (default 125)
	SPB   int // samples per tick (depends on the sample rate we are playing at)
}

// Position holds all the parameters which determine the current play position in a MOD file
type Position struct {
	curPattern int // cur play position part 1: the pattern table index currently played
	curLine    int // cur play position part 2: the position inside the pattern
	curTick    int // cur play position part 3: the current tick (curTempo gives the number of ticks until the next pattern line)
	curTiming  int // cur play position part 4: the number of samples left until the next tick (depends on the sample rate we are playing at)
}

// Player plays a mod file
type Player struct {
	Module

	Position

	Speed

	chans []Channel // the channels for playing
	ended bool      // indicates whether playing has ended
}

// Channel is an individual channel of a Player
type Channel struct {
	index           int     // the number of this channel
	active          bool    // is the channel currently playing something? Set to false if the sample has "played out"
	note            *Note   // currently playing note
	pos, step       float32 // the position inside the sample and the step with which to advance the position
	firstTickOfNote bool    // is this the first tick where we play this note?
	tickCnt         byte    // tick counter for note retrig/cut/delay

	PeriodProcessor // this channel's "PPU" (period/pitch processing unit)
	VolumeProcessor // this channel's "VPU" (volume processing unit)
}

// NewPlayer creates a Player object for the module mod
func NewPlayer(mod Module) *Player {
	p := &Player{
		Module: mod,
		chans:  make([]Channel, 4), // we currently only support 4-channel modules
	}
	p.Speed = Speed{
		Tempo: 6,
		BPM:   125,
		SPB:   int(float64(sampleRate) / (.4 * 125)),
	}

	for i := range p.chans {
		p.chans[i].index = i
	}
	return p
}

// SetPeriod sets the internal "step" according to the given period value.
func (ch *Channel) SetPeriod(period int) {
	// Amiga PAL clock freq. 3546894.6
	ch.step = 3546894.6 / float32(sampleRate*period)
}

// OnNote starts a new note on a channel if the note contains an instrument.
// Some notes only contain effects, which are then applied on the currently playing note.
func (ch *Channel) OnNote(note Note, speed Speed) {
	if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
		// if we have an instrument, start playing a new note
		ch.note = &note
		ch.firstTickOfNote = true
		ch.active = true
		ch.pos = 0
	}
	// If we have an effect, set it on new or currently playing note
	ch.PeriodFromNote(note, speed)
	ch.SetPeriod(ch.period)
	ch.VolumeFromNote(note)

	/*if ch.firstTickOfNote {
		fmt.Printf("ch %d -> active, step %f\n", ch.index, ch.step)
	} //*/

	switch note.EffType {
	case SetSampleOffset:
		if ch.active {
			ch.pos = float32(int(note.Effect.Par()) << 9)
		}
	case RetrigNote, NoteCut, NoteDelay:
		ch.tickCnt = note.Effect.ParY()
		ch.active = note.EffType != NoteDelay
	}

	if ch.pos < 1 {
		ch.pos = 1
	}
}

// OnTick computes the necessary parameters for the given tick
func (ch *Channel) OnTick(curTick int) {
	if !ch.firstTickOfNote {
		ch.PeriodOnTick(curTick)
		ch.SetPeriod(ch.period)
		ch.VolumeOnTick(curTick)
	}
	ch.firstTickOfNote = false

	ch.tickCnt--
	if ch.note == nil || ch.note.Ins == nil {
		return
	}
	switch ch.note.EffType {
	case RetrigNote:
		if ch.tickCnt == 0 {
			ch.pos = 1
			ch.tickCnt = ch.note.Effect.ParY()
		}
	case NoteCut:
		if ch.tickCnt == 0 {
			ch.active = false
		}
	case NoteDelay:
		if ch.tickCnt == 0 {
			ch.pos = 1 // just to be sure...
			ch.active = true
		}
	}

}

// GetNextSample advances the internal counter and returns the value for the next sample to be
// played on this channel.
func (ch *Channel) GetNextSample() int {
	if !ch.active {
		return 0
	}
	if ch.note == nil || ch.note.Ins == nil || ch.note.Ins.Sample == nil {
		fmt.Println("ch.note/ch.note.Ins/ch.note.Ins.Sample nil!")
		return 0
	}
	pos64, subpos64 := math.Modf(float64(ch.pos))
	pos := int(pos64)
	val := Interpolate(
		ch.note.Ins.Sample[pos-1], ch.note.Ins.Sample[pos],
		ch.note.Ins.Sample[pos+1], ch.note.Ins.Sample[pos+2],
		float32(subpos64),
	)
	ch.pos += ch.step
	if ch.pos >= float32(len(ch.note.Ins.Sample)-2) {
		if ch.note.Ins.RepLen > 2 {
			ch.pos = float32(ch.note.Ins.RepStart + 2) // repeat TODO: handle RepLen - but how?!
		} else {
			ch.active = false // played out
		}
	}

	//fmt.Println(ch.pos, ch.step, val, ch.volume)
	return val * ch.volume
}

// GetNextSamples advances the internal counter and returns the values for the next samples to be
// played (for left and right stereo channel).
func (p *Player) GetNextSamples() (int, int) {
	// if we are at the start of a new line, init the notes and effects
	if p.curTick == 0 && p.curTiming == 0 {
		patt := p.Module.PatternTable[p.curPattern]
		notes := p.Module.Patterns[patt][p.curLine]
		fmt.Println(notes[0], notes[1], notes[2], notes[3])

		// process "pattern break" before playing the notes
		pars, isPatternBreak := findEffect(notes, PatternBreak)
		if isPatternBreak {
			p.curPattern++
			p.curLine = int(pars)
			patt = p.Module.PatternTable[p.curPattern]
			notes = p.Module.Patterns[patt][p.curLine]
		}

		for i := range p.chans {
			note := p.Module.Patterns[patt][p.curLine][i]
			if note.EffCode != 0 {
				fmt.Printf("Ch %d: Eff %v\n", i, note.EffType)
			}
			p.chans[i].OnNote(note, p.Speed)

			switch note.EffType {
			// we only take care of global position/timing commands here, the rest are handled by the channel or its PPU/VPU
			/*case PatternBreak:
			p.curPattern++
			p.curTiming, p.curTick = 0, 0
			p.curLine = int(note.Pars) // fixme: apparently Par is "decimal" (BCD?)*/
			case SetSpeed:
				if note.Par() <= 0x1F {
					p.Tempo = int(note.Par())
				} else {
					p.BPM = int(note.Par())
				}
			}
		}
	}

	p.curTiming++
	if p.curTiming >= p.SPB {
		p.curTiming = 0
		p.curTick++

		// some effects have to be reapplied with each tick
		for i := range p.chans {
			p.chans[i].OnTick(p.curTick)
		}
	}
	if p.curTick >= p.Tempo {
		p.curTiming, p.curTick = 0, 0
		p.curLine++
	}
	if p.curLine >= 64 { // pattern len
		p.curTiming, p.curTick, p.curLine = 0, 0, 0
		p.curPattern++
	}
	if p.curPattern >= len(p.Module.PatternTable) {
		p.ended = true
		return 0, 0
	}

	// mix the current value from all channels
	var mix [2]int
	var chanTab = [4]int{0, 1, 1, 0}
	for i := range p.chans {
		mix[chanTab[i]] += p.chans[i].GetNextSample()
	}
	return mix[0], mix[1]
}

// Read implements the Reader interface for Player
func (p *Player) Read(buf []byte) (int, error) {
	if p.ended {
		fmt.Println("EOF")
		return 0, io.EOF
	}

	var bufLen = len(buf)
	for bufIdx := 0; bufIdx < len(buf); bufIdx += bitDepthInBytes * channelNum {
		l, r := p.GetNextSamples()

		if p.ended {
			bufLen = bufIdx + 1
			fmt.Println("read -> end", p.curPattern, len(p.Module.PatternTable), bufLen)
			break
		}

		if bitDepthInBytes == 1 {
			// 8-bit: right-shift the mixed value to avoid overflow (TODO this depends on the number of channels)
			buf[bufIdx] = byte(l>>1 + 127)
			buf[bufIdx+1] = byte(r>>1 + 127)
		} else {
			// 16-bit: split the value in 2 bytes
			buf[bufIdx] = byte(l & 0x00FF)
			buf[bufIdx+1] = byte((l & 0xFF00) >> 8)
			buf[bufIdx+2] = byte(r & 0x00FF)
			buf[bufIdx+3] = byte((r & 0xFF00) >> 8)
		}
	}
	return bufLen, nil
}

// Play plays a module
func Play(mod Module) error {
	p := ctx.NewPlayer()

	mp := NewPlayer(mod)
	if _, err := io.Copy(p, mp); err != nil {
		return err
	}
	if err := p.Close(); err != nil {
		return err
	}
	return nil
}

func init() {
	var err error
	ctx, err = oto.NewContext(sampleRate, channelNum, bitDepthInBytes, bufferSize)
	if err != nil {
		panic("Unable to initialize audio, error " + err.Error())
	}

}
