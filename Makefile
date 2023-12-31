


cli:
	go build -o ./build/cli cmd/cli/main.go
	chmod +x ./build/cli

db:
	go build -o ./build/db cmd/db/db.go
	chmod +x ./build/db

manager:
	go build -o ./build/manager cmd/manager/main.go
	chmod +x ./build/manager
