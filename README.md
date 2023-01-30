B2B GRPC Chat

Chat supports:
1) One-to-one communication
2) One-to-many communication (group channels)

Server:

To run server chat application:
```cgo
make docker-up
```

Client:

```cgo
go run -o b2b-client ./cmd/b2b-client/main.go

./b2b-client -user <username> -port <server_port>
```

Once client is run, menu appears:
```shell
Use the arrow keys to navigate: ↓ ↑ → ←
? choose an action:
  > Create chat group
    Join chat group
    Join chat group
    Leave chat group
↓   Get list of channels
```
By navigating through menu, will be sent desired grpc request