package main

import "fmt"

// PeriodProcessor is responsible for calculating the current period (=pitch) for a channel
// considering currently active effect(s)
type PeriodProcessor struct {
	period       int   // current period
	periodΔ      int   // period delta (value to add/subtract for pitch slides)
	arpeggio     []int // periods for arpeggio
	arpeggioIdx  int   // index in arpeggio array
	targetPeriod int   // target period for "slide to note"
	glissando    bool  // glissando flag (true - "slide to note" slides in halfnotes)

	Ins *Instrument

	EffectWaveform
}

// PeriodFromNote initializes the period (pitch) effects for the given note
func (ppu *PeriodProcessor) PeriodFromNote(note Note, speed Speed) {
	resetSlide := true
	resetVibrato := true
	if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
		// FIXME: check if Portamento effects contain an instrument? Then we need to ignore it here...
		ppu.period = note.Period
		ppu.Ins = note.Ins
	}

	switch note.EffType {
	case Arpeggio:
		switch {
		case note.ParX() > 0 && note.ParY() > 0:
			ppu.arpeggio = []int{ppu.Ins.IncDec(ppu.period, note.ParX()), ppu.Ins.IncDec(ppu.period, note.ParY())}
		case note.ParX() > 0:
			ppu.arpeggio = []int{note.Ins.IncDec(ppu.period, note.ParX())}
		default:
			ppu.arpeggio = []int{}
		}
	case SlideUp:
		ppu.periodΔ = -note.Par()
		resetSlide = false
	case SlideDown:
		ppu.periodΔ = note.Par()
		resetSlide = false
	case Portamento:
		if note.Par() != 0 {
			if note.Period != 0 {
				ppu.targetPeriod = note.Period
				fmt.Println("slide -> ", ppu.targetPeriod)
			}
			if note.Period > ppu.period {
				ppu.periodΔ = note.Par()
			} else {
				ppu.periodΔ = -note.Par()
			}
		}
		resetSlide = false
	case Vibrato, VibratoVolSlide:
		ppu.InitVibratoWaveform(note.ParX(), note.ParY(), note.Period, *note.Ins)
		resetVibrato = false
	case FineSlideUp:
		ppu.period -= note.ParY()
	case FineSlideDown:
		ppu.period += note.ParY()
	case GlissandoControl:
		ppu.glissando = note.ParY() == 1
	case SetVibratoWaveform:
		ppu.DecodeWaveformType(note.ParY())
	case PortamentoVolSlide:
		resetSlide = false
		// TODO: reset vibrato!
	case Tremolo, VolSlide, SetVol, FineVolSlideUp, FineVolSlideDown, NoteCut:
		ppu.periodΔ = 0
		ppu.targetPeriod = 0
	}

	if resetSlide {
		ppu.periodΔ = 0
	}
	if resetVibrato {
		ppu.EffectWaveform.Active = false
	}
}

// PeriodOnTick computes the period value for the given tick
func (ppu *PeriodProcessor) PeriodOnTick(curTick int) {
	if ppu.periodΔ != 0 {
		// FIXME: check period limits!
		if ppu.targetPeriod != 0 && intAbs(ppu.targetPeriod-ppu.period) < intAbs(ppu.periodΔ) {
			fmt.Println("end slide")
			ppu.period = ppu.targetPeriod
			ppu.periodΔ = 0
		}
		ppu.period += ppu.periodΔ
		fmt.Println("per", ppu.period)
	}

	ppu.arpeggioIdx++
	if ppu.arpeggioIdx >= len(ppu.arpeggio) {
		ppu.arpeggioIdx = 0
	}
}

// Next gets the period value for the next sample
func (ppu *PeriodProcessor) Next() int {
	if ppu.arpeggioIdx > 0 {
		return ppu.arpeggio[ppu.arpeggioIdx]
	}
	return ppu.period + ppu.DoStep()
}

func intAbs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
