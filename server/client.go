package main

import (
	"fmt"
	"io"
	"sync"


	"golang.org/x/term"

	"golang.org/x/crypto/ssh"
)

var (
	clients   = map[int]ssh.Channel{}
	clientsMu sync.Mutex
)

func handleClient(ch ssh.Channel, uid int) {
	defer func() {
		clientsMu.Lock()
		if _, ok := clients[uid]; ok {
			delete(clients, uid)
			ch.Close()
		}
		clientsMu.Unlock()
		fmt.Printf("[CLEANUP] Freeing memory used by [%s]. It was tainted anyway", getUserName(uid))
		broadcastLeave(uid)
	}()

	userMu.Lock()
	user := getUser(uid)
	userMu.Unlock()

	// username change feature later.

	fmt.Printf("[NET] Handshake accepted for [%s]. Waste of bandwidth.", user.Username)
	broadcastJoin(uid)

	terminal := term.NewTerminal(ch, "> ")

	for {
		line, err := terminal.ReadLine()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			break
		}

		if line == "" {
			continue
		}

			HandleCommand(ch, uid, line)

		if len(line) > 2048 {
			ch.Write([]byte("Error: I'm not reading all that. Summarize.\n"))
			continue
		}

		fmt.Printf("Received line: %s", line)

		ch.Write([]byte(line + "\n"))

	}

	//	buf := make([]byte, 2048)
	//
	// scanner := bufio.NewScanner(ch)
	//
	// for scanner.Scan() {
	// 	msg := strings.TrimSpace(scanner.Text())
	// 	if msg == "" {
	// 		continue
	// 	}
	//
	// 	if len(msg) > 2048 {
	// 		ch.Write([]byte("Error: I'm not reading all that. Summarize.\n"))
	// 		continue
	// 	}
	//
	// 	fmt.Println("Received:", msg)
	//
	// 	// Echo the message back
	// 	ch.Write([]byte(msg + "\n"))
	// }

	// for {
	// 	n, err := ch.Read(buf)
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		fmt.Println("Read error: %v", err)
	// 		continue
	// 	}
	//
	// 	msg := strings.TrimSpace(string(buf[:n]))
	// 	fmt.Println(msg)
	// 	if msg == "" || len(msg) > 2048 {
	// 		if len(msg) > 2048 {
	// 			ch.Write([]byte("Error: I'm not reading all that. Summarize."))
	// 		}
	// 		continue
	// 	}
	//
	// 	ch.Write([]byte(msg + "\n"))
	//
	// }

}
