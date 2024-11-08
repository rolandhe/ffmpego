package main

import (
	"fmt"
	"github.com/rolandhe/ffmpego"
	"log"
	"os"
	"sync"
	"time"
)

type parent struct {
	name string
}

func (p * parent)show()  {
	fmt.Println(p.name)
}

type child struct {
	*parent
}

func (c *child)show ()  {
	c.parent.show()
	fmt.Println("i am child")
}


func main() {

	c := &child{
		parent: &parent{
			name:"parent",
		},
	}
	c.show()

	//ctx := ffmpego.NewRunPooledContext(3, 10)
	////convertFileInPool(ctx)
	//
	//wg := &sync.WaitGroup{}
	//wg.Add(3)
	//
	//go convertBytesInPool(wg, ctx)
	//go convertFileInPool(wg, ctx)
	//go durationInPool(wg, ctx)
	//
	//wg.Wait()
	//
	//ctx.Shutdown()
	//ctx.WaitShutdown()
}

func convertFileInPool(wg *sync.WaitGroup, ctx *ffmpego.RunPooledContext) {
	defer wg.Done()
	task, err := ctx.AddFileTask("trace_id_10022", "ffmpeg -i test/s95.mp3 test/out_file.aac")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = task.WaitResult()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func durationInPool(wg *sync.WaitGroup, ctx *ffmpego.RunPooledContext) {
	defer wg.Done()
	task, err := ctx.GetFileDuration("trace_id_100266", "test/vvx.m4a")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = task.WaitResult()
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("durationInPool:%d,%v", task.DurationVal, task.Err)
}

func convertBytesInPool(wg *sync.WaitGroup, ctx *ffmpego.RunPooledContext) {
	defer wg.Done()
	format := "ffmpeg -f mp3 -i %s -f adts -acodec aac %s"
	inputData, err := os.ReadFile("test/s95.mp3")
	if err != nil {
		fmt.Println(err)
		return
	}
	task, err := ctx.AddBytesTask("trace_id_9999", format, inputData)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = task.WaitResult()
	if err != nil {
		fmt.Println("wait result", err)
		return
	}
	outputData := task.OutputData
	err = os.WriteFile("test/out_byte.aac", outputData, os.ModePerm)
	fmt.Println("write aac", err)
}

func pf() {
	iFmt := "ffmpeg -i %s -vn -f s16le -ac 1 -ar 16000 %s"

	w, f, err := ffmpego.StartInputStream("trace_id_1999", iFmt, "/home/xiufeng/github/ffmpego/main/ion.pcm")
	if err != nil {
		return
	}

	inputData, _ := os.ReadFile("/home/xiufeng/github/ffmpego/main/ion.ogg")
	for {
		n, err := w.Write(inputData)
		if err != nil {
			return
		}
		if n == len(inputData) {
			break
		}
		inputData = inputData[n:]
	}

	w.Close()
	for {

		ok, err := f.TryGetTimeout(time.Second * 2)
		if err != nil {
			fmt.Println(err)
			return
		}
		if ok {
			break
		}
	}
	fmt.Println("okk...")
}

func pp() {
	iFmt := "ffmpeg -i %s -vn -f s16le -ac 1 -ar 16000 %s"
	fs, _ := os.Create("/home/xiufeng/github/ffmpego/main/ion.pcm")
	defer fs.Close()
	proc := ffmpego.NewFileStreamOutputProcessor(fs)
	w, f, err := ffmpego.StartStream("trace_id_1999", iFmt, proc)
	if err != nil {
		return
	}

	inputData, _ := os.ReadFile("/home/xiufeng/github/ffmpego/main/ion.ogg")
	for {
		n, err := w.Write(inputData)
		if err != nil {
			return
		}
		if n == len(inputData) {
			break
		}
		inputData = inputData[n:]
	}

	w.Close()
	for {

		ok, err := f.TryGetTimeout(time.Second * 2)
		if err != nil {
			fmt.Println(err)
			return
		}
		if ok {
			break
		}
	}
	fmt.Println("okk...")
}
