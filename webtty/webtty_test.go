package webtty

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"sync"
	"testing"
)

type pipePair struct {
	*io.PipeReader
	*io.PipeWriter
}

type pipeSlave struct {
	*io.PipeReader
	*io.PipeWriter
}

func (slave pipeSlave) WindowTitleVariables() map[string]interface{} {
	return nil
}

func (slave pipeSlave) ResizeTerminal(columns int, rows int) error {
	return nil
}

func TestWriteFromPTY(t *testing.T) {
	connInPipeReader, connInPipeWriter := io.Pipe() // in to conn
	connOutPipeReader, _ := io.Pipe()               // out from conn
	slaveOutPipeReader, slaveOutPipeWriter := io.Pipe()
	slaveInPipeReader, slaveInPipeWriter := io.Pipe()
	defer slaveInPipeReader.Close()

	conn := pipePair{
		connOutPipeReader,
		connInPipeWriter,
	}
	slave := pipeSlave{
		slaveOutPipeReader,
		slaveInPipeWriter,
	}
	dt, err := New(conn, slave)
	if err != nil {
		t.Fatalf("Unexpected error from New(): %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		err := dt.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Fatalf("Unexpected error from Run(): %s", err)
		}
	}()

	buf := make([]byte, 1024)
	n, err := connInPipeReader.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected error from Read(): %s", err)
	}
	if !bytes.Equal(buf[:n], []byte{SetWindowTitle}) {
		t.Fatalf("Unexpected message received: `%s`", buf[:n])
	}

	message := []byte("foobar")
	n, err = slaveOutPipeWriter.Write(message)
	if err != nil {
		t.Fatalf("Unexpected error from Write(): %s", err)
	}
	if n != len(message) {
		t.Fatalf("Write() accepted `%d` for message `%s`", n, message)
	}

	n, err = connInPipeReader.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected error from Read(): %s", err)
	}
	if buf[0] != Output {
		t.Fatalf("Unexpected message type `%c`", buf[0])
	}
	decoded := make([]byte, 1024)
	n, err = base64.StdEncoding.Decode(decoded, buf[1:n])
	if err != nil {
		t.Fatalf("Unexpected error from Decode(): %s", err)
	}
	if !bytes.Equal(decoded[:n], message) {
		t.Fatalf("Unexpected message received: `%s`", decoded[:n])
	}

	cancel()
	wg.Wait()
}

func TestWriteFromConn(t *testing.T) {
	connInPipeReader, connInPipeWriter := io.Pipe()   // in to conn
	connOutPipeReader, connOutPipeWriter := io.Pipe() // out from conn
	slaveOutPipeReader, _ := io.Pipe()
	slaveInPipeReader, slaveInPipeWriter := io.Pipe()

	conn := pipePair{
		connOutPipeReader,
		connInPipeWriter,
	}
	slave := pipeSlave{
		slaveOutPipeReader,
		slaveInPipeWriter,
	}

	dt, err := New(conn, slave, WithPermitWrite())
	if err != nil {
		t.Fatalf("Unexpected error from New(): %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		err := dt.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Fatalf("Unexpected error from Run(): %s", err)
		}
	}()

	var (
		message []byte
		n       int
	)
	readBuf := make([]byte, 1024)

	n, err = connInPipeReader.Read(readBuf)
	if err != nil {
		t.Fatalf("Unexpected error from Read(): %s", err)
	}
	if !bytes.Equal(readBuf[:n], []byte{SetWindowTitle}) {
		t.Fatalf("Unexpected message received: `%s`", readBuf[:n])
	}

	// input
	message = []byte{Input, 'h', 'e', 'l', 'l', 'o', '\n'}
	n, err = connOutPipeWriter.Write(message)
	if err != nil {
		t.Fatalf("Unexpected error from Write(): %s", err)
	}
	if n != len(message) {
		t.Fatalf("Write() accepted `%d` for message `%s`", n, message)
	}

	n, err = slaveInPipeReader.Read(readBuf)
	if err != nil {
		t.Fatalf("Unexpected error from Write(): %s", err)
	}
	if !bytes.Equal(readBuf[:n], message[1:]) {
		t.Fatalf("Unexpected message received: `%s`", readBuf[:n])
	}

	// ping
	message = []byte{Ping}
	n, err = connOutPipeWriter.Write(message)
	if n != len(message) {
		t.Fatalf("Write() accepted `%d` for message `%s`", n, message)
	}

	n, err = connInPipeReader.Read(readBuf)
	if err != nil {
		t.Fatalf("Unexpected error from Read(): %s", err)
	}
	if !bytes.Equal(readBuf[:n], []byte{Pong}) {
		t.Fatalf("Unexpected message received: `%s`", readBuf[:n])
	}

	// TODO: resize

	cancel()
	wg.Wait()
}
