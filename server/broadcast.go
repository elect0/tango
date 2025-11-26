package main

import "fmt"

func broadcastJoin(uid int) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for id, ch := range clients {
		var msg string
		if id == uid {
			msg = fmt.Sprintf("Restoring your terminal to its boring default state... Success.")
		} else {
			msg = announceDisconnect(getUserName(uid))
		}
		ch.Write([]byte(msg))
	}
}

func broadcastLeave(uid int) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	msg := announceDisconnect(getUserName(uid))
	for _, ch := range clients {
		ch.Write([]byte(msg))
	}
}
