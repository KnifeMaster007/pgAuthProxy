package utils

import (
	"bytes"
	"io"
	"sync"
	"time"
)

const (
	bufferedWriterTickerInterval = 10 * time.Millisecond
)

type BufferedWriter struct {
	mutex          sync.Mutex
	buffer         *bytes.Buffer
	writer         io.Writer
	flushTicker    *time.Ticker
	flushInterval  time.Duration
	nextFlushDelay time.Duration
	isOpen         bool
}

func NewBufferedWriter(cap int, writer io.Writer, flushInterval time.Duration) *BufferedWriter {
	ret := &BufferedWriter{
		writer:         writer,
		buffer:         bytes.NewBuffer(make([]byte, cap)),
		flushInterval:  flushInterval,
		nextFlushDelay: flushInterval,
		flushTicker:    time.NewTicker(bufferedWriterTickerInterval),
		isOpen:         true,
	}
	ret.buffer.Reset()
	go func() {
		for ret.isOpen {
			select {
			case <-ret.flushTicker.C:
				ret.mutex.Lock()
				ret.nextFlushDelay -= bufferedWriterTickerInterval
				var err error
				if ret.nextFlushDelay <= 0 {
					_, err = ret._flush()
				}
				ret.mutex.Unlock()
				if err != nil {
					ret.Close()
				}
			}
		}
	}()
	return ret
}

func (bw *BufferedWriter) Write(p []byte) (int, error) {
	var (
		n   int
		err error
	)
	if !bw.isOpen {
		return 0, io.EOF
	}
	total := 0
	length := len(p)
	rb := bytes.NewBuffer(p)
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
	if err != nil {
		bw.Close()
	} else {
		bw.nextFlushDelay = bw.flushInterval
	}
	bw.mutex.Unlock()
	return total, err
}

func (bw *BufferedWriter) _flush() (int, error) {
	if bw.buffer.Len() > 0 {
		n, err := bw.writer.Write(bw.buffer.Next(bw.buffer.Len()))
		if err != nil {
			bw.Close()
		}
		return n, err
	}
	return 0, nil
}

func (bw *BufferedWriter) Flush() (int, error) {
	bw.mutex.Lock()
	n, err := bw._flush()
	bw.mutex.Unlock()
	return n, err
}

func (bw *BufferedWriter) Close() {
	bw.mutex.Lock()
	bw.flushTicker.Stop()
	bw.isOpen = false
	_, _ = bw._flush()
	bw.mutex.Unlock()
}
