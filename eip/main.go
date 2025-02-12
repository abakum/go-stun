// Copyright 2016 Cong Ding
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abakum/go-stun/stun"
)

func main() {
	servers := os.Args[1:]
	flag.Usage = func() {
		fmt.Println("Get the external IP address from the STUN servers listed on the command line. For example: `eip stun.sipnet.ru:3478 stun.l.google.com:19302`")
	}
	if len(servers) == 0 {
		flag.Usage()
		fmt.Printf("Now run: `eip %s`\n", stun.DefaultServerAddr)
		servers = append(servers, stun.DefaultServerAddr)
	}
	flag.Parse()
	_, m, _ := GetExternalIP(time.Second, servers...)
	fmt.Println(m)
}

func GetExternalIP(timeout time.Duration, servers ...string) (ip, message string, err error) {
	type IPfromSince struct {
		IP, From string
		Since    time.Duration
		Err      error
	}

	ch := make(chan *IPfromSince)
	defer close(ch)
	t := time.AfterFunc(timeout, func() {
		ch <- &IPfromSince{"", strings.Join(servers, ","), timeout, fmt.Errorf("timeout")}
	})
	defer t.Stop()
	for _, server := range servers {
		go func(s string) {
			client := stun.NewClient()
			client.SetServerAddr(s)
			t := time.Now()
			ip, err := client.GetExternalIP()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err, "from", s)
				return
			}
			// time.Sleep(time.Second)
			ch <- &IPfromSince{ip, s, time.Since(t), nil}
		}(server)
	}
	i := <-ch
	message = fmt.Sprint(i.Err, " get external IP")
	if i.Err == nil {
		message = fmt.Sprint("External IP: ", i.IP)
	}
	message += fmt.Sprint(" from ", i.From, " since ", i.Since.Seconds(), "s")

	if i.Err != nil {
		return "127.0.0.1", message, fmt.Errorf("%s", message)
	}
	return i.IP, message, nil
}
