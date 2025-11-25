package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	terminal "golang.org/x/term"
)

var (
	clients   = make(map[ssh.Channel]struct{})
	clientsMu sync.Mutex
)

func broadcast(msg, user string, currentClient ssh.Channel) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if currentClient != client {
			_, _ = client.Write([]byte(user + " | " + msg + "\r\n"))
		}
	}
}

func main() {

	authorizedKeysBytes, err := os.ReadFile("keys.pub")
	if err != nil {
		log.Fatalf("Failed to load authorization keys: %v", err)
	}

	authorizedKeys := make(map[string]bool)

	for len(authorizedKeysBytes) > 0 {
		pubkey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatal(err)
		}

		authorizedKeys[string(pubkey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "testuser" && string(pass) == "tiger" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for: %q", c.User())
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeys[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("Unknown public key for: %q", c.User())
		},
	}

	privateBytes, err := os.ReadFile("keys")
	if err != nil {
		log.Fatal("failed to load private key ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("failed to parse private key ", err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept incoming connection: ", err)
		}

		go func() {
			conn, chans, reqs, err := ssh.NewServerConn(nConn, config)


			if err != nil {
				log.Fatal("failed to handshake: ", err)
			}

			//	log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])

			var wg sync.WaitGroup

			

			wg.Add(1)
			// go func() {
			// 	ssh.DiscardRequests(reqs)
			// 	wg.Done()
			// }()
			//
			wg.Go(func() {
				ssh.DiscardRequests(reqs)
				wg.Done()
			})

			for newChannel := range chans {
				if newChannel.ChannelType() != "session" {
					newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
					continue
				}
				channel, requests, err := newChannel.Accept()

				clientsMu.Lock()
				clients[channel] = struct{}{}
				clientsMu.Unlock()

				if err != nil {
					log.Fatalf("Could not accept channel: %v", err)
				}

				wg.Add(1)
				go func(in <-chan *ssh.Request) {
					for req := range in {
						switch req.Type {
						case "shell", "pty-req":
							req.Reply(true, nil)
						default:
							req.Reply(false, nil)
						}
					}
					wg.Done()
				}(requests)

				term := terminal.NewTerminal(channel, "> ")

				wg.Add(1)
				go func() {
					defer func() {
						channel.Close()
						clientsMu.Lock()
						delete(clients, channel)
						clientsMu.Unlock()
						wg.Done()
					}()
					for {
						line, err := term.ReadLine()
						if err != nil {
							break
						}
						broadcast(strings.TrimSpace(line), conn.User(),channel)
					}
				}()
			}
			wg.Wait()
			conn.Close()
			log.Printf("connection closed")
		}()
	}
}
