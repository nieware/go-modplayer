# MOD file format docs

(in decreasing order of usefulness)
- [mod-form.txt](http://www.textfiles.com/programming/FORMATS/mod-form.txt)
- [mod-spec.txt](https://eblong.com/zarf/blorb/mod-spec.txt)
- [mod.txt v1.0/1.1](http://lclevy.free.fr/mo3/mod.txt) [or here](https://raw.githubusercontent.com/cmatsuoka/libxmp/c5df8ece33c15ad353e809e52add41a957ab74d0/docs/formats/Modfil11.txt)
- [German Wikipedia article](https://de.wikipedia.org/wiki/Tracker_(Musik))
- [MilkyTracker doc](https://milkytracker.titandemo.org/docs/MilkyTracker.html#effects) -> for an alternative tracker, but good documentation of MOD effects
- [modulesg.txt](http://www.textfiles.com/programming/FORMATS/modulesg.txt)

# Audio processing

- [interpolation](https://stackoverflow.com/questions/1125666/how-do-you-do-bicubic-or-other-non-linear-interpolation-of-re-sampled-audio-da)

# Misc notes

## Show info for all files in current directory

find . -maxdepth 1 -type f -iname "*.mod" -exec go-modplayer -info {} \; > ./info.txt