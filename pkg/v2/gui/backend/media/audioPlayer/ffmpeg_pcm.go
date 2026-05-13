package audioplayer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/faiface/beep"
)

const (
	ffmpegPCMRate     = beep.SampleRate(44100)
	ffmpegPCMChannels = 2
	ffmpegPCMBitDepth = 16
)

type pcmSeekCloser struct {
	streamer beep.StreamSeeker
	closeFn  func() error
}

func (p *pcmSeekCloser) Stream(samples [][2]float64) (n int, ok bool) {
	return p.streamer.Stream(samples)
}

func (p *pcmSeekCloser) Err() error {
	return p.streamer.Err()
}

func (p *pcmSeekCloser) Len() int {
	return p.streamer.Len()
}

func (p *pcmSeekCloser) Position() int {
	return p.streamer.Position()
}

func (p *pcmSeekCloser) Seek(pos int) error {
	return p.streamer.Seek(pos)
}

func (p *pcmSeekCloser) Close() error {
	if p.closeFn != nil {
		return p.closeFn()
	}
	return nil
}

func decodeWithFFmpegPCM(ctx context.Context, path string) (beep.StreamSeekCloser, beep.Format, error) {
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-i", path,

		// Normalize output for Beep.
		"-vn",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ar", "44100",
		"-ac", "2",
		"pipe:1",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, beep.Format{}, fmt.Errorf("start ffmpeg: %w", err)
	}

	data, readErr := io.ReadAll(stdout)
	waitErr := cmd.Wait()

	if readErr != nil {
		return nil, beep.Format{}, fmt.Errorf("read ffmpeg pcm: %w", readErr)
	}

	if waitErr != nil {
		msg := stderr.String()
		if msg == "" {
			msg = waitErr.Error()
		}
		return nil, beep.Format{}, fmt.Errorf("ffmpeg decode: %s", msg)
	}

	if len(data) == 0 {
		return nil, beep.Format{}, fmt.Errorf("ffmpeg returned empty pcm")
	}

	const bytesPerFrame = ffmpegPCMChannels * 2 // stereo * int16
	if rem := len(data) % bytesPerFrame; rem != 0 {
		data = data[:len(data)-rem]
	}

	sampleCount := len(data) / bytesPerFrame
	samples := make([][2]float64, sampleCount)

	for i := 0; i < sampleCount; i++ {
		offset := i * bytesPerFrame

		left := int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
		right := int16(binary.LittleEndian.Uint16(data[offset+2 : offset+4]))

		samples[i][0] = float64(left) / 32768.0
		samples[i][1] = float64(right) / 32768.0
	}

	format := beep.Format{
		SampleRate:  ffmpegPCMRate,
		NumChannels: ffmpegPCMChannels,
		Precision:   ffmpegPCMBitDepth / 8,
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(&pcmMemoryStreamer{
		samples: samples,
	})

	return &pcmSeekCloser{
		streamer: buffer.Streamer(0, buffer.Len()),
	}, format, nil
}

type pcmMemoryStreamer struct {
	samples [][2]float64
	pos     int
	err     error
}

func (s *pcmMemoryStreamer) Stream(out [][2]float64) (n int, ok bool) {
	if s.pos >= len(s.samples) {
		return 0, false
	}

	n = copy(out, s.samples[s.pos:])
	s.pos += n

	return n, s.pos < len(s.samples)
}

func (s *pcmMemoryStreamer) Err() error {
	return s.err
}

func (s *pcmMemoryStreamer) Len() int {
	return len(s.samples)
}

func (s *pcmMemoryStreamer) Position() int {
	return s.pos
}

func (s *pcmMemoryStreamer) Seek(pos int) error {
	if pos < 0 {
		pos = 0
	}
	if pos > len(s.samples) {
		pos = len(s.samples)
	}

	s.pos = pos
	return nil
}

func (s *pcmMemoryStreamer) Close() error {
	s.samples = nil
	s.pos = 0
	s.err = nil
	return nil
}

func pcmDuration(format beep.Format, streamer beep.StreamSeekCloser) time.Duration {
	if streamer == nil || streamer.Len() <= 0 {
		return 0
	}
	return format.SampleRate.D(streamer.Len())
}
