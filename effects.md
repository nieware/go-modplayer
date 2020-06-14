# Effects implementation status

0 ✔️ Normal play or Arpeggio             0xy : x-first halfnote add, y-second           PS VR (w. note assignment, applied 2/3 times per line)
1 ✔️ Slide Up                            1xx : upspeed                                  PS VR (applied once per tick)
2 ✔️ Slide Down                          2xx : downspeed                                PS VR (applied once per tick)
3 ✔️ Tone Portamento                     3xx : up/down speed                            PS VR (applied once per tick)
4 ✖ Vibrato                             4xy : x-speed,   y-depth                       PS VR (w. waveform, applied continuously: amplitude y/16 halfnotes, (x*ticks)/64 cycles per line)   _flippin
5 ✔️ Tone Portamento + Volume Slide      5xy : x-upspeed, y-downspeed                   PK VS (cont. #3, applied once per tick)
6 ✔️ Vibrato + Volume Slide              6xy : x-upspeed, y-downspeed                   PK VS (cont. #4, applied once per tick)
7 ✖ Tremolo                             7xy : x-speed,   y-depth                       PR VS (w. waveform, applied continuously: amplitude y ???, (x*ticks)/64 cycles per line))
8   NOT USED
9 ✔️ Set SampleOffset                    9xx : offset (23 -> 2300)                      CH sample pos. (applied once)
A ✔️ VolumeSlide                         Axy : x-upspeed, y-downspeed                   PR VS (applied once per tick)
B ✖ Position Jump                       Bxx : songposition                             GL position (GLOBAL! applied once)
C ✔️ Set Volume                          Cxx : volume, 00-40                            PR VS (applied once)
D ✔️ Pattern Break (FIXME)               Dxx : break position in next patt              GL position (GLOBAL! applied once)
E   E-Commands                          Exy : see below...
F ✔️ Set Speed                           Fxx : speed (00-1F) / tempo (20-FF)            GL speed (GLOBAL! applied once)
----------------------------------------------------------------------------
E0✖ Set Filter                          E0x : 0-filter on, 1-filter off                IGNORE
E1✔️?FineSlide Up                        E1x : value                                    PS VR (halfnotes (?), applied once)
E2✔️?FineSlide Down                      E2x : value                                    PS VR (halfnotes (?), applied once)
E3✔️?Glissando Control                   E3x : 0-off, 1-on (use with tonep.)            PF (->#3) 1 slide one halfnote at a time
E4✔️?Set Vibrato Waveform                E4x : 0-sine, 1-ramp down, 2-square            PF (->#4)
E5✔️?Set Finetune                        E5x : set finetune                             IS instrument setting
E6✔️?Set Loop/Jump to Loop               E6x : set/jump to loop, play x times           GL position (GLOBAL! applied once)
E7✔️?Set Tremolo Waveform                E7x : 0-sine, 1-ramp down. 2-square            VF (->#7)
E8  NOT USED / Set Panning
E9✔️?Retrig Note                         E9x : retrig from note + x vblanks             CH sample pos. (applied once every x ticks)
EA✔️?Fine VolumeSlide Up                 EAx : add x to volume                          PR VS (applied once)
EB✔️?Fine VolumeSlide Down               EBx : subtract x from volume                   PR VS (applied once)
EC✔️?NoteCut                             ECx : cut from note + x vblanks                CH channel active (applied once after x ticks)
ED✔️?NoteDelay                           EDx : delay note x vblanks                     CH sample pos. (applied once after x ticks)
EE✔️?PatternDelay                        EEx : delay pattern x notes                    GL delay (applied once after x ticks) -> "additional lines" counter
EF✖ Invert Loop                         EFx : speed                                    IGNORE ???

## Key

PS - pitch set (incl. slide/vibrato)
PK - pitch keep (only specified effect, i.e. portamento or vibrato)
PR - pitch reset (reset all pitch effects)
PF - pitch flag (option affecting pitch calculation)

VS - volume set (incl. slide/tremolo)
VR - volume reset (reset all volume effects)
VF - volume flag (option affecting volume calculation)

CH - affects channel (e.g. sample position/active)
GL - global (play speed/position)

## Stats

Implemented Base Effects: 11/14; Tested: ??/14
Implemented Ext. Effects: 13/13; Tested: 00/13

Pitch effects: 9 (7 pitch effects, 2 settings)
Volume effects: 9 (8 volume effects, 1 settings)
Position/timing commands: 10

# Misc Notes

"If you need to implement volume artificially, just multiply by the volume and shift right 6 times." 
(TODO: multiply by volume and use 16 bit, saves right-shifting)
