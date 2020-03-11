package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/hajimehoshi/oto"
)

var ctx *oto.Context
var (
	sampleRate      = flag.Int("samplerate", 16574 /*4143*/, "sample rate")
	channelNum      = flag.Int("channelnum", 1, "number of channel")
	bitDepthInBytes = flag.Int("bitdepthinbytes", 2, "bit depth in bytes")
)

// SamplePlayer plays a single sample
type SamplePlayer struct {
	Instrument
	pos, subpos, rate int
	ended             bool
}

// NewSamplePlayer creates a SamplaPlayer object for the instrument ins
func NewSamplePlayer(ins Instrument) *SamplePlayer {
	return &SamplePlayer{
		Instrument: ins,
		pos:        0,
		subpos:     0,
		rate:       1,
		ended:      false,
	}
}

// Read implements the Reader interface for SamplePlayer
func (sp *SamplePlayer) Read(buf []byte) (int, error) {
	if sp.ended {
		fmt.Println("EOF")
		return 0, io.EOF
	}

	var bufLen = len(buf)
	var maxRate = 4
	for bufIdx := 0; bufIdx < len(buf); bufIdx++ {
		buf[bufIdx] = sp.Sample[sp.pos]

		sp.subpos++
		if sp.subpos >= sp.rate {
			sp.subpos = 0
			sp.pos++
		}
		if sp.pos >= len(sp.Sample) {
			sp.subpos = 0
			sp.pos = 0
			sp.rate *= 2
		}
		if sp.rate > maxRate {
			bufLen = bufIdx + 1
			fmt.Println("read -> end", bufLen)
			sp.ended = true
			break
		}
	}
	return bufLen, nil
}

// PlaySample plays an instrument
func PlaySample(ins Instrument) error {
	p := ctx.NewPlayer()

	sp := NewSamplePlayer(ins)
	if _, err := io.Copy(p, sp); err != nil {
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
