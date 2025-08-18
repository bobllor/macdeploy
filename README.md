# About macDes

***macDES*** is a **mac**OS **De**ployment **S**erver built to automate user creation, package installs, and administrative actions for macOS devices.

It functions as a mini-MDM and is used for people who do not have access to any MDM software.

It is built with Go, Python, and Bash, supported by Docker.

# Getting Started

## Prerequisites

The server must be **ran on a macOS or Linux** operating system.

Below are the required software and tools needed on the server in order to start the deployment process.

Software:
- `Python`
- `Go`
- `Docker`
- `git`

<br/>

Tools:
- zip
- unzip
- curl

macOS includes these tools with the OS, but some Linux distros will require installation in order to use them.

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
    bash scripts/create_zip.sh;
    ```

4. Start the containers using `docker compose up`.

# Usage