package main

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Key      string `json:"key"`
}

type UserStore struct {
	Users []User `json:"users"`
}

var (
	userStore = UserStore{Users: []User{}}
	userMu    sync.Mutex
)

func loadUsers() error {
	data, err := os.ReadFile("users.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &userStore)
	if err != nil {
		return err
	}
	return nil
}

func saveUsers() {
	data, _ := json.MarshalIndent(userStore, "", " ")
	_ = os.WriteFile("users.json", data, 0644)
}

func generateUID() int {
	max := 0
	for _, u := range userStore.Users {
		if u.ID > max {
			max = u.ID
		}
	}
	return max + 1
}

func getUserName(uid int) string {
	for _, u := range userStore.Users {
		if u.ID == uid {
			return u.Username
		}
	}
	return ""
}

func getUser[T interface{ ~int | ~string }](value T) *User {
	switch v := any(value).(type) {
	case int:
		for _, u := range userStore.Users {
			if u.ID == v {
				return &u
			}
		}
	case string:
		if uid, err := strconv.Atoi(v); err == nil {
			for _, u := range userStore.Users {
				if u.ID == uid {
					return &u
				}
			}
		} else {
			for _, u := range userStore.Users {
				if u.Username == v {
					return &u
				}
			}
		}
	}
	return nil
}
