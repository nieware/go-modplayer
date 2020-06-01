# Effects implementation status

0 ✖ Normal play or Arpeggio             0xy : x-first halfnote add, y-second           pitch (w. note assignment)
1 ✔️ Slide Up                            1xx : upspeed                                  pitch
2 ✔️ Slide Down                          2xx : downspeed                                pitch
3 ✖ Tone Portamento                     3xx : up/down speed                            pitch (w. note assignment)
4 ✖ Vibrato                             4xy : x-speed,   y-depth                       pitch (w. waveform)
5 ✖ Tone Portamento + Volume Slide      5xy : x-upspeed, y-downspeed                   pitch/volume (w. note assignment)
6 ✖ Vibrato + Volume Slide              6xy : x-upspeed, y-downspeed                   pitch/volume (w. waveform)
7 ✖ Tremolo                             7xy : x-speed,   y-depth                       volume (w. waveform)
8   NOT USED
9 ✖ Set SampleOffset                    9xx : offset (23 -> 2300)                      sample offset
A ✔️ VolumeSlide                         Axy : x-upspeed, y-downspeed                   volume
B ✖ Position Jump                       Bxx : songposition                             position
C ✔️ Set Volume                          Cxx : volume, 00-40                            volume
D ✔️ Pattern Break (FIXME)               Dxx : break position in next patt              position
E   E-Commands                          Exy : see below...
F ✔️ Set Speed                           Fxx : speed (00-1F) / tempo (20-FF)            speed
----------------------------------------------------------------------------
E0✖ Set Filter                          E0x : 0-filter on, 1-filter off
E1✖ FineSlide Up                        E1x : value
E2✖ FineSlide Down                      E2x : value
E3✖ Glissando Control                   E3x : 0-off, 1-on (use with tonep.)
E4✖ Set Vibrato Waveform                E4x : 0-sine, 1-ramp down, 2-square
E5✖ Set Loop                            E5x : set loop point
E6✖ Jump to Loop                        E6x : jump to loop, play x times
E7✖ Set Tremolo Waveform                E7x : 0-sine, 1-ramp down. 2-square
E8  NOT USED
E9✖ Retrig Note                         E9x : retrig from note + x vblanks
EA✖ Fine VolumeSlide Up                 EAx : add x to volume
EB✖ Fine VolumeSlide Down               EBx : subtract x from volume
EC✖ NoteCut                             ECx : cut from note + x vblanks
ED✖ NoteDelay                           EDx : delay note x vblanks
EE✖ PatternDelay                        EEx : delay pattern x notes
EF✖ Invert Loop                         EFx : speed

Base Effects: 06/14
Ext. Effects: 00/15

"If you need to implement volume artificially, just multiply by the volume and shift right 6 times." 
(TODO: multiply by volume and use 16 bit, saves right-shifting)

