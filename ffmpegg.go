package ffmpego

// #cgo LDFLAGS: -L/usr/local/lib -lrun_ffmpeg
// #include <run_ffmpeg.h>
import "C"
import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"unsafe"
)

func init() {
	C.init_ffmpeg()
}

func runCmdBasingFile(traceId string, cmd string) error {
	ret := int32(C.run_ffmpeg_cmd(C.CString(traceId), C.CString(cmd)))

	if ret < 0 {
		return errors.New("run test failed")
	}
	return nil
}

type fakeString struct {
	Data *C.char
	Len  int
}

// runCmdBasingBytes format:ffmpeg [opts]. -i %s [opts]. %s,
// e.g. ffmpeg -f s16l4 -ac 1 -ar 16000 %s -f mp3 %s
func runCmdBasingBytes(traceId string, format string, data []byte) ([]byte, error) {

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	b64 := "data:content/type;base64," + base64.StdEncoding.EncodeToString(data)
	out := fmt.Sprintf("pipe:%d", uint64(w.Fd()))

	cmd := fmt.Sprintf(format, b64, out)

	holder := &struct {
		data []byte
		err  error
	}{}

	wait := make(chan struct{})
	go func(f *os.File) {
		buf := make([]byte, 4096)
		holder.data = make([]byte, 0, 16*1024)
		for {
			n, err := f.Read(buf)
			if n <= 0 || err == io.EOF {
				close(wait)
				break
			}
			if err != nil {
				holder.err = err
				close(wait)
				break
			}
			holder.data = append(holder.data, buf[:n]...)
		}
	}(r)

	fs := (*fakeString)(unsafe.Pointer(&cmd))

	ret := int32(C.run_ffmpeg_cmd(C.CString(traceId), fs.Data))
	w.Close()
	<-wait
	r.Close()

	if ret < 0 {
		return nil, errors.New("run test failed")
	}
	return holder.data, holder.err
}
