package main

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	ConnectPhrases = []string{
		"Unfortunately, %s has arrived.",
		"System alert: %s detected.",
		"Loading %s... (sigh)",
		"New object %s instantiated. Why?",
		"Oh great, it's %s.",
	}
	DisconnectPhrases = []string{
		"%s rage quit. Probably.",
		"Garbage collection ran on %s.",
		"Finally, %s left.",
		"%s process terminated.",
		"Releasing resources held by %s.",
	}
)

func announceConnect(username string) string {
	return fmt.Sprintf("%s %s", getTimestamp(), rand.Intn(len(ConnectPhrases)))
}

func announceDisconnect(username string) string {
	return fmt.Sprintf("%s %s", getTimestamp(), rand.Intn(len(DisconnectPhrases)))
}

func getTimestamp() string {
	return fmt.Sprintf("[%s]", time.Now().Format("15:04"))
}
