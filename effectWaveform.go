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
	Type WaveformType
}
