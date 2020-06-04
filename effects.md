# Effects implementation status

0 ✖ Normal play or Arpeggio             0xy : x-first halfnote add, y-second           pitch (w. note assignment, applied 2/3 times per line)
1 ✔️ Slide Up                            1xx : upspeed                                  pitch (applied once per tick)
2 ✔️ Slide Down                          2xx : downspeed                                pitch (applied once per tick)
3 ✖ Tone Portamento                     3xx : up/down speed                            pitch (applied once per tick)
4 ✖ Vibrato                             4xy : x-speed,   y-depth                       pitch (w. waveform, applied continuously: amplitude y/16 halfnotes, (x*ticks)/64 cycles per line)   _flippin
5 ✖ Tone Portamento + Volume Slide      5xy : x-upspeed, y-downspeed                   volume (cont. #3, applied once per tick)
6 ✖ Vibrato + Volume Slide              6xy : x-upspeed, y-downspeed                   volume (cont. #4, applied once per tick)
7 ✖ Tremolo                             7xy : x-speed,   y-depth                       volume (w. waveform, applied continuously: amplitude y ???, (x*ticks)/64 cycles per line))
8   NOT USED
9 ✖ Set SampleOffset                    9xx : offset (23 -> 2300)                      sample pos. (applied once)
A ✔️ VolumeSlide                         Axy : x-upspeed, y-downspeed                   volume (applied once per tick)
B ✖ Position Jump                       Bxx : songposition                             position (GLOBAL! applied once)
C ✔️ Set Volume                          Cxx : volume, 00-40                            volume (applied once)
D ✔️ Pattern Break (FIXME)               Dxx : break position in next patt              position (GLOBAL! applied once)
E   E-Commands                          Exy : see below...
F ✔️ Set Speed                           Fxx : speed (00-1F) / tempo (20-FF)            speed (GLOBAL! applied once)
----------------------------------------------------------------------------
E0✖ Set Filter                          E0x : 0-filter on, 1-filter off                IGNORE
E1✖ FineSlide Up                        E1x : value                                    pitch (halfnotes (?), applied once)
E2✖ FineSlide Down                      E2x : value                                    pitch (halfnotes (?), applied once)
E3✖ Glissando Control                   E3x : 0-off, 1-on (use with tonep.)            flag (->#3) 1 slide one halfnote at a time
E4✖ Set Vibrato Waveform                E4x : 0-sine, 1-ramp down, 2-square            flag (->#4)
E5✖ Set Loop                            E5x : set loop point / set finetune?!          instrument setting                 
E6✖ Jump to Loop                        E6x : set/jump to loop, play x times           position (GLOBAL! applied once)
E7✖ Set Tremolo Waveform                E7x : 0-sine, 1-ramp down. 2-square            flag (->#7)
E8  NOT USED / Set Panning
E9✖ Retrig Note                         E9x : retrig from note + x vblanks             sample pos. (applied once every x ticks)
EA✖ Fine VolumeSlide Up                 EAx : add x to volume                          volume (applied once)
EB✖ Fine VolumeSlide Down               EBx : subtract x from volume                   volume (applied once)
EC✖ NoteCut                             ECx : cut from note + x vblanks                volume set to 0 (applied once after x ticks)
ED✖ NoteDelay                           EDx : delay note x vblanks                     sample pos. (applied once after x ticks)
EE✖ PatternDelay                        EEx : delay pattern x notes                    position (applied once after x ticks) 
EF✖ Invert Loop                         EFx : speed                                    ???

Base Effects: 06/14
Ext. Effects: 00/15

"If you need to implement volume artificially, just multiply by the volume and shift right 6 times." 
(TODO: multiply by volume and use 16 bit, saves right-shifting)


ISSUES:

go-modplayer BeatWave.mod 
BeatWave.mod
panic: runtime error: index out of range [255] with length 32

goroutine 1 [running]:
main.ReadNote(0xc0000b4c40, 0x4, 0x4f10e, 0xc0000a4000, 0xc0000a4038, 0x0, 0x0, 0x0)
        /home/rnieren/go/src/github.com/nieware/go-modplayer/file.go:317 +0xf1
main.ReadModFile(0x7fffe351f0ba, 0xc, 0xc0000a1248, 0x1, 0x1, 0xd, 0x0, 0x0, 0x0, 0x0, ...)
        /home/rnieren/go/src/github.com/nieware/go-modplayer/file.go:246 +0x714
main.main()
        /home/rnieren/go/src/github.com/nieware/go-modplayer/main.go:21 +0x20f

