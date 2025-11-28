package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"golang.org/x/crypto/ssh"
	// terminal "golang.org/x/term"
)

//
// var (
// 	clients   = make(map[ssh.Channel]struct{})
// 	clientsMu sync.Mutex
// )

func main() {

	loadUsers()

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
			pub := string(ssh.MarshalAuthorizedKey(pubKey))

			userMu.Lock()
			var user *User
			for _, u := range userStore.Users {
				if u.Key == pub {
					user = &u
					break
				}
			}
			if user == nil {
				user = &User{
					ID:  generateUID(),
					Username: c.User(),
					Key: pub,
				}
				userStore.Users = append(userStore.Users, *user)
				saveUsers()
			}
			userMu.Unlock()

			// if authorizedKeys[string(pubKey.Marshal())] {
			// 	return &ssh.Permissions{
			// 		Extensions: map[string]string{
			// 			"pubkey-fp": ssh.FingerprintSHA256(pubKey),
			// 		},
			// 	}, nil
			// }
			return &ssh.Permissions{
				Extensions: map[string]string{
					"id":       strconv.Itoa(user.ID),
					"username": user.Username,
				},
			}, nil
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
			continue
		}

		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			fmt.Println("SSH Handshake failure: %v", err)
			continue
		}

		userId, _ := strconv.Atoi(conn.Permissions.Extensions["id"])
		client := getUser(userId)
		if client == nil {
			fmt.Printf("User not found.")
		}

		var wg sync.WaitGroup

		wg.Add(1)
		wg.Go(func() {
			ssh.DiscardRequests(reqs)
			wg.Done()
		})

		go func() {
			for newChannel := range chans {
				if newChannel.ChannelType() != "session" {
					newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
					continue
				}

				clientsMu.Lock()
				if _, exists := clients[userId]; exists {
					msg := fmt.Sprintf("Login Failed: We already have a [%s], and one is plenty.", getUserName(userId))
					newChannel.Reject(ssh.Prohibited, msg)
					continue
				}
				clientsMu.Unlock()

				channel, _, _ := newChannel.Accept()
				clientsMu.Lock()
				clients[userId] = channel
				clientsMu.Unlock()

				go handleClient(channel, userId)
			}
		}()

		//
		// 	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		//
		// 	if err != nil {
		// 		log.Fatal("failed to handshake: ", err)
		// 	}
		//
		// 	//	log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])
		//
		// 	var wg sync.WaitGroup
		//
		// 	wg.Add(1)
		// 	wg.Go(func() {
		// 		ssh.DiscardRequests(reqs)
		// 		wg.Done()
		// 	})
		//
		// 	for newChannel := range chans {
		// 		if newChannel.ChannelType() != "session" {
		// 			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		// 			continue
		// 		}
		// 		channel, requests, err := newChannel.Accept()
		//
		// 		clientsMu.Lock()
		// 		if _, exists := clients[uid]; exists {
		// 			msg := fmt.Sprintf("Login Failed: We already have a [%s], and one is plenty.", getUserName(uid))
		// 			channel.Write([]byte(msg))
		// 			channel.Close()
		// 		}
		//
		// }
	}

	//
	// 	clientsMu.Lock()
	// 	clients[channel] = struct{}{}
	// 	clientsMu.Unlock()
	//
	// 	if err != nil {
	// 		log.Fatalf("Could not accept channel: %v", err)
	// 	}
	//
	// 	wg.Add(1)
	// 	go func(in <-chan *ssh.Request) {
	// 		for req := range in {
	// 			switch req.Type {
	// 			case "shell", "pty-req":
	// 				req.Reply(true, nil)
	// 			default:
	// 				req.Reply(false, nil)
	// 			}
	// 		}
	// 		wg.Done()
	// 	}(requests)
	//
	// 	term := terminal.NewTerminal(channel, "> ")
	//
	// 	wg.Add(1)
	// 	go func() {
	// 		defer func() {
	// 			channel.Close()
	// 			clientsMu.Lock()
	// 			delete(clients, channel)
	// 			clientsMu.Unlock()
	// 			wg.Done()
	// 		}()
	// 		for {
	// 			line, err := term.ReadLine()
	// 			if err != nil {
	// 				break
	// 			}
	// 			broadcast(strings.TrimSpace(line), conn.User(), channel)
	// 		}
	// 	}()
	// }
	// wg.Wait()
	// conn.Close()
	// log.Printf("connection closed")
}
