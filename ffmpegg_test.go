package ffmpego

import (
	"fmt"
	"testing"
)

func TestRunQuickDuration(t *testing.T) {
	dur, err := runQuickDuration("traceId-199191", "ffmpeg -i /Users/hexiufeng/github/ffmpego/test/s95.mp3")
	fmt.Println(dur, err)
}
