build:
	bash scripts/build.sh

clean:
	bash scripts/clean.sh

docker-build:
	docker build -t citrixadc-backup:dev-latest .

run:
	go run main.go