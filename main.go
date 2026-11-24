package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"sync"
)

var store = map[int][]byte{}
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

func Store(id int, data []byte) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	_, ok := store[id]
	if !ok {
		key := generateKey()
		store[id] = data
		return key, nil
	}
	return nil, errors.New("ID already exists")
}

func Retrieve(id int, key []byte) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	val, ok := store[id]
	if ok {
		return val, nil
	}
	return nil, errors.New("ID NOT FOUND")
}

func Update(id int, key, data []byte) error {
	mu.Lock()
	defer mu.Unlock()
	_, ok := store[id]
	if !ok {
		return errors.New("ID NOT FOUND")
	}

	store[id] = data
	return nil
}

func Delete(id int, key []byte) error {
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
