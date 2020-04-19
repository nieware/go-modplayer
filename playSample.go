package main

import (
	"io"
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
		step:       3546894.6 / float32(sampleRate*periods[0]),
		ended:      false,
	}
}

// Read implements the Reader interface for SamplePlayer
// Amiga PAL clock freq. 3546895
func (sp *SamplePlayer) Read(buf []byte) (int, error) {
	if sp.ended {
		//fmt.Println("EOF")
		return 0, io.EOF
	}

	var bufLen = len(buf)
	for bufIdx := 0; bufIdx < len(buf); bufIdx++ {
		buf[bufIdx] = byte(sp.Sample[int(sp.pos)])

		sp.pos += sp.step
		if int(sp.pos) >= len(sp.Sample) {
			sp.curPeriod++
			sp.pos = 0
			if sp.curPeriod >= len(sp.periods) {
				bufLen = bufIdx + 1
				//fmt.Println("read -> end", bufLen)
				sp.ended = true
				break
			}
			sp.step = 3546894.6 / float32(sampleRate*sp.periods[sp.curPeriod])
			//sp.step /= 10
			//fmt.Println(sp.periods[sp.curPeriod], sp.step)
		}
	}
	return bufLen, nil
}

// PlaySample plays an instrument
func PlaySample(ins Instrument) error {
	p := ctx.NewPlayer()

	sp := NewSamplePlayer(ins, []int{856, 428, 214})
	if _, err := io.Copy(p, sp); err != nil {
		return err
	}
	if err := p.Close(); err != nil {
		return err
	}
	return nil
}
