package main

import (
	"sync"
	"testing"
)

func TestStoreAndRetrieve(t *testing.T) {
	var wg sync.WaitGroup
	for id := 0; id < 50; id++ {
		wg.Add(1)
		go func(id int, t *testing.T) {
			defer wg.Done()
			err := Store(id, "hello")
			if err != nil {
				t.Error("Store failed:", err)
			}

			got, err := Retrieve(id)
			if err != nil {
				t.Error("Retrieve failed:", err)
			}

			if got != "hello" {
				t.Error("expected hello, got:", got)
			}
		}(id, t)
	}
	wg.Wait()

}

func TestEncryptionRoundTrip(t *testing.T) {
	key := generateKey()

	// Encrypt the original data.
	encrypted, err := encrypt([]byte("hello"), key)
	if err != nil {
		t.Fatal("encrypt failed:", err)
	}

	// The correct key must recover the original data.
	got, err := decrypt(encrypted, key)
	if err != nil {
		t.Fatal("decrypt failed:", err)
	}

	if string(got) != "hello" {
		t.Fatal("expected hello, got:", string(got))
	}

	// A different key must fail.
	wrongKey := generateKey()

	data, err := decrypt(encrypted, wrongKey)
	if err == nil {
		t.Fatal("expected an error for the wrong key")
	}

	if len(data) != 0 {
		t.Fatal("decryption returned data despite failing")
	}
}
