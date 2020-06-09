# SSH

Example

```
package main

import (
	"log"
	"fmt"

	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/GolangResources/ssh/v1"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			ForIni:
				log.Println("Starting connection")
				config := &ssh.ClientConfig{
					User: "ubuntu",
					Auth: []ssh.AuthMethod{ssh.Password("12345678")},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}
				serverConn, err := ssh.Dial("tcp", "127.0.0.1:22", config)
				if err != nil {
					fmt.Printf("Server dial error: %s\n", err)
					time.Sleep(1* time.Second)
					goto ForIni
				}

				tunConfig := gssh.SSHClient{
					Debug: false,
					Client: serverConn,
				}
				d := gssh.Init(&tunConfig)
				var wgt sync.WaitGroup
				// ssh -D
				wgt.Add(1)
				go func(s *gssh.SSHClient) {
					defer wgt.Done()
					for {
						d.Dynamic("127.0.0.1:1111")
					}
				}(&d)
				// ssh -L 
				wgt.Add(1)
				go func(s *gssh.SSHClient) {
					defer wgt.Done()
					for {
						d.TCPTunnel("127.0.0.1:2222", "127.0.0.1:22")
					}
				}(&d)
				// ssh -R
                                wgt.Add(1)
                                go func(s *gssh.SSHClient) {
                                        defer wgt.Done()
                                        for {
                                                d.RTCPTunnel("127.0.0.1:2222", "127.0.0.1:22")
                                        }
                                }(&d)
				wgt.Wait()
		}
	}(&wg)
	wg.Wait()
}
```
