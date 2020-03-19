package main

import (
	"flag"
	"fmt"
	"io"

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
	channelNum      = flag.Int("channelnum", 1, "number of channel")
	bitDepthInBytes = flag.Int("bitdepthinbytes", 2, "bit depth in bytes")
)

// SamplePlayer plays a single sample
type SamplePlayer struct {
	Instrument
	periods   []int
	curPeriod int
	pos, step float32
	ended     bool
}

// NewSamplePlayer creates a SamplaPlayer object for the instrument ins
func NewSamplePlayer(ins Instrument, periods []int) *SamplePlayer {
	return &SamplePlayer{
		Instrument: ins,
		periods:    periods,
		step:       3546894.6 / float32(*sampleRate*periods[0]),
		ended:      false,
	}
}

// Read implements the Reader interface for SamplePlayer
// Amiga PAL clock freq. 3546895
func (sp *SamplePlayer) Read(buf []byte) (int, error) {
	if sp.ended {
		fmt.Println("EOF")
		return 0, io.EOF
	}

	var bufLen = len(buf)
	for bufIdx := 0; bufIdx < len(buf); bufIdx++ {
		buf[bufIdx] = sp.Sample[int(sp.pos)]

		sp.pos += sp.step
		if int(sp.pos) >= len(sp.Sample) {
			sp.pos = 0
			sp.curPeriod++
			if sp.curPeriod >= len(sp.periods) {
				bufLen = bufIdx + 1
				fmt.Println("read -> end", bufLen)
				sp.ended = true
				break
			}
			sp.step = 3546894.6 / float32(*sampleRate*sp.periods[sp.curPeriod])
			//sp.step /= 10
			fmt.Println(sp.periods[sp.curPeriod], sp.step)
		}
	}
	return bufLen, nil
}

// PlaySample plays an instrument
func PlaySample(ins Instrument) error {
	p := ctx.NewPlayer()

	sp := NewSamplePlayer(ins, []int{320, 285, 254, 214, 254})
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
