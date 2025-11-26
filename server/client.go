package main

import (
	"fmt"
	"sync"

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
}
