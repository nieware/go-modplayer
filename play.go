package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/hajimehoshi/oto"
)

var ctx *oto.Context
var (
	sampleRate      = flag.Int("samplerate", 4143, "sample rate")
	channelNum      = flag.Int("channelnum", 1, "number of channel")
	bitDepthInBytes = flag.Int("bitdepthinbytes", 2, "bit depth in bytes")
)

// SamplePlayer plays a single sample
type SamplePlayer struct {
	Instrument
	pos int
}

// NewSamplePlayer creates a SamplaPlayer object for the instrument ins
func NewSamplePlayer(ins Instrument) *SamplePlayer {
	return &SamplePlayer{
		Instrument: ins,
		pos:        0,
	}
}

// Read implements the Reader interface for SamplePlayer
func (sp *SamplePlayer) Read(buf []byte) (int, error) {
	if sp.pos == sp.Len {
		fmt.Println("EOF")
		return 0, io.EOF
	}

	if sp.Len-sp.pos <= len(buf) {
		n := copy(buf, sp.Sample[sp.pos:])
		sp.pos += n
		fmt.Println("read -> end", n)
		return n, nil
	}
	n := copy(buf, sp.Sample[sp.pos:sp.pos+len(buf)])
	sp.pos += n
	fmt.Println("read", n)
	return n, nil
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
