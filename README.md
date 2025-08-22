# About

The macOS Deployment File Server is built to deploy macOS devices automatically without the use of MDM software.
It automates user creation, admin tools, and package installations.
It is built and powered by *Go, Python, Bash, and Docker*.

***Security warning***: This server was built with the intention to be running on a *secure, private network*.
It uses HTTPS to encrypt data with a self-signed cert. There is no additional security implemented.

# Getting Started

## Server Prerequisites

The server must **run on a macOS or Linux** operating system.
Windows is not supported.

Below are the tools and software required on the server before starting the deployment process.
- `Go`
- `docker`
- `docker compose`
- `git`
- `zip`
- `unzip`

`zip`, `unzip`, and `curl` are required on the clients.
macOS devices have these installed by default.

## YAML Configuration File

The YAML configuration file is used for **default options** of the final binary build. The deployment
on the client's device is based around the configuration.

The ***YAML should be configured prior to building the binary*** or *before the deployment process begins*.
It is <u>embed into the binary</u>, and any changes will require an update to the binary 
via `bash scripts/go_build.sh` and `bash scripts/create_zip.sh`.

There is a sample configuration file with all options in the repository and also below.

***All YAML files are expected to end in `.yml`***.

```yaml
accounts:
  account_one:
    user_name: "EXAMPLE.NAME"
    password: "PASSWORD"
    ignore_admin: true
  account_two:
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
  pkg_three_name:
    -
search_directories:
  - "/search_dir_one" 
  - "/search_dir_two" 
server_host: "http://127.0.0.1:5000" # REQUIRED
file_vault: false
firewall: false
```

Some of the script functionality *will be skipped* if no value is given.
- For example, if no `packages` are given, then no attempts are made to install any packages.

### YAML Reference

`accounts`: Section of user data that creates the user.
- `account_name`: Name of a user group, can be named anything but *must be unique*.
    - `user_name`: The username of the user, this value *must be unique*. If omitted, the binary
    will prompt for an input to create the user.
    - `password` (REQUIRED): The password of the user used to login. Required if a user is being made.
    - `ignore_admin`: Ignores creating the user as admin if the `-a` flag is used. This is only used for
    default accounts in the YAML config.

`admin` (REQUIRED): Section of user data that represents the main admin/first account of the device.
  - `user_name` (REQUIRED): The username of the admin account.
  - `password` (REQUIRED): The password of the admin account.

`packages`: Section of package file names being installed from the `pkg-files` directory.
  - `package_name`: The package file, *it is case sensitive and must have the same name*. Do not include
  the extension `.pkg`.
    - `installed_file_name`: The application or directory of the package files after installation, *it is
    case sensitive and must have the same name*. Do not include the extension `.app` if applicable. Can be
    omitted but must pass an empty value `-` or `- ""`.

`search_directories`: Section of an array of paths that are used for `installed_file_name` to search.

`server_host` (REQUIRED): The FQDN of the server, this is required for communications and must be in HTTPS. 
By default it is the private IP of the server. 

`file_vault`: Boolean used to enable or disable FileVault activation in the deployment.

`firewall`: Boolean used to enable or disable Firewall activation in the deployment.

## Installation

1. Clone the repository: `git clone REPLACE_ME_HERE`

2. Change the current directory into the newly added repository: `cd macos-deployment`.

3. Run the following commands with the scripts to initialize the files and container:
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

The `curl` command uses the `--insecure` option to bypass the verify check (also similar with the Go code).
Although this is not recommended, it is used in this case due to the nature of macOS deployment, in other words
the devices that accesses the file server are fully wiped prior to deployment.

The command below is an example one liner. It installs all packages and creates a standard user. 
Replace `<YOUR_DOMAIN>` with your domain (by default the server's private IP).
```shell
curl https://<YOUR_DOMAIN>:5000/api/packages/deploy.zip --insecure -o deploy.zip && \
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

`deploy.bin` has three flags and can be used based on the requirements of the device.

**NOTE**: It is not possible to fully automate macOS deployments due to Apple's policies.
Some processes will still require manual interactions.

## Deploy Flags

| Flag | Usage | Example |
| ---- | ---- | ---- |
| `-a` | Gives admin to the user. | `./deploy.bin -a` |
| `--exclude <file>` | Excludes a package from installation. | `./deploy.bin --exclude "Chrome"` |
| `--include <file/installed_file_1>` | Include a package to install. | `./deploy.bin --include "zoomUSInstaller/zoom.us"` |

`--exclude <file>` is used to prevent packages defined in the YAML config file from 
being installed on a device.

`--include <file>` is used to *download a package found in the package folder*, but *not in the YAML config*. 
This is intended to be used to separate the packages in the YAML config as default applications 
to install on all devices.
- Strings past the delimiter (`/`) are used as search values in the given search directories. 
These files usually end with the `.app` extension.
- If the delimiter is omitted, then the deployment will attempt to install without checking for previous
installs.

## Action Runner

This is an optional feature and is not required to be used, it is for users who are looking to integrate
a CI/CD pipeline.

There is an included action runner Docker container for the server, called `gopipe`.
This requires two action runners, one for the container and one one a spare macOS (or another work around).

By default the action runner build is not built with `docker_build.sh`, and is enabled by including
the flag argument `--action`. 
Additional checks for `.github/workflows` or the `*.yml` in the repository if the flag is used.

*NOTE*: This is not a great implementation of a CI/CD pipeline as I am using this to learn how to
implement it.

## Logging

The log file is created in the temporary folder `/tmp` by default. 
The log name follows the format: `%m-%dT-%H-%M-%S.<SERIAL>.log`.