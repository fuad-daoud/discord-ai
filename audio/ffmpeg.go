package audio

import (
	"golang.org/x/net/context"
	"io"
	"os/exec"
)

func FFMPEG(in io.Reader) (io.ReadCloser, error) {
	cmd := exec.CommandContext(context.Background(), "ffmpeg",
		"-i",
		"pipe:0",
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	)
	cmd.Stdin = in

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	return pipe, err
}
