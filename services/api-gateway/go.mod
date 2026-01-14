module github.com/nahouchimohamed/orbital-ops-console/services/api-gateway

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/cors v1.2.1
	github.com/nahouchimohamed/orbital-ops-console/services/orbit-service v0.0.0
	google.golang.org/grpc v1.67.1
)

require (
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)

replace github.com/nahouchimohamed/orbital-ops-console/services/orbit-service => ../orbit-service
