/*
	Provides a Linux Printer Daemon server

	Ref: https://tools.ietf.org/html/rfc1179
*/
package lpdsrv

import (
	"net"
	"strconv"

	"github.com/kgolding/go-decoder"
)

type Server struct {
	conn net.Listener
	// Job emits received jobs
	Job chan Job
}

const (
	ACK   = 0x00
	LF    = 0x0A
	SPACE = 0x20

	STATE_IDLE = iota
	STATE_RECEIVE_JOB
	STATE_RECEIVE_DATA
)

// New returns and starts a new Server instance on the given local IP/port
func New(address string) (*Server, error) {
	s := &Server{
		Job: make(chan Job, 10),
	}
	var err error

	s.conn, err = net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			c, err := s.conn.Accept()
			if err != nil {
				close(s.Job)
				return
			}
			go s.handleConnection(c)
		}
	}()

	return s, nil
}

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) handleConnection(c net.Conn) {

	state := STATE_IDLE

	// job holds the currently processed job until it's ready to emit on the C channel
	job := Job{}

	// buf holds unprocessed data, and it has readBuf appended to it
	buf := make([]byte, 0)
	readBuf := make([]byte, 1024*24)

	for {
		n, err := c.Read(readBuf)
		if err != nil {
			if state == STATE_RECEIVE_DATA {
				c.Write([]byte{ACK}) // ACK
				job.Data = buf
				s.Job <- job
			}
			c.Close()
			return
		}
		if n == 0 {
			continue
		}

		buf = append(buf, readBuf[:n]...)

		dec := decoder.New(buf)

		// function is the first byte which is the linux print daemon protocol function
		function := dec.Byte()

		switch state {
		case STATE_IDLE:
			switch function {
			case 0x02: // Receive a printer job
				job.Que = dec.StringByDelimiter(LF)
				c.Write([]byte{ACK}) // ACK
				state = STATE_RECEIVE_JOB
				buf = dec.PeekRemainingBytes()
			}

		case STATE_RECEIVE_JOB:
			switch function {
			case 0x01: // Abort
				state = STATE_IDLE
				buf = dec.PeekRemainingBytes()

			case 0x02: // Control file
				countStr := dec.StringByDelimiter(SPACE)
				count, err := strconv.Atoi(countStr)
				if err != nil {
					c.Close()
					return
				}
				if dec.RemainingLength() < count {
					// Waiting for more control file bytes
					c.Write([]byte{ACK}) // ACK
					continue
				}

				if cfa := string(dec.Bytes(3)); cfa != "cfA" {
					// Expected cfA
					c.Close()
					return
				}
				job.Job, _ = strconv.Atoi(string(dec.Bytes(3)))

				job.Host = dec.StringByDelimiter(LF)

				job.controlFile = dec.Bytes(count + 1) // +1 for the 0x00

				// @TODO Decode the control file fields

				c.Write([]byte{ACK}) // ACK
				buf = dec.PeekRemainingBytes()

			case 0x03: // Receive file
				countStr := dec.StringByDelimiter(SPACE)
				count, err := strconv.ParseUint(countStr, 10, 64)
				if err != nil {
					c.Close()
					return
				}

				if count > 99999999 { // Microsoft Windows sends silly size when it doesn't know the size of the print job
					c.Write([]byte{ACK}) // ACK
					// Next packet(s) will be the data
					state = STATE_RECEIVE_DATA
					buf = []byte{} // Start with a clean buf
					continue
				} else {
					if uint64(dec.RemainingLength()) < count {
						// Waiting for more bytes
						c.Write([]byte{ACK}) // ACK
						continue
					}
					dec.StringByDelimiter(LF) // Name of datafile (dfA)

					data := dec.Bytes(int(count))
					c.Write([]byte{ACK}) // ACK
					state = STATE_IDLE
					job.Data = data
					s.Job <- job
					job = Job{}
				}
			}
		}
	}
}
