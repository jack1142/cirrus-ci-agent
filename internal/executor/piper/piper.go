package piper

import (
	"context"
	"errors"
	"io"
	"os"
)

type Piper struct {
	r, w    *os.File
	errChan chan error
}

func New(output io.Writer) (*Piper, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	piper := &Piper{
		r:       r,
		w:       w,
		errChan: make(chan error),
	}

	go func() {
		_, err := io.Copy(output, r)
		piper.errChan <- err
	}()

	return piper, nil
}

func (piper *Piper) Input() *os.File {
	return piper.w
}

func (piper *Piper) Close(ctx context.Context) (result error) {
	// Close our writing end (if not closed yet)
	if err := piper.w.Close(); err != nil && !errors.Is(err, os.ErrClosed) && result == nil {
		result = err
	}
	// Terminate the Goroutine started in New()
	err := piper.r.Close()
	if err != nil && result == nil {
		result = err
	}

	// Wait for the Goroutine started in New(): it will reach EOF once
	// all the copies of the writing end file descriptor are closed

	select {
	case <-ctx.Done():
		if result == nil {
			result = context.Canceled
		}
	case err := <-piper.errChan:
		if err != nil && !errors.Is(err, os.ErrClosed) {
			result = err
		}
	}

	return result
}
