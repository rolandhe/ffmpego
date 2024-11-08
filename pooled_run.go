package ffmpego

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

var TimeoutErr = errors.New("timeout")
var ExceedErr = errors.New("exceed worker size")

type Task struct {
	TraceId       string
	Cmd           string
	InputData     []byte
	OutputData    []byte
	Err           error
	wait          chan struct{}
	quickDuration bool
	DurationVal   int64
}

func (task *Task) finish() {
	close(task.wait)
}
func (task *Task) WaitResult() error {
	<-task.wait
	return task.Err
}

func (task *Task) WaitResultWithTimeout(timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-task.wait:
		return task.Err
	case <-timer.C:
		return TimeoutErr
	}
}

func NewRunPooledContext(workerSize int, backTaskSize int) *RunPooledContext {
	ctx := &RunPooledContext{
		q:   make(chan *Task, backTaskSize),
		end: make(chan struct{}),
		threadEnding: &countDown{
			trigger: make(chan struct{}),
		},
	}
	ctx.threadEnding.Add(int32(workerSize))
	buildWorkers(workerSize, ctx)
	return ctx
}

type RunPooledContext struct {
	q            chan *Task
	end          chan struct{}
	threadEnding *countDown
}

type countDown struct {
	atomic.Int32
	trigger chan struct{}
}

func (cd *countDown) Done() {
	newValue := cd.Add(-1)
	if newValue == 0 {
		close(cd.trigger)
	}
}

func (ctx *RunPooledContext) AddFileTask(traceId string, cmd string) (*Task, error) {
	task := &Task{
		TraceId: traceId,
		Cmd:     cmd,
		wait:    make(chan struct{}),
	}
	return addTask(task, ctx.q)
}

// GetFileDuration  return nano
func (ctx *RunPooledContext) GetFileDuration(traceId string, filePath string) (*Task, error) {
	task := &Task{
		TraceId:       traceId,
		Cmd:           fmt.Sprintf("ffmpeg -i %s", filePath),
		quickDuration: true,
		wait:          make(chan struct{}),
	}
	return addTask(task, ctx.q)
}

// AddBytesTask format:ffmpeg [opts]. -i %s [opts]. %s,
// e.g. ffmpeg -f s16l4 -ac 1 -ar 16000 %s -f mp3 %s
func (ctx *RunPooledContext) AddBytesTask(traceId string, format string, inputData []byte) (*Task, error) {
	task := &Task{
		TraceId:   traceId,
		Cmd:       format,
		InputData: inputData,
		wait:      make(chan struct{}),
	}
	return addTask(task, ctx.q)
}

func (ctx *RunPooledContext) Shutdown() {
	close(ctx.end)
}

func (ctx *RunPooledContext) WaitShutdown() {
	<-ctx.threadEnding.trigger
}
func (ctx *RunPooledContext) WaitShutdownTimeout(duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.threadEnding.trigger:
		return nil
	case <-timer.C:
		return TimeoutErr
	}
}

func buildWorkers(workerSize int, ctx *RunPooledContext) {
	for i := 0; i < workerSize; i++ {
		go func(tid int) {
			runtime.LockOSThread()
			d := time.Second * 10
			timer := time.NewTimer(d)
			defer timer.Stop()
			for {
				timer.Reset(d)
				select {
				case <-ctx.end:
					ctx.threadEnding.Done()
					log.Printf("i am worker thread %d,end\n", tid)
					return
				case task := <-ctx.q:
					processTask(tid, task)
				case <-timer.C:
					log.Printf("i am worker thread %d,no task\n", tid)
				}
			}
		}(i)
	}
}

func processTask(tid int, task *Task) {
	if task.quickDuration {
		task.DurationVal, task.Err = RunQuickDuration(task.TraceId, task.Cmd)
		log.Printf("worker thread id:%d,err:%v\n", tid, task.Err)
		task.finish()
		return
	}
	if task.InputData == nil {
		task.Err = RunCmdBasingFile(task.TraceId, task.Cmd)
	} else {
		task.OutputData, task.Err = RunDataProtoUseOutPipe(task.TraceId, task.Cmd, task.InputData)
	}
	log.Printf("worker thread id:%d,err:%v\n", tid, task.Err)
	task.finish()
}

func addTask(task *Task, q chan *Task) (*Task, error) {
	select {
	case q <- task:
		return task, nil
	default:
		return nil, ExceedErr
	}
}
