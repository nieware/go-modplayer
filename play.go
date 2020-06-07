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

// Player plays a mod file
type Player struct {
	Module

	curPattern int // cur play position part 1: the pattern table index currently played
	curLine    int // cur play position part 2: the position inside the pattern
	curTick    int // cur play position part 3: the current tick (curTempo gives the number of ticks until the next pattern line)
	curTiming  int // cur play position part 4: the number of samples left until the next tick (depends on the sample rate we are playing at)

	curTempo int // play speed part 1: number of ticks per pattern line (default 6)
	curBPM   int // play speed part 2: so-called "beats per minute", but actually freq = curBPM * 0,4 Hz (default 125)
	curSPB   int // samples per tick (depends on the sample rate we are playing at)

	chans []Channel // the channels for playing
	ended bool      // indicates whether playing has ended
}

// Channel is an individual channel of a Player
type Channel struct {
	active          bool        // is the channel currently playing something? Set to false if the sample has "played out"
	ins             *Instrument // the instrument currently played
	pos, step       float32     // the position inside the sample and the step with which to advance the position
	firstTickOfNote bool        // is this the first tick where we play this note?

	PeriodProcessor
	VolumeProcessor
}

// NewPlayer creates a Player object for the module mod
func NewPlayer(mod Module) *Player {
	fmt.Println("curSPB", int(float64(sampleRate)/(.4*125)))
	return &Player{
		Module:   mod,
		chans:    make([]Channel, 4), // we currently only support 4-channel modules
		curTempo: 6,
		curBPM:   125,
		curSPB:   int(float64(sampleRate) / (.4 * 125)),
	}
}

func findEffect(notes []Note, eff EffectType) (byte, bool) {
	for _, note := range notes {
		if note.EffType == eff {
			return note.Par(), true
		}
	}
	return 0, false
}

// Read implements the Reader interface for Player
func (p *Player) Read(buf []byte) (int, error) {
	if p.ended {
		fmt.Println("EOF")
		return 0, io.EOF
	}

	var bufLen = len(buf)
	for bufIdx := 0; bufIdx < len(buf); bufIdx += bitDepthInBytes * channelNum {
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
				if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
					// FIXME: check if Portamento effects contain an instrument? Then we need to ignore it here...
					// if we have an instrument, start playing a new note
					p.chans[i].active = true
					p.chans[i].ins = note.Ins
					p.chans[i].pos = 1 // needed because of interpolation
					p.chans[i].firstTickOfNote = true
					// Amiga PAL clock freq. 3546894.6
					p.chans[i].step = 3546894.6 / float32(sampleRate*p.chans[i].period)
					//fmt.Println("ch", i, "-> active, step", p.chans[i].step)
				}
				if note.EffCode != 0 {
					fmt.Printf("Eff %v\n", note.EffType)
				}
				// If we have an effect, set it on new or currently playing note
				p.chans[i].PeriodFromNote(note)
				p.chans[i].VolumeFromNote(note)
				switch note.EffType {
				// we only take care of position/timing commands here, the rest are handled by PPU/VPU
				/*case PatternBreak:
				p.curPattern++
				p.curTiming, p.curTick = 0, 0
				p.curLine = int(note.Pars) // fixme: apparently Par is "decimal" (BCD?)*/
				case SetSpeed:
					if note.Par() <= 0x1F {
						p.curTempo = int(note.Par())
					} else {
						p.curBPM = int(note.Par())
					}
				}

				//fmt.Printf("N %#v Δ %d\n", note, p.chans[i].periodΔ)
			}
		}

		p.curTiming++
		if p.curTiming >= p.curSPB {
			p.curTiming = 0
			p.curTick++

			// some effects have to be reapplied with each tick
			for i := range p.chans {
				if !p.chans[i].firstTickOfNote {
					p.chans[i].PeriodOnTick(p.curTick)
					p.chans[i].step = 3546894.6 / float32(sampleRate*p.chans[i].period)
					p.chans[i].VolumeOnTick(p.curTick)
				}
				p.chans[i].firstTickOfNote = false
			}
		}
		if p.curTick >= p.curTempo {
			p.curTiming, p.curTick = 0, 0
			p.curLine++
		}
		if p.curLine >= 64 { // pattern len
			p.curTiming, p.curTick, p.curLine = 0, 0, 0
			p.curPattern++
		}
		if p.curPattern >= len(p.Module.PatternTable) {
			bufLen = bufIdx + 1
			fmt.Println("read -> end", p.curPattern, len(p.Module.PatternTable), bufLen)
			p.ended = true
			break
		}

		// mix the current value from all channels
		var mix [2]int
		var chanTab = [4]int{0, 1, 1, 0}
		for i, ch := range p.chans {
			if !ch.active {
				continue
			}
			if ch.ins == nil {
				fmt.Println("ch.ins nil!")
				continue
			}
			if ch.ins.Sample == nil {
				fmt.Println("ch.ins.Sample nil!")
				continue
			}
			pos64, subpos64 := math.Modf(float64(ch.pos))
			pos := int(pos64)
			val := Interpolate(
				ch.ins.Sample[pos-1], ch.ins.Sample[pos],
				ch.ins.Sample[pos+1], ch.ins.Sample[pos+2],
				float32(subpos64),
			)
			mix[chanTab[i]] += val * ch.volume
			p.chans[i].pos += ch.step
			if p.chans[i].pos >= float32(len(ch.ins.Sample)-2) {
				if ch.ins.RepLen > 2 {
					p.chans[i].pos = float32(ch.ins.RepStart + 2) // repeat TODO: handle RepLen - but how?!
				} else {
					p.chans[i].active = false // played out
				}
				//fmt.Println("ch", i, "-> inactive")
			}
		}
		if bitDepthInBytes == 1 {
			// 8-bit: right-shift the mixed value to avoid overflow (TODO this depends on the number of channels)
			buf[bufIdx] = byte(mix[0]>>1 + 127)
			buf[bufIdx+1] = byte(mix[1]>>1 + 127)
		} else {
			// 16-bit: split the value in 2 bytes
			buf[bufIdx] = byte(mix[0] & 0x00FF)
			buf[bufIdx+1] = byte((mix[0] & 0xFF00) >> 8)
			buf[bufIdx+2] = byte(mix[1] & 0x00FF)
			buf[bufIdx+3] = byte((mix[1] & 0xFF00) >> 8)
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
