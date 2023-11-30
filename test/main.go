package main

import (
	"fmt"
	"github.com/rolandhe/ffmpego"
	"log"
	"os"
	"sync"
)

func main() {
	ctx := ffmpego.NewRunPooledContext(3, 10)
	//convertFileInPool(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	//go convertBytesInPool(wg, ctx)
	//go convertFileInPool(wg, ctx)
	go durationInPool(wg, ctx)

	wg.Wait()

	ctx.Shutdown()
	ctx.WaitShutdown()
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
