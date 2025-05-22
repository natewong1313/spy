include .env

create-migration:
	migrate create -ext=sql -dir=internal/db/migrations -seq init

migrate-up:
	migrate -path=internal/db/migrations -database="${DATABASE_URL}" -verbose up

migrate-down:
	migrate -path=internal/db/migrations -database="${DATABASE_URL}" -verbose down
