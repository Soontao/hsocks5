package hsocks5

import (
	"io"
	"net"
)

func pipe(c1, c2 net.Conn, errChan chan error) {
	_, err := io.Copy(c1, c2)
	c1.Close()
	c2.Close()
	errChan <- err
}
