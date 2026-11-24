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
			key, err := Store(id, []byte("hello"))
			if err != nil {
				t.Error("Store failed:", err)
				return
			}

			got, err := Retrieve(id, key)
			if err != nil {
				t.Error("Retrieve failed:", err)
				return
			}

			if string(got) != "hello" {
				t.Error("expected hello, got:", string(got))
			}
		}(id, t)
	}
	wg.Wait()

}

func TestUpdate(t *testing.T) {
	const id = 1001

	key, err := Store(id, []byte("original"))
	if err != nil {
		t.Fatal("setup failed:", err)
	}
	t.Cleanup(func() { _ = Delete(id, key) })

	err = Update(id, key, []byte("updated"))
	if err != nil {
		t.Fatal("Update failed:", err)
	}

	err = Update(id, generateKey(), []byte("unauthorized"))
	if err == nil {
		t.Fatal("Update accepted the wrong key")
	}

	got, err := Retrieve(id, key)
	if err != nil {
		t.Fatal("Retrieve failed:", err)
	}

	if string(got) != "updated" {
		t.Fatal("expected updated, got:", string(got))
	}

	err = Update(1999, key, []byte("anything"))
	if err == nil {
		t.Fatal("expected an error for a missing ID")
	}

	_, err = Retrieve(1999, key)
	if err == nil {
		t.Fatal("failed Update created an entry")
	}
}

func TestDelete(t *testing.T) {
	const id = 2001

	key, err := Store(id, []byte("hello"))
	if err != nil {
		t.Fatal("setup failed:", err)
	}
	t.Cleanup(func() { _ = Delete(id, key) })

	// Wrong key: deletion must fail.
	err = Delete(id, generateKey())
	if err == nil {
		t.Fatal("Delete accepted the wrong key")
	}

	// Correct key: deletion must succeed.
	err = Delete(id, key)
	if err != nil {
		t.Fatal("Delete failed:", err)
	}

	_, err = Retrieve(id, key)
	if err == nil {
		t.Fatal("entry still exists after Delete")
	}

	// Deleting it again must fail.
	err = Delete(id, key)
	if err == nil {
		t.Fatal("expected an error for a missing ID")
	}
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
