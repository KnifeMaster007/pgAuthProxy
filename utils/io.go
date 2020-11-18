package utils

import (
	"bytes"
	"io"
	"sync"
)

type BufferedWriter struct {
	mutex  sync.Mutex
	buffer *bytes.Buffer
	writer io.Writer
}

func NewBufferedWriter(cap int, writer io.Writer) *BufferedWriter {
	ret := &BufferedWriter{
		writer: writer,
		buffer: bytes.NewBuffer(make([]byte, cap)),
	}
	ret.buffer.Reset()
	return ret
}

func (bw *BufferedWriter) Write(p []byte) (int, error) {
	var (
		n   int
		err error
	)
	total := 0
	length := len(p)
	rb := bytes.NewBuffer(p)
	//rb.Reset()
	bw.mutex.Lock()
	for total < length {
		n, err = bw.buffer.Write(rb.Next(bw.buffer.Cap() - bw.buffer.Len()))
		if err != nil {
			break
		}
		if bw.buffer.Len() == bw.buffer.Cap() {
			_, err = bw._flush()
		}
		if err != nil {
			break
		}
		total += n
	}
	bw.mutex.Unlock()
	return total, err
}

func (bw *BufferedWriter) _flush() (int, error) {
	if bw.buffer.Len() > 0 {
		return bw.writer.Write(bw.buffer.Next(bw.buffer.Len()))
	}
	return 0, nil
}

func (bw *BufferedWriter) Flush() (int, error) {
	bw.mutex.Lock()
	n, err := bw._flush()
	bw.mutex.Unlock()
	return n, err
}
