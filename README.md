# About macDes

***macDES*** is a **mac**OS **De**ployment **S**erver built to automate user creation, package installs, and administrative actions for macOS devices.

It functions as a mini-MDM and is used for people who do not have access to any MDM software.

It is built with Go, Python, and Bash, supported by Docker.

# Getting Started

## Prerequisites

The server must be **ran on a macOS or Linux** operating system.
Windows is not supported.

Below are the tools and software required on the server in before beginning the deployment process.
- `Go`
- `docker`
- `docker compose`
- `git`
- `zip`
- `unzip`

`zip`, `unzip`, and `curl` are required on the clients,
however macOS devices have these installed by default.

## Action Runner

There is an included action runner Docker container for the server, called `gopipe`.
However this requires your own configuration to use. 
Since this is also on a private network, a spare macOS is required in order to utilize a second action runner.

By default the action runner build is not included with 
`docker_build.sh`, but to include it add the flag `--action`. 

If you do not have a `.github/workflows` directory 
or an `actions.yml` located in the repository, 
then this will fail to run.

Otherwise, manual `go run build` and `git pull` are required if updates are made to the repository.

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

The server is expected to be ran on a ***private network*** 
and <u>*should not be exposed to the Internet*</u>.

## Clients

The client devices must be on the same LAN, VLAN, subnet, etc... as the server. In other words, on the same network.

