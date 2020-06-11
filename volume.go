package main

import "fmt"

// VolumeProcessor is responsible for calculating the current volume for a channel
// considering currently active effect(s)
type VolumeProcessor struct {
	volume  int // current volume
	volumeΔ int // volume delta (value to add/subtract for volume slides)

	EffectWaveform
}

// VolumeFromNote initializes the volume effects for the given note
func (vpu *VolumeProcessor) VolumeFromNote(note Note) {
	resetSlide := true
	if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
		vpu.volume = note.Ins.Volume
	}

	switch note.EffType {
	case VolSlide, PortamentoVolSlide, VibratoVolSlide:
		if note.Par() != 0 {
			if note.ParX() > 0 {
				vpu.volumeΔ = note.ParX()
			} else {
				vpu.volumeΔ = -note.ParY()
			}
		}
		resetSlide = false
	case Tremolo:
		// TODO
	case SetVol:
		vpu.volume = note.Par()
	case SetTremoloWaveform:
		vpu.EffectWaveform = DecodeEffectWaveform(note.ParY())
	case FineVolSlideUp:
		vpu.volume += note.ParY()
	case FineVolSlideDown:
		vpu.volume -= note.ParY()
	}

	if resetSlide {
		vpu.volumeΔ = 0
	}

}

// VolumeOnTick computes the period value for the given tick
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
