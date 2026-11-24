package main

import (
	"context"
	kvv1 "yoti/api/kv/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

type kvServer struct {
	kvv1.UnimplementedKeyValueStoreServer
}

func (s *kvServer) Store(
	ctx context.Context,
	req *kvv1.StoreRequest,
) (*kvv1.StoreResponse, error) {
	key, err := Store(req.GetId(), req.GetData())
	if err != nil {
		return nil, err
	}

	return &kvv1.StoreResponse{Key: key}, nil
}

func (s *kvServer) Retrieve(
	ctx context.Context,
	req *kvv1.RetrieveRequest,
) (*kvv1.RetrieveResponse, error) {
	data, err := Retrieve(req.GetId(), req.GetKey())
	if err != nil {
		return nil, err
	}

	return &kvv1.RetrieveResponse{Data: data}, nil
}

func (s *kvServer) Update(
	ctx context.Context,
	req *kvv1.UpdateRequest,
) (*emptypb.Empty, error) {
	err := Update(req.GetId(), req.GetKey(), req.GetData())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *kvServer) Delete(
	ctx context.Context,
	req *kvv1.DeleteRequest,
) (*emptypb.Empty, error) {
	err := Delete(req.GetId(), req.GetKey())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
