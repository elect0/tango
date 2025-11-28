package main

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

func listOnline(out ssh.Channel) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if len(clients) == 1 {
		out.Write([]byte("Query returned 0 friends. Accuracy: 100%."))
		return
	}
	var ids []string
	for id := range clients {
		ids = append(ids, fmt.Sprintf("%s", getUserName(id)))
	}
	out.Write([]byte("< SYSTEM > Online Users: " + strings.Join(ids, ", ") + "\n"))
}

func HandleCommand(out ssh.Channel, id int, msg string) {
	fields := strings.Fields(msg)

	if strings.HasPrefix(msg, "#") {
		cmd := strings.Trim(fields[0], "#")

		switch cmd {
		case "help":

			helpText := `
================================================================
 < SYSTEM > I assumed you were smart enough to use a chat client.
================================================================

 /quit   : Rage quit. Go touch grass.
 /nick   : Change identity. The FBI is still watching.
 /clear  : Wipe screen. Pretend that typo didn't happen.
 /list   : List other victims currently online.
 /shout  : Send in BOLD. Attention seeker mode.
 /dm     : /dm [user] [msg]. Gossip behind people's backs.
 /info   : /info [user]. Stalking made easy.
 /help   : Recursion error. You are looking at it.

----------------------------------------------------------------
 > End of manual. Please try to keep up.
			`
			out.Write([]byte(helpText))
		case "quit":
			out.Write([]byte("Restoring your terminal to its boring default state... Success." + "\n"))
			out.Close()
		case "list":
			listOnline(out)
		}
	}
}
