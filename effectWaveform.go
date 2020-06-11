package main

// WaveformType is the type of the envelope waveform used for vibrato/tremolo
type WaveformType int

const (
	// Sine - sine waveform
	Sine WaveformType = iota
	// RampDown - "sawtooth" waveform
	RampDown
	// Square - square wave
	Square
	// Random - select one of Sine/RampDown/Square randomly
	Random
)

// EffectWaveform contains the parameters for a waveform assigned to an effect
type EffectWaveform struct {
	Type    WaveformType
	CurType WaveformType
	Retrig  bool
}

// DecodeEffectWaveform creates an EffectWaveform struct from a "set waveform" command parameter
func DecodeEffectWaveform(par int) (ew EffectWaveform) {
	ew.Retrig = par >= 4
	switch par {
	case 0, 4:
		ew.Type = Sine
	case 1, 5:
		ew.Type = RampDown
	case 2, 6:
		ew.Type = Square
	case 3, 7:
		ew.Type = Random
	}
	return
}
