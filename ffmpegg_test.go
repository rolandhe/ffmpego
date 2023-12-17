package ffmpego

import (
	"fmt"
	"os"
	"testing"
)

func TestRunQuickDuration(t *testing.T) {
	dur, err := RunQuickDuration("traceId-199191", "ffmpeg -i /Users/hexiufeng/github/ffmpego/test/s95.mp3")
	fmt.Println(dur, err)
}

func TestRunCmdMemInOut(t *testing.T) {
	data, _ := os.ReadFile("/home/xiufeng/github/ffmpego/recorder.webm")
	cmdFmt := "ffmpeg -f webm -i %s -vn -f s16le %s.pcm"
	out, err := RunCmdMemInOut("traceId-199191", data, cmdFmt)

	fmt.Println(err)
	if err != nil {
		return
	}
	if len(out) > 0 {
		os.WriteFile("/home/xiufeng/github/ffmpego/rs16.pcm", out, os.ModePerm)
	}
}
