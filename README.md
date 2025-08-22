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
  pkg_three_name:
    -
search_directories:
  - "/search_dir_one" 
  - "/search_dir_two" 
server_ip: "http://127.0.0.1:5000" # REQUIRED
firewall: false
```

Some of the script functionality *will be skipped* if no value is given.
- For example, if no `packages` are given, then no attempts are made to install any packages.

### YAML Configuration Reference

`server_ip` is a ***required*** field. This represents the server's domain, and how the communication
occurs with the client.
- HTTPS should be used here if the server container is used, *by default it generates a self-signed cert*.

`firewall` is an optional boolean field. If `true`, then Firewall will be enabled on the device.

#### Accounts

The *Accounts* section represents the accounts to be made on the device.

It consists of two levels:
1. **Account Names**: This is used to group up the third level, naming each key is preference 
but each name **must be unique**.
2. **Account Info**: Consists of three keys: `user_name` (string), `password` (string), and 
`ignore_admin` (bool).

The **account info** allows the script to create the user into the system. 

The only **required** option is `password` key. 

If the `user_name` field is omitted, the script will *prompt for an input* to create the user.
- To ensure macOS standards, the naming should consist of alphanumeric characters, periods, spaces, and
dashes.

The `ignore_admin` field **does not create the user as admin**. It should be used if the `-a`
flag for admin is used. This will prevent all `false` value users from gaining admin rights.

#### Admin

The ***Admin*** section is *required*, this is the user info of the main admin account of the macOS device,
i.e. the very first user made for the device.

Its keys are the same as the *account info* section in the *Accounts* section, but `ignore_admin` is not
used here.

`user_name` and `password` are both required. This is used to *enable FileVault* on the device.

#### Packages

The *Packages* section is the packages to be installed on the device. This installs on ***all devices by
default*** *unless* one of the packages defined is excluded by `--exclude <file>`.

It consists of two levels:
1. **Package Name**: The matching package name of the `*.pkg` file located in the `pkg-files` directory.
It is *case insensitive*, but must match the name. **Do not include** `.pkg` with the name.
2. **Installed Name**: An array of strings that is the installed file name of the package. 
This is used together with the *Search Directores* section, used to check if a file is already installed. 
It is ***case sensitive***, *do not include extensions* with the value.

If **Installed Name** is *unable to find the name* or *an empty string is given*, then it will assume
the package has not been installed and attempt to install it.
- This can be omitted and assigned an empty string to always attempt an installation.

#### Search Directories

The *Search Directories* section are an array of strings that represents the path of directories to search
for the **Installed Name** strings of the *Packages* section.

If found, then the installation attempt of the package will be skipped. 

This can be omitted to always attempt an installation.

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
This is intended to be used to separate the packages defined in the YAML config as 
default applications to install on all devices, and allow certain devices to have different installations.
- The delimiter `/` indicates that string past the first one (which is the `*.pkg` name) are values that the
`<file>.app` contains. It is used to indicate whether or not the package is installed.
- If the delimiter is omitted, then the deployment will attempt to install every run without checking.

## Action Runner

This is an optional feature and is not required to be used. 
It is safe to skip this section, it is for users who are looking to integrate a CI/CD pipeline.

There is an included action runner Docker container for the server, called `gopipe`.
This requires two action runners, one for the container and one one a spare macOS (or another work around).

By default the action runner build is not built with `docker_build.sh`, and is enabled by including
the flag argument `--action`. 
Additional checks for `.github/workflows` or the `*.yml` in the repository if the flag is used.

## Logging

The log file is created in the temporary folder `/tmp` by default. 
The log name follows the format: `%m-%dT-%H-%M-%S.<SERIAL>.log`.