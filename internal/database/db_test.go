package database

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestNewDB(t *testing.T) {
	cases := []struct {
		path string
	}{
		{path: "db.json"},
		{path: "eatass.json"},
	}

	for _, cs := range cases {
		defer os.Remove(cs.path)
		_, err := NewDB(cs.path)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestLoadDB(t *testing.T) {
	db, err := NewDB("test.json")
	if err != nil {
		t.Error(err)
	}

	dbSCas := DBStructure{
		Chirps: make(map[int]Chirp),
	}

	for i := 1; i <= 5; i++ {
		dbSCas.Chirps[i] = Chirp{
			ID:   i,
			Body: fmt.Sprintf("body%v", i),
		}
	}

	dbSAct, err := db.loadDB()
	if err != nil {
		t.Error(err)
	}

	ok := reflect.DeepEqual(dbSAct, dbSCas)
	if !ok {
		t.Error("DBStructure actual is not the same as the case's")
	}
}

func TestGetChirps(t *testing.T) {
	db, err := NewDB("test.json")
	if err != nil {
		t.Error(err)
	}

	chpsCas := []Chirp{}
	for i := 1; i <= 5; i++ {
		chpsCas = append(chpsCas, Chirp{
			ID:   i,
			Body: fmt.Sprintf("body%v", i),
		})
	}

	chpsAct, err := db.GetChirps()
	if err != nil {
		t.Error(err)
	}

	t.Log(chpsAct)
	t.Log(chpsCas)

	ok := reflect.DeepEqual(chpsAct, chpsCas)
	if !ok {
		t.Error("actual and case chirps are not equal")
	}
}

func TestCreateChirp(t *testing.T) {
	db, err := NewDB("test.json")
	if err != nil {
		t.Error(err)
	}

	body := `
	{
		"body": "Hello, world!"
	}`

	_, err = db.CreateChirp(body)
	if err != nil {
		t.Error(err)
	}
}
