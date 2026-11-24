package main

import (
	"context"
	"fmt"
	"log"
	"time"

	kvv1 "yoti/api/kv/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	conn, err := grpc.NewClient(
		"127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := kvv1.NewKeyValueStoreClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-name", "example")

	id := fmt.Sprint(time.Now().UnixNano())

	stored, err := client.Store(ctx, &kvv1.StoreRequest{
		Id:   id,
		Data: []byte("hello"),
	})
	if err != nil {
		log.Fatal(err)
	}

	key := stored.GetKey()

	retrieved, err := client.Retrieve(ctx, &kvv1.RetrieveRequest{
		Id:  id,
		Key: key,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(retrieved.GetData()))

	_, err = client.Update(ctx, &kvv1.UpdateRequest{
		Id:   id,
		Key:  key,
		Data: []byte("updated"),
	})
	if err != nil {
		log.Fatal(err)
	}

	retrieved, err = client.Retrieve(ctx, &kvv1.RetrieveRequest{
		Id:  id,
		Key: key,
	})
	if err != nil {
		log.Fatal(err)
	}

	if string(retrieved.GetData()) != "updated" {
		log.Fatal("Update did not change the stored value")
	}
	fmt.Println("updated")

	_, err = client.Delete(ctx, &kvv1.DeleteRequest{
		Id:  id,
		Key: key,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Retrieve(ctx, &kvv1.RetrieveRequest{
		Id:  id,
		Key: key,
	})
	if status.Code(err) != codes.NotFound {
		log.Fatalf("expected NotFound after Delete, got: %v", err)
	}
	fmt.Println("deleted")
}
