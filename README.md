# vidcap

Extract screenshots and GIFs from video files via the command line.

## Requirements

**Windows**
```
winget install Gyan.FFmpeg
```

**Mac**
```
brew install ffmpeg
```

## Usage

### Windows

Screenshots every N seconds
```
./vidcap.exe screenshots --every 10s input.mp4
```

N screenshots evenly spaced across the video
```
./vidcap.exe screenshots --count 10 input.mp4
```

GIF clips of X duration, every N seconds
```
./vidcap.exe gifs --duration 3s --every 30s input.mp4
```

N GIF clips of X duration, evenly spaced across the video
```
./vidcap.exe gifs --duration 3s --count 5 input.mp4
```

### Mac

Screenshots every N seconds
```
./vidcap screenshots --every 10s input.mp4
```

N screenshots evenly spaced across the video
```
./vidcap screenshots --count 10 input.mp4
```

GIF clips of X duration, every N seconds
```
./vidcap gifs --duration 3s --every 30s input.mp4
```

N GIF clips of X duration, evenly spaced across the video
```
./vidcap gifs --duration 3s --count 5 input.mp4
```

Output is saved next to the input file:
- `input-screenshots/001.png`, `002.png`, ...
- `input-gifs/001.gif`, `002.gif`, ...

## File Structure

```
├── main.go
├── cmd/
│   ├── root.go          # CLI entry point, ffmpeg presence check
│   ├── screenshots.go   # screenshots subcommand
│   └── gifs.go          # gifs subcommand
└── internal/
    ├── ffmpeg/
    │   ├── probe.go     # get video duration via ffprobe
    │   └── extract.go   # extract frames and GIF clips via ffmpeg
    └── timestamps/
        ├── calc.go      # calculate timestamps (--every / --count)
        └── calc_test.go
```

## Build

**Windows**
```
go build -o vidcap.exe .
```

**Mac**
```
go build -o vidcap .
```
