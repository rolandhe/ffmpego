package ffmpego

import "os"

func NewFileStreamOutputProcessor(fs *os.File) StreamOutputProcessor {
	return &fileStreamOutputProcessor{
		fs: fs,
	}
}

type fileStreamOutputProcessor struct {
	fs *os.File
}

func (fp *fileStreamOutputProcessor) Process(data []byte) error {
	_, err := fp.fs.Write(data)
	return err
}
