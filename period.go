package main

import "fmt"

type arpeggioEntry struct {
	period  int
	maxTick int
}

// PeriodProcessor is responsible for calculating the current period (=pitch) for a channel
// considering currently active effect(s)
type PeriodProcessor struct {
	period       int             // current period
	periodΔ      int             // period delta (value to add/subtract for pitch slides)
	arpeggio     []arpeggioEntry // periods for arpeggio
	targetPeriod int             // target period for "slide to note"
	glissando    bool            // glissando flag (true - "slide to note" slides in halfnotes)

	EffectWaveform
}

// PeriodFromNote initializes the period (pitch) effects for the given note
func (ppu *PeriodProcessor) PeriodFromNote(note Note, speed Speed) {
	resetSlide := true
	if note.Ins != nil && note.Ins.Sample != nil && note.Period > 0 {
		// FIXME: check if Portamento effects contain an instrument? Then we need to ignore it here...
		ppu.period = note.Period
	}

	switch note.EffType {
	case Arpeggio:
		if note.ParX() > 0 && note.ParY() > 0 {
			ppu.arpeggio = make([]arpeggioEntry, 3)
			div := speed.Tempo / 3
			maxTick := speed.Tempo - 2*div - 1
			ppu.arpeggio[0] = arpeggioEntry{note.Period, maxTick}
			ppu.arpeggio[1] = arpeggioEntry{note.IncDec(int(note.ParX())), maxTick + div}
			ppu.arpeggio[2] = arpeggioEntry{note.IncDec(int(note.ParY())), speed.Tempo - 1}
			fmt.Println(ppu.arpeggio)
		} else if note.ParX() > 0 {
			ppu.arpeggio = make([]arpeggioEntry, 2)
			div := speed.Tempo / 2
			maxTick := speed.Tempo - div - 1
			ppu.arpeggio[0] = arpeggioEntry{note.Period, maxTick}
			ppu.arpeggio[1] = arpeggioEntry{note.IncDec(int(note.ParX())), speed.Tempo - 1}
			fmt.Println(ppu.arpeggio)
		} else {
			ppu.arpeggio = make([]arpeggioEntry, 0)
		}
	case SlideUp:
		ppu.periodΔ = -int(note.Par())
		resetSlide = false
	case SlideDown:
		ppu.periodΔ = int(note.Par())
		resetSlide = false
	case Portamento:
		if note.Par() != 0 {
			ppu.targetPeriod = note.Period
			if note.Period > ppu.period {
				ppu.periodΔ = int(note.Par())
			} else {
				ppu.periodΔ = -int(note.Par())
			}
		}
		resetSlide = false
	case Vibrato, VibratoVolSlide:
		// TODO
	case FineSlideUp:
		ppu.period -= int(note.ParY())
	case FineSlideDown:
		ppu.period += int(note.ParY())
	case GlissandoControl:
		ppu.glissando = note.ParY() == 1
	case SetVibratoWaveform:
		ppu.EffectWaveform = DecodeEffectWaveform(note.ParY())
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

	//fmt.Printf("N %#v Δ %d\n", note, ppu.periodΔ)
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

	if len(ppu.arpeggio) > 0 {
		fmt.Println("cur", curTick)
		for _, arp := range ppu.arpeggio {
			if curTick > arp.maxTick {
				continue
			}
			ppu.period = arp.period
			fmt.Println("arp", ppu.period)
			break
		}
	}
}

func intAbs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
