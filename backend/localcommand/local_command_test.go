package localcommand

import (
	"bytes"
	"io"
	"os/exec"
	"testing"
	"time"
)

func TestNewStartsBashWithPTY(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not found")
	}

	cmd, err := New("bash", []string{"-lc", "printf GOTTY_PTY_OK"})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer cmd.Close()

	read := make(chan []byte, 1)
	readErr := make(chan error, 1)
	go func() {
		buf := make([]byte, 1024)
		var output []byte
		for {
			n, err := cmd.Read(buf)
			if n > 0 {
				output = append(output, buf[:n]...)
				if bytes.Contains(output, []byte("GOTTY_PTY_OK")) {
					read <- output
					return
				}
			}
			if err != nil {
				if err == io.EOF {
					read <- output
					return
				}
				readErr <- err
				return
			}
		}
	}()

	select {
	case err := <-readErr:
		t.Fatalf("Read() failed: %v", err)
	case got := <-read:
		if !bytes.Contains(got, []byte("GOTTY_PTY_OK")) {
			t.Fatalf("unexpected PTY output: %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for PTY output")
	}
}
