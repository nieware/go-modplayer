package main

import "fmt"

// PeriodProcessor is responsible for calculating the current period (=pitch) for a channel
// considering currently active effect(s)
type PeriodProcessor struct {
	period       int // current period
	periodΔ      int // period delta (value to add/subtract for pitch slides)
	targetPeriod int // target period for "slide to note"

	EffectWaveform
}

// PeriodFromNote initializes the period (pitch) effects for the given note
func (ppu *PeriodProcessor) PeriodFromNote(note Note) {
	if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
		// FIXME: check if Portamento effects contain an instrument? Then we need to ignore it here...
		ppu.period = note.Period
		ppu.periodΔ = 0
	}

	switch note.EffType {
	case Arpeggio:
		ppu.periodΔ = 0
	case SlideUp:
		ppu.periodΔ = -int(note.Par())
	case SlideDown:
		ppu.periodΔ = int(note.Par())
	case Portamento:
		if note.Par() != 0 {
			ppu.targetPeriod = note.Period
			if note.Period > ppu.period {
				ppu.periodΔ = int(note.Par())
			} else {
				ppu.periodΔ = -int(note.Par())
			}
		}
	//case Vibrato:
	case FineSlideUp:
		ppu.period -= int(note.ParY())
	case FineSlideDown:
		ppu.period += int(note.ParY())
	//case GlissandoControl:
	//case SetVibratoWaveform:
	case VibratoVolSlide:
		ppu.periodΔ = 0
	case PortamentoVolSlide:
		// TODO: reset vibrato!
	case Tremolo, VolSlide, SetVol, FineVolSlideUp, FineVolSlideDown, NoteCut:
		ppu.periodΔ = 0
		ppu.targetPeriod = 0
	}

}

// PeriodOnTick computes the period value for the given tick
func (ppu *PeriodProcessor) PeriodOnTick(curTick int) {
	if ppu.periodΔ != 0 {
		// FIXME: check period limits!
		if ppu.targetPeriod != 0 && intAbs(ppu.targetPeriod-ppu.period) < intAbs(ppu.periodΔ) {
			ppu.period = ppu.targetPeriod
			ppu.periodΔ = 0
		}
		ppu.period += ppu.periodΔ
		fmt.Println("per", ppu.period)
	}
}

func intAbs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
