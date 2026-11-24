package main

import (
	"errors"
	"sync"
)

var store = map[int]string{}
var mu sync.Mutex

func Store(id int, data string) error {
	_, ok := store[id]
	if !ok {
		store[id] = data
		return nil
	}
	return errors.New("ID already exists")
}

func Retrieve(id int) (string, error) {
	val, ok := store[id]
	if ok {
		return val, nil
	}
	return "", errors.New("ID NOT FOUND")
}

func Update(id int, data string) error {
	_, ok := store[id]
	if !ok {
		return errors.New("ID NOT FOUND")
	}

	store[id] = data
	return nil
}

func Delete(id int) error {
	_, ok := store[id]
	if !ok {
		return errors.New("ID NOT FOUND")
	}

	delete(store, id)
	return nil
}

func main() {

}
