# About macDes

***macDES*** is a **mac**OS **De**ployment **S**erver built to automate user creation, package installs, and administrative actions for macOS devices.

It functions as a mini-MDM and is used for people who do not have access to any MDM software.

It is built with Go, Python, and Bash, supported by Docker.

# Getting Started

## Prerequisites

The server must be **ran on a macOS or Linux** operating system.
Windows is not supported.

Below are the required languages and tools needed on the server in order to start the deployment process.

Required languages:
- `Go`

<br/>

Tools:
- `zip`
- `unzip`
- `curl`
- `docker compose`
- `docker`
- `git`

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

