help:
	@echo "Usage: make <command>\n"
	@echo "Available commands:"
	@printf "%3sbuild: Builds the Docker containers and creates the project files"
	@printf "%3sstart: Starts the server\n"
	@printf "%3sstop: Stops the server\n"
	@printf "%3sstart-test: Starts the test server\n"
	@printf "%3sstop-test: Stops the test server\n"
	@exit 0

build:
	@bash ./scripts/build.sh

start:
	@bash ./scripts/start.sh

start-test:
	@bash ./scripts/start.sh -t

stop:
	@bash ./scripts/stop.sh

stop-test:
	@bash ./scripts/stop.sh -t