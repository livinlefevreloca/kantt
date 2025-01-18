build:
	./build.sh -a kantt

build-and-push:
	./build.sh -p -a kantt

build-no-cache:
	./build.sh -c -a kantt

build-no-cache-and-push:
	./build.sh -c -p -a kantt

build-collector:
	./build.sh -a collector

build-collector-no-cache:
	./build.sh -c -a collector

build-dashboard:
	./build.sh -a dashboard

build-dashboard-no-cache:
	./build.sh -c -a dashboard

run-local:
	docker compose up
