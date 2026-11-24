# Encrypted key-value store

Note: about 70% of this is hand written, including all of the main code and tests. The exceptions are the gRPC stubs and similar generated files.

Note: I had to take a few breaks in between as I had multiple interviews, and I hope that is okay. 

## Decisions

- I chose gRPC because this is a service-to-service API with a required Go client package. The schema gives both sides a shared contract, and code generation supplies the client. I committed the generated files so reviewers don't need protoc to run it.
- Global store and mutex, no MemoryStore struct: excessive for now. Revisit if tests need isolation.
- No dependency injection: the gRPC server calls the package-level store functions directly. Tests share the global store and stay isolated with unique IDs and cleanups. Revisit with a store struct injected into the server if a second store or per-test instances are needed.
- The store takes a caller and keys entries as caller/id. This keeps one service from reading or overwriting another service's entries even when IDs overlap. gRPC is transport only.
- The caller is the x-service-name header. Ideally callers authenticate and the service infers the caller from headers or a client cert (mTLS).
- Retrieve, Update and Delete return NotFound: entry unavailable for both missing entries and wrong keys. The real reason is logged.

-  the server generates each key, sees the plaintext and receives the key on every call. Encryption protects data at rest, not a compromised server or the plaintext local gRPC transport.
- Plaintext gRPC is acceptable for this exercise because the server binds to (127.0.0.1) and only handles demo data. Any remote use requires TLS.

## Run

Requires the Go version in go.mod.

```bash
go run .
```

In another terminal:

```bash
go run ./cmd/example
```

Expected output: hello, updated, deleted. The server listens on 127.0.0.1:50051 and loses data on restart.

## Verify

```bash
go test -race ./...
go vet ./...
```

## Regenerate protobuf

Requires protoc, protoc-gen-go and protoc-gen-go-grpc.

```bash
go generate ./api/...
```
