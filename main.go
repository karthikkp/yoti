package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"log"
	"net"
	"os"
	"sync"

	kvv1 "yoti/api/kv/v1"

	"google.golang.org/grpc"
)

var store = map[string][]byte{}
var mu sync.Mutex

var (
	ErrNotFound      = errors.New("ID not found")
	ErrAlreadyExists = errors.New("ID already exists")
	ErrInvalidKey    = errors.New("invalid encryption key")
)

var debugMode = os.Getenv("DEBUG") == "1"

func printStoreLocked(operation string) {
	if !debugMode {
		return
	}

	log.Printf("%s: store has %d entries", operation, len(store))
	for id, encrypted := range store {
		log.Printf("id=%s ciphertext=%x", id, encrypted)
	}
}

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
		return nil, ErrInvalidKey
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	data, err := gcm.Open(nil, nil, encrypted, nil)
	if err != nil {
		return nil, ErrInvalidKey
	}

	return data, nil
}

func Store(id string, data []byte) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	defer printStoreLocked("STORE")

	_, ok := store[id]
	if !ok {
		key := generateKey()

		encrypted, err := encrypt(data, key)
		if err != nil {
			return nil, err
		}

		store[id] = encrypted
		return key, nil
	}
	return nil, ErrAlreadyExists
}

func Retrieve(id string, key []byte) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	defer printStoreLocked("STORE")

	val, ok := store[id]
	if ok {
		return decrypt(val, key)
	}
	return nil, ErrNotFound
}

func Update(id string, key, data []byte) error {
	mu.Lock()
	defer mu.Unlock()
	defer printStoreLocked("STORE")

	existing, ok := store[id]
	if !ok {
		return ErrNotFound
	}

	_, err := decrypt(existing, key)
	if err != nil {
		return err
	}

	encrypted, err := encrypt(data, key)
	if err != nil {
		return err
	}

	store[id] = encrypted
	return nil
}

func Delete(id string, key []byte) error {
	mu.Lock()
	defer mu.Unlock()
	defer printStoreLocked("STORE")

	encrypted, ok := store[id]
	if !ok {
		return ErrNotFound
	}

	if _, err := decrypt(encrypted, key); err != nil {
		return err
	}

	delete(store, id)
	return nil
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	kvv1.RegisterKeyValueStoreServer(server, &kvServer{})

	log.Println("listening on 127.0.0.1:50051")

	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
