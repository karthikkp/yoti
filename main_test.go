package main

import "testing"

func TestStoreAndRetrieve(t *testing.T) {
	err := Store(1, "hello")
	if err != nil {
		t.Fatal("Store failed:", err)
	}

	got, err := Retrieve(1)
	if err != nil {
		t.Fatal("Retrieve failed:", err)
	}

	if got != "hello" {
		t.Fatal("expected hello, got:", got)
	}
}
