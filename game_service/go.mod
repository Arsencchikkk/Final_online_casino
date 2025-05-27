module github.com/Arsencchikkk/casino/game_service

go 1.24.2

require (
	github.com/Arsencchikkk/casino/proto/game v0.0.0
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.72.1
)

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/Arsencchikkk/casino/proto/game => ../proto/game
