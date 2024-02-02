package database

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	err := db.ensureDB()
	if err != nil {
		return &DB{}, err
	}

	return &db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbS, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbS.Chirps) + 1

	req := Chirp{}
	err = json.Unmarshal([]byte(body), &req)
	if err != nil {
		return Chirp{}, errors.New("unmarshall error")
	}

	chp := Chirp{
		ID:   id,
		Body: req.Body,
	}

	dbS.Chirps[id] = chp
	err = db.writeDB(dbS)
	if err != nil {
		return Chirp{}, errors.New("writeDB error")
	}

	return chp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbS, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, len(dbS.Chirps))
	for _, chirp := range dbS.Chirps {
		chirps = append(chirps, chirp)
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})
	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)

	if os.IsNotExist(err) {
		file, err := os.Create(db.path)
		if err != nil {
			return err
		}
		defer file.Close()
	} else if err != nil {
		return err
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	body, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	dbS := DBStructure{}
	// If file isn't empty, unmarshal
	if len(body) != 0 {
		err = json.Unmarshal(body, &dbS)
		if err != nil {
			return DBStructure{}, errors.New("unmarshal loadDB error")
		}
	} else {
		dbS.Chirps = make(map[int]Chirp)
		dbS.Users = make(map[int]User)
		dbS.Tokens = make(map[string]int64)
	}

	return dbS, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0644)
	if err != nil {
		return err
	}

	return nil
}

// creates a new user and saves it to disk
func (db *DB) CreateUser(body string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbS, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id := len(dbS.Users) + 1

	req := User{}
	err = json.Unmarshal([]byte(body), &req)
	if err != nil {
		return User{}, errors.New("CreatUser: unmarshall error")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	u := User{
		ID:       id,
		Email:    req.Email,
		Password: string(hash),
	}

	dbS.Users[id] = u
	err = db.writeDB(dbS)
	if err != nil {
		return User{}, errors.New("couldn't write to db")
	}

	return u, nil
}

func (db *DB) GetUsers() ([]User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbS, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	users := make([]User, len(dbS.Users))
	for _, user := range dbS.Users {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	return users, nil
}

func (db *DB) UpdateUser(user *User) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	id := user.ID

	dbS, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user.Password = string(hash)

	dbS.Users[id] = *user
	err = db.writeDB(dbS)
	if err != nil {
		return User{}, errors.New("couldn't write to db")
	}

	return *user, nil
}

func (db *DB) RJWTNotExp(rt string) (bool, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return false, err
	}

	revokedTime, ok := dbs.Tokens[rt]
	if revokedTime != 0 {
		return !ok, nil
	}

	return ok, nil
}

func (db *DB) WriteRefreshToken(jwtString string, time int64) (string, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbS, err := db.loadDB()
	if err != nil {
		return "", err
	}

	dbS.Tokens[jwtString] = time
	err = db.writeDB(dbS)
	if err != nil {
		return "", errors.New("couldn't write to db")
	}

	return jwtString, nil
}

func (db *DB) WriteAccessToken(jwtString string) (string, error) {
	db.mux.Lock()
	defer db.mux.Lock()

	dbS, err := db.loadDB()
	if err != nil {
		return "", err
	}
	dbS.AToken = jwtString
	err = db.writeDB(dbS)
	if err != nil {
		return "", errors.New("couldn't write to db")
	}

	return jwtString, nil
}

func (db *DB) DeleteAccessToken() (string, error) {
	db.mux.Lock()
	defer db.mux.Lock()

	dbS, err := db.loadDB()
	if err != nil {
		return "", err
	}

	deleted := dbS.AToken
	dbS.AToken = ""
	err = db.writeDB(dbS)
	if err != nil {
		return "", errors.New("couldn't write to db")
	}

	return deleted, nil
}
