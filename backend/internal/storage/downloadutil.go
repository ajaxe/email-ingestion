package storage

import (
	"context"
	"fmt"
	"io"
)

// pipeWriterAt wraps an io.PipeWriter to satisfy the io.WriterAt interface.
// It tracks expected offsets to ensure sequential streaming.
type pipeWriterAt struct {
	w       io.Writer
	written int64
}

func (pwa *pipeWriterAt) WriteAt(p []byte, off int64) (int, error) {
	if off != pwa.written {
		return 0, fmt.Errorf("out-of-order write at offset %d, expected %d", off, pwa.written)
	}
	n, err := pwa.w.Write(p)
	pwa.written += int64(n)
	return n, err
}

// pipeReadCloser wraps an io.PipeReader to ensure the background
// context is canceled when the consumer closes the reader early.
type pipeReadCloser struct {
	*io.PipeReader
	cancel context.CancelFunc
}

func (prc *pipeReadCloser) Close() error {
	prc.cancel() // Stop the background S3 download immediately
	return prc.PipeReader.Close()
}
