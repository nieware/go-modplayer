package main

import "fmt"

// VolumeProcessor is responsible for calculating the current volume for a channel
// considering currently active effect(s)
type VolumeProcessor struct {
	volume  int // current volume
	volumeΔ int // volume delta (value to add/subtract for volume slides)

	EffectWaveform
}

func (vpu *VolumeProcessor) VolumeFromNote(note Note) {
	if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
		vpu.volume = note.Ins.Volume
		vpu.volumeΔ = 0
	}

	switch note.EffType {
	case VolSlide, PortamentoVolSlide, VibratoVolSlide:
		if note.ParX() > 0 {
			vpu.volumeΔ = int(note.ParX())
		} else {
			vpu.volumeΔ = -int(note.ParY())
		}
	// case Tremolo:
	case SetVol:
		vpu.volume = int(note.Par())
		vpu.volumeΔ = 0
	//case SetTremoloWaveform:
	case FineVolSlideUp:
		vpu.volume += int(note.ParY())
	case FineVolSlideDown:
		vpu.volume -= int(note.ParY())
	//case NoteCut:

	// CHECK: Arpeggio is 0 - reset only if data present?!
	case Arpeggio, SlideUp, SlideDown, Portamento, Vibrato, FineSlideUp, FineSlideDown:
		vpu.volumeΔ = 0
	}
}

func (vpu *VolumeProcessor) VolumeOnTick(curTick int) {
	if vpu.volumeΔ != 0 {
		vpu.volume += vpu.volumeΔ // FIXME: not sure if this is correct, seems to be too fast!
		if vpu.volume > 64 {
			vpu.volume = 64
		}
		if vpu.volume < 0 {
			vpu.volume = 0
		}
		fmt.Println("vol", vpu.volume)
	}

}
