package ffmpego

// #cgo CFLAGS:  -I/usr/local/include
// #cgo LDFLAGS: -L/usr/local/lib -lrun_ffmpeg
// #include <run_ffmpeg.h>
import "C"
import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"
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

func runQuickDuration(traceId string, cmd string) (int64, error) {
	var duration C.int64_t
	ret := int32(C.quick_duration(C.CString(traceId), C.CString(cmd), &duration))
	if ret < 0 {
		return 0, errors.New("run test failed")
	}
	return int64(duration), nil
}

type fakeString struct {
	Data *C.char
	Len  int
}

type Future struct {
	ch  chan int32
	err error
}

func (f *Future) TryGet() (bool, error) {
	select {
	case v := <-f.ch:
		if f.err != nil {
			return true, f.err
		}
		if v != 0 {
			return true, errors.New(fmt.Sprintf("return err:%d", v))
		}
		return true, nil
	default:
		return false, nil
	}
}

type StreamOutputProcessor interface {
	Process(data []byte) error
}

// StartStream runCmdBasingBytes format:ffmpeg [opts]. -i %s [opts]. %s,
// e.g. ffmpeg -f s16l4 -ac 1 -ar 16000 -i %s -f mp3 %s
// 通过 io.WriteCloser 可以持续不断的喂数据给ffmpeg
func StartStream(traceId string, format string, process StreamOutputProcessor) (io.WriteCloser, *Future, error) {
	inReader, inWriter, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		inReader.Close()
		inWriter.Close()
		return nil, nil, err
	}

	in := fmt.Sprintf("pipe:%d", uint64(inReader.Fd()))
	out := fmt.Sprintf("pipe:%d", uint64(outWriter.Fd()))

	cmd := fmt.Sprintf(format, in, out)

	resultChan := make(chan int32, 1)
	future := &Future{
		ch: resultChan,
	}

	ppWriter := &pipeWriter{
		writer: inWriter,
	}

	waiter := make(chan struct{})
	go func(r *os.File) {
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if err == io.EOF {
				close(waiter)
				break
			}
			if err != nil {
				future.err = err
				ppWriter.Close()
				close(waiter)
				break
			}
			if n > 0 {
				if err = process.Process(buf[:n]); err != nil {
					future.err = err
					ppWriter.Close()
					close(waiter)
					break
				}
			}
		}
	}(outReader)

	go func() {
		ret := int32(C.run_ffmpeg_cmd(C.CString(traceId), C.CString(cmd)))
		outWriter.Close()
		<-waiter
		outReader.Close()
		inReader.Close()
		resultChan <- ret
	}()

	return ppWriter, future, nil
}

type pipeWriter struct {
	closed atomic.Bool
	writer *os.File
}

func (pw *pipeWriter) Write(p []byte) (n int, err error) {
	if pw.closed.Load() {
		return 0, errors.New("has closed")
	}
	return pw.writer.Write(p)
}

func (pw *pipeWriter) Close() error {
	if pw.closed.Load() {
		return nil
	}
	if !pw.closed.CompareAndSwap(false, true) {
		return nil
	}
	return pw.Close()
}

// runCmdBasingBytes format:ffmpeg [opts]. -i %s [opts]. %s,
// e.g. ffmpeg -f s16l4 -ac 1 -ar 16000 -i %s -f mp3 %s
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
