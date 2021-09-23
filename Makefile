build:
	bash scripts/build.sh

docker-build:
	docker build -t citrixadc-backup:dev-latest .

run:
	go run main.go