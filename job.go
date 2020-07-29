package lpdsrv

import (
	"fmt"
)

type Job struct {
	Host        string
	Que         string
	Job         int
	controlFile []byte
	Data        []byte
}

func (j *Job) String() string {
	return fmt.Sprintf("ID: %d\nQue: %s\nHost: %s\nData: % X", j.Job, j.Que, j.Host, j.Data)
}
