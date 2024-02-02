package database

import (
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp    `json:"chirps"`
	Users  map[int]User     `json:"users"`
	Tokens map[string]int64 `json:"refresh_tokens"`
}

type Chirp struct {
	ID     int    `json:"id"`
	Body   string `json:"body"`
	UserID int    `json:"user_id"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
