package hsocks5

import (
	"io"
)

func pipe(c1, c2 io.ReadWriter, errChan chan error) {
	_, err := io.CopyBuffer(c1, c2, make([]byte, 32*1024))
	errChan <- err
}
