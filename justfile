build:
    go build -o ./bin/player ./main.go

run: build
    ./bin/player


test:
    go test -v ./...
