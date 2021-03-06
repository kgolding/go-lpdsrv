/*
	Super simple Linux Printer Daemon
*/
package main

import (
	"log"
	"strings"

	"github.com/kgolding/go-lpdsrv"
)

func main() {
	/*
		LPD uses port 515, which means you're need to run this as root or a clever way around this is
		to port forward 515 to 1515 using `sudo iptables -t nat -A PREROUTING -p tcp --dport 515 -j REDIRECT --to-port 1515`
	*/
	s, err := lpdsrv.New("0.0.0.0:515")
	if err != nil {
		panic(err)
	}
	for {
		select {
		case job, ok := <-s.Job:
			if !ok {
				log.Println("Print server has closed")
				return
			}
			log.Println("JOB:", job.String())
			log.Printf("Trimed: '%s'\n--------------------------------------------------------------------\n",
				strings.TrimSpace(string(job.Data)),
			)
		}
	}
}
