up:
	docker-compose up --build -d
	go run cmd/main.go

down:
	docker-compose down 

nuke:
	docker-compose down -v