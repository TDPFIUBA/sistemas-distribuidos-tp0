package common

import (
	"bytes"
	"io"
	"net"
)

const (
	BUFFER_SIZE = 4
	END_MESSAGE = "\n"
)

type Protocol struct {
	conn net.Conn
}

func NewProtocol(conn net.Conn) *Protocol {
	return &Protocol{
		conn: conn,
	}
}

func (p *Protocol) SendMessage(data []byte) error {
	dataQtySent := 0
	dataQty := len(data)

	for dataQtySent < dataQty {
		sent, err := p.conn.Write(data[dataQtySent:])
		if err != nil {
			return err
		}
		dataQtySent += sent
	}
	return nil
}

func (p *Protocol) ReceiveMessage() (string, error) {
	data := make([]byte, 0)
	buffer := make([]byte, BUFFER_SIZE)
	end := []byte(END_MESSAGE)

	for {
		n, err := p.conn.Read(buffer)
		if err != nil && err != io.EOF {
			return "", err
		}

		if n == 0 {
			return "", io.EOF
		}

		data = append(data, buffer[:n]...)

		if bytes.Contains(buffer[:n], end) {
			break
		}
	}

	data = bytes.TrimSuffix(data, end)

	return string(data), nil
}
