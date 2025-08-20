# About macDES

***macDES*** is a **mac**OS **De**ployment File **S**erver built to automate user creation, package installs, and administrative actions for macOS devices.

It functions as a mini-MDM and is used for people who do not have access to any MDM software.

It is built with Go, Python, and Bash, supported by Docker.

***Security warning***: This server was built with the intention to be running on a *secure, private network*.
Although it has HTTPS encryption with a self-signed cert, it is a basic implementation and there are no
other additional security measures in place.

The **configuration YAML file is included in the ZIP file** and placed on the client device during deployment.
By default this is removed after the script.
This contains sensitive information, ensure its removal after the script or if it fails.

# Getting Started

## Server Prerequisites

The server must **run on a macOS or Linux** operating system.
Windows is not supported.

Below are the tools and software required on the server in before beginning the deployment process.
- `Go`
- `docker`
- `docker compose`
- `git`
- `zip`
- `unzip`

`zip`, `unzip`, and `curl` are required on the clients, however macOS devices have these installed by default.

## Logging

The log file is created in the temporary folder `/tmp` by default. 
The log name follows the format: `%m-%dT-%H-%M-%S.<SERIAL>.log`.

## Configuration YAML

It is important to configure the YAML configuration file prior to starting the deployment process.

There is a sample configuration file with all the available options for the deployment, and also below.

```yaml
accounts:
  help:
    user_name: "EXAMPLE.NAME"
    password: "PASSWORD"
    ignore_admin: true
  client:
    password: "PASSWORD"
admin: # REQUIRED
  user_name: "USERNAME"
  password: "PASSWORD"
packages:
  pkg_one_name:
    - "pkg_one_folder_name_one"
    - "pkg_one_folder_name_two"
  pkg_two_name:
    - "pkg_two_folder_name_one"
search_directories:
  - "/search_dir_one" 
  - "/search_dir_two" 
server_ip: "http://127.0.0.1:5000" # REQUIRED
file_vault: false
firewall: false
```

There is only ***two required options in this configuration YAML***:
1. `admin`: The credentials to the main/first account of the macOS.
2. `server_ip`: The domain/IP that the server is hosted on. 

The otehr options are not required, and if these have no options then the default value will be used in its place. 

Some of the script functionality *will be skipped* if no value is given.
- For example, if no `packages` are given, then there will not be an attempt to install any packages.

The deployment will proceed as normal for most functions even on success or fail.

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

## macOS Deployment

The macOS devices must be connected to the same network as the server.
The server must also be reachable, for example via `ping`.

Run the command below, replacing `<YOUR_DOMAIN>` with your domain (by default the server's private IP).
```shell
curl https://<YOUR_IP_HERE>:5000/api/packages/deploy.zip --insecure -o deploy.zip && \
unzip deploy.zip && \
./deploy.bin
```

1. Access the ZIP file endpoint to obtain the deployment zip file. Replace the `<YOUR_DOMAIN>`
with your domain (by default the server's private IP). 
```shell
curl https://<YOUR_DOMAIN>:5000/api/packages/deploy.zip --insecure -o deploy.zip
```

2. Unzip the contents of the ZIP file to the home directory of the client: `unzip deploy.zip`.

3. Run `./deploy.bin` to start the deployment process.

**NOTE**: It is not possible to fully automate macOS deployments due to Apple's policies.

## Action Runner

This is an optional feature and is not required to be used. 
It is safe to skip this section, it is for users who are looking to integrate a CI/CD pipeline.

There is an included action runner Docker container for the server, called `gopipe`.
However this requires two self-hosted runner to use. 
Since this is intended for a private netowrk, a spare macOS is required in order to utilize a second action runner.

By default the action runner build is not built with `docker_build.sh`, but can be enabled by including the flag argument `--action`. 
There will be additional checks for `.github/workflows` or `actions.yml` in the repository 
if the flag is given.

WIPanual interactions are required at different stages.