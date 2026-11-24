package main

import (
	"context"
	"net"
	"testing"
	"time"

	kvv1 "yoti/api/kv/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func newClient(t *testing.T) kvv1.KeyValueStoreClient {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer()
	kvv1.RegisterKeyValueStoreServer(server, &kvServer{})
	go server.Serve(listener)
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return kvv1.NewKeyValueStoreClient(conn)
}

func testCtx(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	return ctx
}

func TestGRPCFlow(t *testing.T) {
	client := newClient(t)
	ctx := testCtx(t)
	id := t.Name()

	stored, err := client.Store(ctx, &kvv1.StoreRequest{Id: id, Data: []byte("hello")})
	if err != nil {
		t.Fatal("Store failed:", err)
	}
	key := stored.GetKey()
	t.Cleanup(func() { _ = Delete(id, key) })

	if len(key) != 32 {
		t.Fatal("expected a 32-byte key")
	}

	_, err = client.Store(ctx, &kvv1.StoreRequest{Id: id, Data: []byte("replacement")})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatal("expected AlreadyExists, got:", err)
	}

	_, err = client.Update(ctx, &kvv1.UpdateRequest{Id: id, Key: key, Data: []byte("updated")})
	if err != nil {
		t.Fatal("Update failed:", err)
	}

	retrieved, err := client.Retrieve(ctx, &kvv1.RetrieveRequest{Id: id, Key: key})
	if err != nil || string(retrieved.GetData()) != "updated" {
		t.Fatal("updated value was not retrieved:", retrieved, err)
	}

	_, err = client.Delete(ctx, &kvv1.DeleteRequest{Id: id, Key: key})
	if err != nil {
		t.Fatal("Delete failed:", err)
	}

	_, err = client.Retrieve(ctx, &kvv1.RetrieveRequest{Id: id, Key: key})
	if status.Code(err) != codes.NotFound {
		t.Fatal("expected NotFound, got:", err)
	}
}

func TestGRPCPublicErrors(t *testing.T) {
	client := newClient(t)
	ctx := testCtx(t)
	id := t.Name()

	stored, err := client.Store(ctx, &kvv1.StoreRequest{Id: id, Data: []byte("hello")})
	if err != nil {
		t.Fatal("Store failed:", err)
	}
	key := stored.GetKey()
	t.Cleanup(func() { _ = Delete(id, key) })

	ops := map[string]func(string, []byte) error{
		"Retrieve": func(id string, key []byte) error {
			_, err := client.Retrieve(ctx, &kvv1.RetrieveRequest{Id: id, Key: key})
			return err
		},
		"Update": func(id string, key []byte) error {
			_, err := client.Update(ctx, &kvv1.UpdateRequest{Id: id, Key: key, Data: []byte("x")})
			return err
		},
		"Delete": func(id string, key []byte) error {
			_, err := client.Delete(ctx, &kvv1.DeleteRequest{Id: id, Key: key})
			return err
		},
	}

	cases := []struct {
		name string
		id   string
		key  []byte
	}{
		{"wrong key", id, generateKey()},
		{"missing entry", "missing", key},
	}

	for op, call := range ops {
		for _, c := range cases {
			err := call(c.id, c.key)
			if status.Code(err) != codes.NotFound || status.Convert(err).Message() != "entry unavailable" {
				t.Errorf("%s/%s: expected NotFound entry unavailable, got: %v", op, c.name, err)
			}
		}
	}
}
