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
