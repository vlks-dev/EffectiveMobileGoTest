build:
	go build -o bin/main cmd/main.go
run:
	go mod tidy
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate create -ext sql -dir migrations/pgx_migrations -format "20060102150405" create_songs_table
	go run cmd/main.go -log-level debug

compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm go build -o bin/main-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go

start: build run