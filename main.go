package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"sync"
)

var store = map[int]string{}
var mu sync.Mutex

func generateKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

func encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nil, nil, data, nil)
	return encrypted, nil
}

func decrypt(encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nil, encrypted, nil)
}

func Store(id int, data string) error {
	mu.Lock()
	defer mu.Unlock()

	_, ok := store[id]
	if !ok {
		store[id] = data
		return nil
	}
	return errors.New("ID already exists")
}

func Retrieve(id int) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	val, ok := store[id]
	if ok {
		return val, nil
	}
	return "", errors.New("ID NOT FOUND")
}

func Update(id int, data string) error {
	mu.Lock()
	defer mu.Unlock()
	_, ok := store[id]
	if !ok {
		return errors.New("ID NOT FOUND")
	}

	store[id] = data
	return nil
}

func Delete(id int) error {
	mu.Lock()
	defer mu.Unlock()
	_, ok := store[id]
	if !ok {
		return errors.New("ID NOT FOUND")
	}

	delete(store, id)
	return nil
}

func main() {

}
