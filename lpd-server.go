/*
	Provides a Linux Printer Daemon server

	Ref: https://tools.ietf.org/html/rfc1179
*/
package lpdsrv

import (
	"io"
	"net"
	"strconv"

	"github.com/kgolding/go-decoder"
)

type Server struct {
	conn net.Listener
	// C emits received jobs
	C chan Job
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
		C: make(chan Job, 10),
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
				continue
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
	// job holds the currently processed job until it's ready to emit of the C channel
	job := Job{}

	// buf holds unprocessed data, and it has readBuf appended to it
	buf := make([]byte, 0)
	readBuf := make([]byte, 1024*24)

	for {
		// time.Sleep(time.Second)

		n, err := c.Read(readBuf)
		if err == io.EOF {
			if state == STATE_RECEIVE_DATA {
				// log.Printf("DO MS PRINT JOB:\n\tQue: %s\n\tJob: %s\n\tHost: %s\n %d bytes\n\n%s\n", que, job, host, len(buf), string(buf))
				c.Write([]byte{ACK}) // ACK
				c.Write([]byte{ACK}) // ACK
				c.Close()
				job.Data = buf
				s.C <- job
				job = Job{}
			}
			return
		} else if err != nil {
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
				// log.Printf("Receive a printer job for que '%s'\n", que)
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
					// log.Println("Waiting for more control file bytes...")
					c.Write([]byte{ACK}) // ACK
					continue
				}

				if cfa := string(dec.Bytes(3)); cfa != "cfA" {
					panic("Expected cfA got" + cfa)
				}
				job.Job, _ = strconv.Atoi(string(dec.Bytes(3)))

				job.Host = dec.StringByDelimiter(LF)

				job.controlFile = dec.Bytes(count + 1) // +1 for the 0x00

				// @TODO Decode the control file fields

				c.Write([]byte{ACK}) // ACK
				buf = dec.PeekRemainingBytes()

			case 0x03: // Receive file
				countStr := dec.StringByDelimiter(SPACE)
				count, err := strconv.Atoi(countStr)
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
					if dec.RemainingLength() < count {
						// log.Println("Waiting for more file bytes...")
						c.Write([]byte{ACK}) // ACK
						continue
					}
					data := dec.Bytes(count)
					// log.Printf("DO MS PRINT JOB:\n\tQue: %s\n\tJob: %s\n\tHost: %s\n %d bytes\n\n%s\n", que, job, host, count, string(data))
					c.Write([]byte{ACK}) // ACK
					c.Write([]byte{ACK}) // ACK
					state = STATE_IDLE
					job.Data = data
					s.C <- job
					job = Job{}
				}
			}
		}
	}
}
