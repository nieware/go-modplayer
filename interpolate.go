package main

// Interpolate interpolates the output waveform
func Interpolate(x0, x1, x2, x3 int8, t float32) int {
	return InterpolateLinear(x0, x1, x2, x3, t)
}

// InterpolateNone interpolates the output waveform by not interpolating at all
func InterpolateNone(x0, x1, x2, x3 int8, t float32) int {
	return int(x0)
}

// InterpolateLinear interpolates the output waveform with linear interpolation
func InterpolateLinear(x0, x1, x2, x3 int8, t float32) int {
	c0 := t * float32(x1)
	c1 := (1.0 - t) * float32(x2)
	return int((c0 + c1) / 2.0)
}

// InterpolateHermite4pt3oX interpolates the output waveform with Hermite interpolation
func InterpolateHermite4pt3oX(x0, x1, x2, x3 int8, t float32) int {
	c0 := float32(x1)
	c1 := float32(x2-x0) * .5
	c2 := float32(x0) - float32(x1)*2.5 + float32(x2)*2 - float32(x3)*.5
	c3 := float32(x3-x0)*.5 + float32(x1-x2)*1.5
	return int((((((c3 * t) + c2) * t) + c1) * t) + c0)
}
