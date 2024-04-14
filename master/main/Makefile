.PHONY: run test docker integration_test linter

# Путь до файла main.go
MAIN_PATH=./cmd/main.go
# Путь до интеграционного теста
INTEGRATION_TEST_PATH=./cmd/test/main.go

# Команда запуска main.go
run:
	go run $(MAIN_PATH)

# Команда запуска всех тестовых файлов
test:
	go test ./...

# Команда запуска docker контейнера
docker:
	docker-compose up -d

# Команда запуска линтера Go
linter:
	golangci-lint run
# Команда запуска интеграционного теста на windows
integration_test:
	cmd /c start go run ./cmd/main.go
	timeout /t 7
	go run ./cmd/test/integral.go

# Команда запуска интеграционного теста на linux
#integration_test:
#	xterm -e "go run ./cmd/main.go" &
#	sleep 7
#	go run ./cmd/test/integral.go
#integration_test:
#	gnome-terminal -e "go run ./cmd/main.go" &
#	sleep 7
#	go run ./cmd/test/integral.go