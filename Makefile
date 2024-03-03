include .env

docker-up:
	docker compose --env-file=.env --project-name=${PROJECT_NAME} up -d

docker-down:
	docker compose --project-name=${PROJECT_NAME} down

docker-restart: docker-down docker-up

migrate:
	./goose -dir db/migrations/postgres postgres "host=localhost user=user password=password dbname=postgres sslmode=disable" up
	./goose -dir db/migrations/clickhouse clickhouse "http://localhost:8123" up