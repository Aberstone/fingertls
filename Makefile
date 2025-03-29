all:
	go build -o build/mitm cmd/mitm/main.go
	go build -o build/gen-ca cmd/generate-ca/main.go