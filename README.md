# About macDes

***macDES*** is a **mac**OS **De**ployment **S**erver built to automate user creation, package installs, and administrative actions for macOS devices.

It functions as a mini-MDM and is used for people who do not have access to any MDM software.

It is built with Go, Python, and Bash, supported by Docker.

# Getting Started

## Prerequisites

The server must be **ran on a macOS or Linux** operating system.
Windows is not supported.

Below are the tools and software required on the server in order to start the deployment process.

Required languages:
- `Go`

<br/>

Tools:
- `docker`
- `docker compose`
- `git`

`zip`, `unzip`, and `curl` are also used but are installed by default on macOS devices.
Others are handled by the Docker container.

## Action Runner

There is an included action runner Docker container for the server, called `gopipe`.
However this requires a user setup, as it uses a second action runners on a physical macOS device.

## Installation

1. Clone the repository:
    ```shell
    git clone https://github.com/TGSDepot/macos-deployment.git
    ```

2. Change the current directory into the newly added repository: `cd macos-deployment`.

3. Run the following commands with the scripts to initialize the files:
    ```shell
    bash scripts/docker_build.sh; \
    bash scripts/go_build.sh; \
    bash scripts/create_zip.sh
    ```

4. Create the containers using `docker compose create`.

5. Run the containers using `docker compose start`.

# Usage

The server is expected to be ran inside a ***private network***, and should not be exposed publically to the Internet.

