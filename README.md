# ffmpego

golang调用[run_ffmpeg库](https://github.com/rolandhe/run-ffmpeg),支持golang在goroutine内调用ffmpeg的能力。

ffmpego 内置线程池，可以并行执行ffmpeg指令，你可以指定线程的多少，也可以指定待执行任务池的大小。

# 前置条件

先安装[run_ffmpeg库](https://github.com/rolandhe/run-ffmpeg)库

# 使用

## 初始化上下文

```
ctx := ffmpego.NewRunPooledContext(3, 10)
```

第一个参数指定工作线程数，第二参数指定待执行的任务数，如果待执行队列中任务超过了这个值，则提交任务会返回 ffmpego.ExceedErr。工作线程，这里是线程不是goroutine, 因为其内部会调用run_ffmpeg库，绑定线程会更安全。

每一个上下文会持有自己的线程池及任务队列，你可以在应用内初始化多个上下文。

## 执行基于文件的转换任务

```
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
```

task.WaitResult()会等待任务执行完成，并返回运行错误。

task.WaitResultWithTimeout(timeout time.Duration), 支持等待超时，如果超时则返回 ffmpego.TimoutErr


## 执行基于数组的转换任务

执行非文件的转换需要传入如下的指令，

```
format := "ffmpeg -f mp3 -i %s -f adts %s"
```

-f mp3 指的是输入数据的格式
-f adts, 指的是输出数据的格式，adts指的是aac

其中两个 %s 保持不变，在执行函数的内部会进行数据替换。

调用task, err := ctx.AddBytesTask("trace_id_9999", format, inputData)函数执行。

```
    format := "ffmpeg -f mp3 -i %s -f adts %s"
    // 从文件中读取需要转换的数据
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
	// 获取输出数据，然后写入文件
	outputData := task.OutputData
	err = os.WriteFile("test/out_byte.aac", outputData, os.ModePerm)
	fmt.Println("write aac", err)
```
