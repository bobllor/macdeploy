<div align="center">
  <img src="https://www.svgrepo.com/show/528339/laptop-3.svg" 
  height="130" width="125">

  <h3 align="center">
    MacDeploy
  </h3>

  <p align="center">
    An IT solution for MacBook deployment.
  </p>
</div>

## About the Project

Looking to automate MacBook deployments? No problem! 

*MacDeploy* is a light-weight server and CLI automation tool used to deploy MacBooks with minimal manual interactions 
needed. It features:
- Automation of package installation, DMG extraction, user creation, admin tools, script executions, and more.
- A lightweight file server to facilitate client-server communication and file distributing.
- Automated storage of the FileVault key to the server when generation.
- Password policies for user created accounts.
- Uses HTTPS for encrypted communications.
- Customizable YAML configuration.

***DISCLAIMER***: MacDeploy is provided only to automate MacBook deployment. Securing and management of the 
hardware itself is the responsibility of the user.

The HTTPS file server was built with the intention to be running on a *secure, private network*.
There is *no additional security* implemented to handle a public facing server.

### Powered By

[![Go](https://img.shields.io/badge/Go-%2300ADD8.svg?&logo=go&logoColor=white)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3776AB?logo=python&logoColor=fff)](https://www.python.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](https://www.docker.com/)
![Bash](https://img.shields.io/badge/Bash-4EAA25?logo=gnubash&logoColor=fff)

## Table of Contents

- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation and Setup](#installation-and-setup)
- [Usage](#usage)
  - [Deployment](#deployment) 
  - [Deployment options](#deployment-options)
  - [Updating](#updating)
- [YAML Configuration File](#yaml-configuration-file)
  - [YAML Reference](#yaml-reference)
- [Server and Deployment](#server-and-deployment)
  - [Zipping](#zipping)
  - [FileVault](#filevault)
  - [Logging](#logging)
- [Other](#Other)
  - [Security](#security)
  - [Other Features](#other-features)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Getting Started

### Prerequisites

The server must **run on a macOS, Unix, or Linux** operating system.
Windows is not supported (WSL is fine).

Below are the tools and software required on the server before starting the deployment process.
- `Go`
- `docker`
- `docker compose`
- `git`
- `zip`

`zip`, `unzip`, and `curl` are required on the clients. MacBook devices have these installed by default.

### Installation and Setup

For the latest release you can use `git checkout $(git describe --tags $(git rev-list --tags --max-count=1))`.
- If you need a specific version: `git checkout <TAG_VERSION>`.

Ensure to run `bash build.sh`, which creates the required directories and creates the containers prior to setup.
This is required to prevent permission issues with bind mounts.

This creates the server setup, but *not the deployment binaries*:
```shell
git clone REPLACE_ME_HERE && \
cd macos-deployment && \
git checkout $(git describe --tags $(git rev-list --tags --max-count=1)) && \
bash build.sh
```

To start the containers and server:
```shell
docker compose up -d
```

About the script `build.sh`:
- It has an optional flag `-z` to generate the ZIP file. This *requires the YAML configuration file* to be present.
- It *only creates the Docker containers*, it does not start them.
- It removes the Docker volume, which is only used for the server code.

**IMPORTANT**: Before usage, the *YAML* configuration is required to be created in order for the deployment binaries to work 
properly. Click [here](#yaml-configuration) to get started.

`go_zip.sh` located in the `scripts` folder generates the *deployment binaries* used for the deployments and creates the ZIP file.
In order for the binaries to work, *the YAML configuration must be configured prior* to executing `go_zip.sh`.

***Two binaries are generated*** in the `dist` folder: 
`macdeploy` and `x86_64-macdeploy`.
- `macdeploy` is used for *Apple Silicon* MacBooks.
- `x86_64-macdeploy` is used for *Intel* MacBooks.

As Intel MacBooks are being phased out, the following examples will be using `macdeploy`, however the 
commands will be the same if the *Intel* version is used.

## Usage

The client devices *must be connected to the same network* as the server.
All commands will be used on the terminal of the device.

### Deployment

Replace `SERVER_IP_DOMAIN` with the IP or domain name to the server.
```shell
curl https://SERVER_IP_DOMAIN:5000/api/packages/deploy.zip -o deploy.zip --insecure && \
unzip deploy.zip
```
This downloads the ZIP file from the server and unzips the contents of the `dist` directory into the 
working directory of the client, by default it is where the terminal opens: the home directory.

To start the deployment process (no flags):
```shell
./dist/macdeploy
```

The binary supports *flags* options which can be found [here](#deployment-options).

### Deployment Options

| Options | Description |
| ---- | ---- |
| `--admin`, `-a` | Gives admin to a created user. If `ignore_admin` is true in the YAML, this is ignored. |
| `--skip-local`, `-s` | Skips the creation of the local user account, if configured in the YAML. |
| `--create-local`, `-c` | Enables the local user account creation process. |
| `--remove-files` | Removes deployment files upon successful completion. |
| `--verbose`, `-v` | Output info level logging to the terminal. |
| `--debug` | Output debug level logging to the terminal, this includes info level logging. |
| `--no-send` | Prevents the log from being sent to the server. |
| `--pwlist "/path/to/plist"` | Apply password policies using a plist path. |
| `--exclude "file"` | Excludes a package from installation. |
| `--include "<file/installed_file_1/installed_file_2>"` | Include a package to install. |

The `installed_file_1/installed_file_2` arguments of the`--include` flag is the installed file name, 
i.e. the files on the device after installing the package.
- For example, if `Chrome.pkg` is installed a file will be created named `Google Chrome.app` 
found inside `/Applications`. 
- To install the package and check if it is already installed: 
`--include "chrome.pkg/Google Chrome"`.

### Updating

```shell
git fetch origin && \
git checkout $(git describe --tags $(git rev-list --tags --max-count=1)) && \
docker down -v && \
docker compose build && docker compose up -d
```

Alternatively, there is a script for updating in the scripts folder. 
However, this is *only available for release v1.2.3 and above*:
```shell
bash scripts/update.sh
``` 

## YAML Configuration

The YAML configuration file is used for configuration of the binary, and must be 
***configured prior to building the binary***.
It is *embedded into the binary*, meaning any new updates will require a new binary to be generated
via `bash scripts/go_zip.sh`. 

The YAML file *must be named `config`* and can end in `.yaml`, `.yml`, `.YAML`, or `.YML`. 

A sample file can be found in the repository.

### YAML Reference

`accounts`: Creates the users on the client device. Can be omitted if no users need to be created.
- `account_name`: It can be named anything but *must be unique*.
    - `username`: This value *must be unique*. If omitted, an input prompt for a
    username will be displayed.
    - `password`: Can be omitted, a password input prompt will appear.
    - `apply_policy`: Apply password policies to the user from the given values.
    - `ignore_admin`: Ignores giving admin to the user if the *admin flag* is used. 
    This applies only for accounts defined in the YAML config.

`admin`: A user info map for the main account. 
  - `username`: It must match the same internal username during the initial account creation.
  For example, if the display name is `Admin User` the *internal username* is `adminuser`. Can be omitted.
  - `password`: It can be omitted, but will prompt for the password. If it fails to validate then the program will exit.
  - `apply_policy`: Apply password policies to the admin account. Must be `true` if the admin account requires
  policies applied.

`packages`: Packages that are being installed from the distribution directory.
  - `package_name`: The package name ending in `.pkg`. Used to execute scripts found in the distribution directory. 
  It is *case insensitive* and *looks for a name match*. 
    - `installed_file_name`: The folder added after installing a `.pkg` file. It is not case sensitive, 
    and matches the file name in the search directories. 
    Example: `Microsoft Word.app` can be found by `"microsoft word"` or `"Word.app"`.
    If omitted then the package will be installed on every attempt.

`scripts`: Scripts to be executed on the client device. It executes in three deployment stages: before, during, and after.
The script files *must have the correct permission* prior to being compressed into the ZIP file. 
Each section is an array of script names, it is *case insensitive* and *looks for a name match*.
  - `pre`: Scripts to be executed before deployment.
  - `inter`: Scripts to be executed during deployment, this is executed after installation of packages. 
  - `post`: Scripts to be executed after deployment.
  
`policies`: A map of password policies applied to chosen accounts in the config.
  - `reuse_password`: Determines if the user can reuse a password. Ranges from 0 to 15, with 1 being the default. 
  - `require_alpha`: Requires the password to have at least one letter.
  - `require_numeric`: Requires the password to have at least one number.
  - `min_characters`: Minimum characters for the password.
  - `max_characters`: Maxmimum characters for the password. 
  - `change_on_login`: Requires a password change before logging in. This is **required** in order 
  to apply the password policies.

`search_directories`: Array of paths that are used for `installed_file_name` to search for applications.

`server_host`: The IP or domain of the server, used for client-server communication. *This is required* for the
deployment to work.

`log`: The path to the folder of the log files. By default, it is stored at `~/logs/macdeploy` if omitted or an error
occurs.

`filevault`: Enable or disable FileVault activation in the deployment.

`firewall`: Enable or disable Firewall activation in the deployment.

## Server and Deployment

### Gunicorn Configuration

In the root directory is the `gunicorn.conf.py` file, the configuration for the server.
By default there are three values defined: the preload status, log level, and number of workers.
- It is recommended to keep the preload status enabled, ensuring the thread locks work during the ZIP file update.
- The worker amounts defaults to four. Additional workers beyond that are unnecessary, depending on the amount of device
deployments.

These can be modified as needed, based on the [Gunicorn's documentation](https://docs.gunicorn.org/en/stable/settings.html).


Any changes to the configuration *will require a reset of the server*, but does not require the containers to be rebuilt.

### Zipping

When generated, the ZIP file is placed inside the `zip-build` directory in the project's root directory.

The *deployment binary reads all files* in the `dist` directory.
- This is the folder for all packages, scripts, DMGs, etc. that need to be on the client device.

Directory structure does not matter in `dist`. It is recommended to create separate directories in order
to prevent naming conflicts. Some *packages may require config files* alongside the installer, which the folders
ensure separation.
- During deployment, all files are obtained recursively.

It is *important to update the ZIP file after any changes*, running `bash scripts/go_zip.sh` will generate a new ZIP file
and deployment binaries.

### Server Zipping

The server contains an endpoint used for updating and creating (if missing) the ZIP file. 
This is intended for use with the `zip-updater` container in Docker.
- The ZIP file update/creation *works similar to* the `zip` command on Linux.

It *does not create the binary*, the binary is generated via `bash scripts/go_zip.sh`.

The container accesses the endpoint *every 2 hours*, which can be modified inside the `compose.yml` file:
```yaml
zip-updater:
  build:
    args:
      TIMER: 2h # valid arguments: [0-9]+[HMShms]
```

Expected arguments are numbers followed by H, M, or S; hours, minutes, and seconds respectively.
The script *uses seconds for the timer*, with H and M converting into seconds. If S is used, the script
will use the argument as-is.
- The timer is automatically converted to seconds, e.g. `3H` -> `10800`.
- The argument is case insensitive.
- In the event of a parsing failure, it will always default back to 2 hours or 7200 seconds.

### FileVault

FileVault is recommended to be turned on for security purposes. Nearly all the processes in the deployment expects
FileVault to be enabled, otherwise it will not work as intended.

Upon successful execution, the key will be generated. This key *is not logged on the client* but will be 
outputted onto the terminal.

The key is sent over HTTPS to the server for storage. It can be found in the `keys` directory, created as a subfolder
under a parent folder that is the *serial tag* of the device.
If this process fails in any way, the key *must be saved* manually.
- Any created account (not the root/main admin) is removed. 
- Example of a folder structure: `./SERIAL_TAG/1234-5678-9000-ABCD`.

### Logging

The log output location can be defined inside the YAML configuration with the field `log`.

The value to the log path is expected to be *a directory*, and if a `.log` extension is attached to the
the file will be dropped, taking the parent directory instead.
- All parent directories are created automatically.

In the event of a failure, the default log output will be set to the client's home directory: `~/logs/macdeploy`.

The log name follows the format: `2006-01-02T-15-04-05.<SERIAL>.log`.
- The permissions are 0600 by default.

Logs from the client and server are found in the `logs` folder in the project's root directory. 
The server logs are located in the subdirectory `server-logs`.

## Other

### Supported MacBook Versions

Below is a table of confirmed MacBook versions that work.

| Version |
| ------- |
| Ventura 13.7.x |
| Sequoia 15.x |

Other versions are not confirmed yet.

### Security

**Do not run this** publicly, which will cause the endpoints to be accessible to everyone. The lacks proper
authentication and rate-limiting protections.

The deployment process is expected to run ***on a private, isolated network***. Exposure can cause unintended consequences. 

### Features

Various features expected with JAMF or other MDM tools may be missing. 

Contributions and suggestions are always welcome.

## License

MacDeploy is available under the [MIT License](https://opensource.org/license/MIT).

## Acknowledgements

Special thanks to these resources:

- [Cobra CLI](https://github.com/spf13/cobra)
- [Go YAML](https://github.com/goccy/go-yaml)
- [MD Badges](https://github.com/inttter/md-badges)
- [Flask](https://github.com/pallets/flask)
- [Gunicorn](https://github.com/benoitc/gunicorn)
- Docker
- and various other Python libraries.