package main

import (
	"context"
	"log"
	kvv1 "yoti/api/kv/v1"

	"errors"

	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type kvServer struct {
	kvv1.UnimplementedKeyValueStoreServer
}

func accessError(op, id string, err error) error {
	log.Printf("%s id=%s error=%v", op, id, err)

	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidKey) {
		return status.Error(codes.NotFound, "entry unavailable")
	}

	return status.Error(codes.Internal, "internal error")
}

func toRPCError(op, id string, err error) error {
	log.Printf("%s id=%s error=%v", op, id, err)

	if errors.Is(err, ErrAlreadyExists) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	return status.Error(codes.Internal, "internal error")
}

func (s *kvServer) Store(
	ctx context.Context,
	req *kvv1.StoreRequest,
) (*kvv1.StoreResponse, error) {
	key, err := Store(req.GetId(), req.GetData())
	if err != nil {
		return nil, toRPCError("Store", req.GetId(), err)
	}

	return &kvv1.StoreResponse{Key: key}, nil
}

func (s *kvServer) Retrieve(
	ctx context.Context,
	req *kvv1.RetrieveRequest,
) (*kvv1.RetrieveResponse, error) {
	data, err := Retrieve(req.GetId(), req.GetKey())
	if err != nil {
		return nil, accessError("Retrieve", req.GetId(), err)
	}

	return &kvv1.RetrieveResponse{Data: data}, nil
}

func (s *kvServer) Update(
	ctx context.Context,
	req *kvv1.UpdateRequest,
) (*emptypb.Empty, error) {
	err := Update(req.GetId(), req.GetKey(), req.GetData())
	if err != nil {
		return nil, accessError("Update", req.GetId(), err)
	}

	return &emptypb.Empty{}, nil
}

func (s *kvServer) Delete(
	ctx context.Context,
	req *kvv1.DeleteRequest,
) (*emptypb.Empty, error) {
	err := Delete(req.GetId(), req.GetKey())
	if err != nil {
		return nil, accessError("Delete", req.GetId(), err)
	}

	return &emptypb.Empty{}, nil
}
