.PHONY: build run test clean reset-data

build:
	go build -o bin/orchestra ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -f data/*.db data/*.db-wal data/*.db-shm data/*.db-journal

# Wipe local DB only (paths match configs/config.yaml storage.database).
# Stop the server first or rm may fail while WAL is open.
reset-data:
	@echo "Removing SQLite + WAL under ./data …"
	rm -f data/*.db data/*.db-wal data/*.db-shm data/*.db-journal
	@echo "Done. Run: make run (empty schema from migrations)."