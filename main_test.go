package main

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

func mustStore(t *testing.T, id string, data []byte) []byte {
	t.Helper()

	key, err := Store(id, data)
	if err != nil {
		t.Fatal("setup failed:", err)
	}
	t.Cleanup(func() { _ = Delete(id, key) })

	return key
}

func wantErr(t *testing.T, got, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got: %v", want, got)
	}
}

func TestConcurrentStoreSameID(t *testing.T) {
	type result struct {
		key []byte
		err error
	}

	id := t.Name()
	results := make(chan result, 20)

	var wg sync.WaitGroup
	for i := 0; i < cap(results); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key, err := Store(id, []byte("hello"))
			results <- result{key, err}
		}()
	}
	wg.Wait()
	close(results)

	wins := 0
	for r := range results {
		if r.err == nil {
			wins++
			t.Cleanup(func() { _ = Delete(id, r.key) })
		} else {
			wantErr(t, r.err, ErrAlreadyExists)
		}
	}

	if wins != 1 {
		t.Fatal("expected exactly one Store to win, got:", wins)
	}
}

func TestPerEntryEncryption(t *testing.T) {
	plaintext := []byte("secret-value")
	id1, id2 := t.Name()+"-1", t.Name()+"-2"

	key1 := mustStore(t, id1, plaintext)
	key2 := mustStore(t, id2, plaintext)

	if bytes.Equal(key1, key2) {
		t.Fatal("entries must get different keys")
	}

	if bytes.Contains(store[id1], plaintext) {
		t.Fatal("stored bytes must not contain the plaintext")
	}
}

func TestUpdate(t *testing.T) {
	id := t.Name()
	key := mustStore(t, id, []byte("original"))
	before := string(store[id])

	if err := Update(id, key, []byte("updated")); err != nil {
		t.Fatal("Update failed:", err)
	}

	if string(store[id]) == before {
		t.Fatal("update must change the ciphertext")
	}

	for _, badKey := range [][]byte{generateKey(), []byte("short")} {
		wantErr(t, Update(id, badKey, []byte("unauthorized")), ErrInvalidKey)
	}

	got, err := Retrieve(id, key)
	if err != nil || string(got) != "updated" {
		t.Fatal("expected updated, got:", string(got), err)
	}

	wantErr(t, Update("missing", key, []byte("anything")), ErrNotFound)
}

func TestDelete(t *testing.T) {
	id := t.Name()
	key := mustStore(t, id, []byte("hello"))

	// Wrong key: deletion must fail.
	wantErr(t, Delete(id, generateKey()), ErrInvalidKey)

	// Correct key: deletion must succeed.
	if err := Delete(id, key); err != nil {
		t.Fatal("Delete failed:", err)
	}

	_, err := Retrieve(id, key)
	wantErr(t, err, ErrNotFound)

	// Deleting it again must fail.
	wantErr(t, Delete(id, key), ErrNotFound)
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
	if !errors.Is(err, ErrInvalidKey) {
		t.Fatal("expected ErrInvalidKey, got:", err)
	}

	if len(data) != 0 {
		t.Fatal("decryption returned data despite failing")
	}
}
