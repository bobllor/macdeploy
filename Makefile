help:
	@echo "Usage: make <command>\n"
	@echo "Available commands:"
	@printf "%3sbuild: Builds the Docker containers and creates the project files\n"
	@printf "%3szip: Creates the Go binary and the ZIP file for deployment\n"
	@printf "%3szip-intel: Creates the Go binary for x86-64 Intel MacBooks and the ZIP file for deployment\n"
	@printf "%3sstart: Starts the server\n"
	@printf "%3sstop: Stops the server\n"
	@printf "%3sstart-test: Starts the test server\n"
	@printf "%3sstop-test: Stops the test server\n"
	@exit 0

build:
	@bash ./scripts/build.sh

zip:
	@bash ./scripts/go_zip.sh

zip-intel:
	@bash ./scripts/go_zip.sh -x

start:
	@bash ./scripts/start.sh

start-test:
	@bash ./scripts/start.sh -t

stop:
	@bash ./scripts/stop.sh

stop-test:
	@bash ./scripts/stop.sh -t