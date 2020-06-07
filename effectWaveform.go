package main

type WaveformType int

const (
	Sine WaveformType = iota
	RampDown
	Square
	Random
)

type EffectWaveform struct {
	Type WaveformType
}
