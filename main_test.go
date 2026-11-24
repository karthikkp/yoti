package main

import (
	"errors"
	"sync"
	"testing"
)

func mustStore(t *testing.T, caller, id string, data []byte) []byte {
	t.Helper()

	key, err := Store(caller, id, data)
	if err != nil {
		t.Fatal("setup failed:", err)
	}
	t.Cleanup(func() { _ = Delete(caller, id, key) })

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
			key, err := Store("svc", id, []byte("hello"))
			results <- result{key, err}
		}()
	}
	wg.Wait()
	close(results)

	wins := 0
	for r := range results {
		if r.err == nil {
			wins++
			t.Cleanup(func() { _ = Delete("svc", id, r.key) })
		} else {
			wantErr(t, r.err, ErrAlreadyExists)
		}
	}

	if wins != 1 {
		t.Fatal("expected exactly one Store to win, got:", wins)
	}
}

func TestUpdate(t *testing.T) {
	id := t.Name()
	key := mustStore(t, "svc", id, []byte("original"))
	before := string(store["svc/"+id])

	if err := Update("svc", id, key, []byte("updated")); err != nil {
		t.Fatal("Update failed:", err)
	}

	if string(store["svc/"+id]) == before {
		t.Fatal("update must change the ciphertext")
	}

	for _, badKey := range [][]byte{generateKey(), []byte("short")} {
		wantErr(t, Update("svc", id, badKey, []byte("unauthorized")), ErrInvalidKey)
	}

	got, err := Retrieve("svc", id, key)
	if err != nil || string(got) != "updated" {
		t.Fatal("expected updated, got:", string(got), err)
	}

	wantErr(t, Update("svc", "missing", key, []byte("anything")), ErrNotFound)
}

func TestDelete(t *testing.T) {
	id := t.Name()
	key := mustStore(t, "svc", id, []byte("hello"))

	// Wrong key: deletion must fail.
	wantErr(t, Delete("svc", id, generateKey()), ErrInvalidKey)

	// Correct key: deletion must succeed.
	if err := Delete("svc", id, key); err != nil {
		t.Fatal("Delete failed:", err)
	}

	_, err := Retrieve("svc", id, key)
	wantErr(t, err, ErrNotFound)

	// Deleting it again must fail.
	wantErr(t, Delete("svc", id, key), ErrNotFound)
}

func TestCallerIsolation(t *testing.T) {
	id := t.Name()
	keyA := mustStore(t, "a", id, []byte("a-value"))
	keyB := mustStore(t, "b", id, []byte("b-value"))

	_, err := Retrieve("b", id, keyA)
	wantErr(t, err, ErrInvalidKey)

	got, err := Retrieve("a", id, keyA)
	if err != nil || string(got) != "a-value" {
		t.Fatal("caller a must keep its own entry, got:", string(got), err)
	}

	got, err = Retrieve("b", id, keyB)
	if err != nil || string(got) != "b-value" {
		t.Fatal("caller b must keep its own entry, got:", string(got), err)
	}

	_, err = Store("a/b", id, []byte("x"))
	wantErr(t, err, ErrInvalidCaller)
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
