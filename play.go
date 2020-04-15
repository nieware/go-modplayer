package main

import (
	"flag"
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
var (
	sampleRate = flag.Int("samplerate",
		48000,
		//16574,
		//4143,
		"sample rate")
	channelNum      = flag.Int("channelnum", 2, "number of channels")
	bitDepthInBytes = flag.Int("bitdepthinbytes", 2, "bit depth in bytes")
)

// Player plays a mod file
type Player struct {
	Module
	curPattern int       // the pattern table index currently played
	curLine    int       // the position inside the pattern
	curBeat    int       // the current beat (curTempo gives the number of beats until the next pattern line)
	curTiming  int       // the number of samples left until the next beat
	curSPB     int       // samples per beat
	curBPM     int       // so-called "beats per minute", but actually freq = curBPM * 0,4 Hz
	curTempo   int       // the number of beats per pattern line
	chans      []Channel // the channels for playing
	ended      bool
}

// Channel is an individual channel of a Player
type Channel struct {
	active    bool        // is the channel currently playing something? Set to false if the sample has "played out"
	ins       *Instrument // the instrument currently played
	pos, step float32     // the position inside the sample and the step with which to advance the position
	period    int         // current period
	periodΔ   int         // period delta (value to add/subtract for pitch bending)
	volume    int         // current volume
	volumeΔ   int         // volume delta (value to add/subtract for volume slides)
}

// NewPlayer creates a Player object for the module mod
func NewPlayer(mod Module) *Player {
	fmt.Println("curSPB", int(float64(*sampleRate)/(.4*125)))
	return &Player{
		Module:   mod,
		chans:    make([]Channel, 4), // we currently only support 4-channel modules
		curBPM:   125,
		curTempo: 6,
		curSPB:   int(float64(*sampleRate) / (.4 * 125)),
	}
}

// InterpolateHermite4pt3oX interpolates the output waveform
func (p *Player) InterpolateHermite4pt3oX(x0, x1, x2, x3 int8, t float32) int {
	return int(x1)

	// doesn't seem to make a difference (actually sounds slightly worse?!) - maybe not correct...
	c0 := float32(x1)
	c1 := float32(x2-x0) * .5
	c2 := float32(x0) - float32(x1)*2.5 + float32(x2)*2 - float32(x3)*.5
	c3 := float32(x3-x0)*.5 + float32(x1-x2)*1.5
	return int((((((c3 * t) + c2) * t) + c1) * t) + c0)
}

// Read implements the Reader interface for Player
func (p *Player) Read(buf []byte) (int, error) {
	if p.ended {
		fmt.Println("EOF")
		return 0, io.EOF
	}

	var bufLen = len(buf)
	for bufIdx := 0; bufIdx < len(buf); bufIdx += *bitDepthInBytes * *channelNum {
		// if we are at the start of a new line, init the notes and effects
		if p.curBeat == 0 && p.curTiming == 0 {
			patt := p.Module.PatternTable[p.curPattern]
			notes := p.Module.Patterns[patt][p.curLine]
			fmt.Println(notes[0], notes[1], notes[2], notes[3])
			for i := range p.chans {
				note := p.Module.Patterns[patt][p.curLine][i]
				if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
					// if we have an instrument, start playing a new note
					p.chans[i].active = true
					p.chans[i].ins = note.Ins
					p.chans[i].pos = 1 // needed because of interpolation
					p.chans[i].period = note.Period
					p.chans[i].periodΔ = 0
					p.chans[i].volume = note.Ins.Volume
					p.chans[i].volumeΔ = 0
					// Amiga PAL clock freq. 3546894.6
					p.chans[i].step = 3546894.6 / float32(*sampleRate*p.chans[i].period)
					//fmt.Println("ch", i, "-> active, step", p.chans[i].step)
				}
				p.chans[i].periodΔ = 0
				if note.EffCode != 0 {
					fmt.Printf("Eff %v\n", note.Eff)
					// If we have an effect, set it on new or currently playing note
					switch note.Eff {
					case SlideUp:
						p.chans[i].periodΔ = -int(note.Pars)
					case SlideDown:
						p.chans[i].periodΔ = int(note.Pars)
					case SetVol:
						p.chans[i].volume = int(note.Pars)
					}
					fmt.Printf("N %#v Δ %d\n", note, p.chans[i].periodΔ)
				}
			}
		}

		p.curTiming++
		if p.curTiming >= p.curSPB {
			p.curTiming = 0
			p.curBeat++

			// some effects have to be reapplied with each beat
			for i := range p.chans {
				if p.chans[i].periodΔ != 0 {
					p.chans[i].period += p.chans[i].periodΔ
					fmt.Println("per", p.chans[i].period)
					p.chans[i].step = 3546894.6 / float32(*sampleRate*p.chans[i].period)
				}
			}
		}
		if p.curBeat >= p.curTempo {
			p.curTiming, p.curBeat = 0, 0
			p.curLine++
		}
		if p.curLine >= 64 { // pattern len
			p.curTiming, p.curBeat, p.curLine = 0, 0, 0
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
			//subpos = 0 // This disables the "interpolation"
			val := p.InterpolateHermite4pt3oX(
				ch.ins.Sample[pos-1], ch.ins.Sample[pos],
				ch.ins.Sample[pos+1], ch.ins.Sample[pos+2],
				float32(subpos64),
			)
			//val := int(float64(ch.ins.Sample[int(pos)])*(1-subpos) + float64(ch.ins.Sample[int(pos)+1])*subpos)
			mix[chanTab[i]] += val * ch.volume
			p.chans[i].pos += ch.step
			if p.chans[i].pos >= float32(len(ch.ins.Sample)-2) {
				p.chans[i].active = false // played out (TODO: repeat!)
				//fmt.Println("ch", i, "-> inactive")
			}
		}
		if *bitDepthInBytes == 1 {
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
	ctx, err = oto.NewContext(*sampleRate, *channelNum, *bitDepthInBytes, 4096)
	if err != nil {
		panic("Unable to initialize audio, error " + err.Error())
	}

}
