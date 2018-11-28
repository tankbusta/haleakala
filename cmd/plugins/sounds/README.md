
## Conversion

MP3 to DCA:

    ffmpeg -i birthcontrol.mp3 -f s16le -ar 48000 -ac 2 pipe:1 | dca > birthcontrol.dca
