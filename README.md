# Simple Linux Printer Daemon in Go

Implements a minimum subset of https://tools.ietf.org/html/rfc1179 as tested using Windows 10 as the print client.

## Usage

See [cmd/main.go](cmd/main.go) for an example.

`import  "github.com/kgolding/go-lpdsrv"`

Create a new server instance

`s, err := lpdsrv.New("0.0.0.0:515")`

Wait for a incoming request, and check for the channel closing

`job, ok := <-s.Job`

The job includes the raw data along with the host-name of the client, the name of the printer queue and the job identification number.

```Go
type Job struct {
	Host        string
	Que         string
	Job         int
	Data        []byte
}
```

The server can be closed using `s.Close()`.

## Creating a Windows 10 virtual printer

See [docs/windows-10-setup.md](docs/windows-10-setup.md)