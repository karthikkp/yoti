package main

import (
	"context"
	"log"
	kvv1 "yoti/api/kv/v1"

	"errors"

	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type kvServer struct {
	kvv1.UnimplementedKeyValueStoreServer
}

func namespace(ctx context.Context) (string, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	v := md.Get("x-service-name")
	if len(v) != 1 || v[0] == "" {
		return "", status.Error(codes.Unauthenticated, "missing or invalid service name")
	}
	return v[0], nil
}

func accessError(op, ns, id string, err error) error {
	log.Printf("%s ns=%s id=%s error=%v", op, ns, id, err)

	if errors.Is(err, ErrInvalidCaller) {
		return status.Error(codes.Unauthenticated, "missing or invalid service name")
	}

	if errors.Is(err, ErrEmptyData) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidKey) {
		return status.Error(codes.NotFound, "entry unavailable")
	}

	return status.Error(codes.Internal, "internal error")
}

func toRPCError(op, ns, id string, err error) error {
	log.Printf("%s ns=%s id=%s error=%v", op, ns, id, err)

	if errors.Is(err, ErrInvalidCaller) {
		return status.Error(codes.Unauthenticated, "missing or invalid service name")
	}

	if errors.Is(err, ErrEmptyData) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if errors.Is(err, ErrAlreadyExists) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	return status.Error(codes.Internal, "internal error")
}

func (s *kvServer) Store(
	ctx context.Context,
	req *kvv1.StoreRequest,
) (*kvv1.StoreResponse, error) {
	ns, err := namespace(ctx)
	if err != nil {
		return nil, err
	}

	key, err := Store(ns, req.GetId(), req.GetData())
	if err != nil {
		return nil, toRPCError("Store", ns, req.GetId(), err)
	}

	return &kvv1.StoreResponse{Key: key}, nil
}

func (s *kvServer) Retrieve(
	ctx context.Context,
	req *kvv1.RetrieveRequest,
) (*kvv1.RetrieveResponse, error) {
	ns, err := namespace(ctx)
	if err != nil {
		return nil, err
	}

	data, err := Retrieve(ns, req.GetId(), req.GetKey())
	if err != nil {
		return nil, accessError("Retrieve", ns, req.GetId(), err)
	}

	return &kvv1.RetrieveResponse{Data: data}, nil
}

func (s *kvServer) Update(
	ctx context.Context,
	req *kvv1.UpdateRequest,
) (*emptypb.Empty, error) {
	ns, err := namespace(ctx)
	if err != nil {
		return nil, err
	}

	err = Update(ns, req.GetId(), req.GetKey(), req.GetData())
	if err != nil {
		return nil, accessError("Update", ns, req.GetId(), err)
	}

	return &emptypb.Empty{}, nil
}

func (s *kvServer) Delete(
	ctx context.Context,
	req *kvv1.DeleteRequest,
) (*emptypb.Empty, error) {
	ns, err := namespace(ctx)
	if err != nil {
		return nil, err
	}

	err = Delete(ns, req.GetId(), req.GetKey())
	if err != nil {
		return nil, accessError("Delete", ns, req.GetId(), err)
	}

	return &emptypb.Empty{}, nil
}
