# <p align="center">MacDeploy</p>

Looking to automate MacBook deployments? No MDM? No JAMF? No problem! 

*MacDeploy* is a light-weight server and automation framework used to deploy MacBooks with minimal manual interactions needed.
It features:
- Automation of user creation, SecureToken handling, package installations, FileVault key handling, logging, and more.
- Automated storage to the server of the FileVault key upon generation.
- Password policies similar to Windows (no expiration dates).
- A lightweight file server to facilitate client-server communication and file distributing.
- Easy deployment of the server and scripts anywhere, on any device.
- Uses a self-signed certificate to enable HTTPS for encryption.
- Customizable YAML configuration.

It is powered by Go, Python, Bash, and Docker.

***DISCLAIMER***: This server was built with the intention to be running on a *secure, private network*.
There is no additional security implemented to handle a public facing server.

# Table of Contents

- [Getting Started](#getting-started)
  - [Server Prerequisites](#server-prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
  - [Deployment](#deployment) 
  - [Deploy Flags](#deploy-flags)
  - [Zipping](#zipping)
  - [Logging](#logging)
- [YAML Configuration File](#yaml-configuration-file)
  - [YAML Reference](#yaml-reference)
- [Limitations and Security](#limitations-and-security)
  - [Security](#security)
  - [curl](#curl)
- [License](#license)

# Getting Started

## Server Prerequisites

The server must **run on a macOS or Linux** operating system.
Windows is not supported (WSL is fine).

Below are the tools and software required on the server before starting the deployment process.
- `Go`
- `docker`
- `docker compose`
- `git`
- `zip`
- `unzip`

`zip`, `unzip`, and `curl` are required on the clients. MacBook devices have these installed by default.

## Installation

One-liner for installation.
```shell
git clone REPLACE_ME_HERE && \
cd macos-deployment && \
bash scripts/docker_build.sh && bash scripts/go_zip.sh && \
docker compose create && docker compose start
```

Clone the repository and change the working directory: 
```shell
git clone REPLACE_ME_HERE && cd macos-deployment
```

Create the Docker images, the deployment binary, and the ZIP file:
```shell
bash scripts/docker_build.sh && bash scripts/go_zip.sh
```

Create and run the containers:
```shell
docker compose create && docker compose start
```

Alternatively, you can run `docker compose up` to create and start the containers.

**IMPORTANT**: Before usage, it is required to configure the *YAML* configuration file in order
for the deployment process to work.
Click [here](#yaml-configuration-file) to get started on the YAML configuration file.

# Usage

The MacBook devices must be connected to the same network as the server.

You must have a **YAML configuration file** set up prior to deploying, otherwise there will be issues running the
deployment process.
- Run `bash scripts/go_zip.sh` after configuring to setup the ZIP file for deployment.

There are ***two binaries generated*** when running the script: `deploy-arm.bin` and `deploy-x86_64.bin`.
- `deploy-arm.bin` is used on *Apple Silicon* MacBooks.
- `deploy-x86_64.bin` is used on *Intel* MacBooks.

As Intel MacBooks are being phased out, the most often use case would be `deploy-arm.bin`.

The binary has three flags that can be used, `-a`, `--exclude`, and `--include`.

## Deployment

Access the ZIP file endpoint to obtain the deployment zip file. 
Replace the `<YOUR_DOMAIN>` with your domain (by default the server's private IP): 
```shell
curl https://<YOUR_DOMAIN>:5000/api/packages/deploy.zip --insecure -o deploy.zip
```

Unzip the contents of the ZIP file: 
```shell
unzip deploy.zip
```
This unzips the `dist` directory into the current working directory, which contains all the files for the clients.

Run the binary to start the deployment process (`deploy-x86_64.bin` if Intel is required):
```shell
./dist/deploy-arm.bin
```

**DISCLAIMER**: It is not possible to fully automate macOS deployments due to Apple's policies.
Some processes will still require manual interactions.

## Deploy Flags

| Flag | Usage | Example |
| ---- | ---- | ---- |
| `-a` | Gives admin to the user. If `ignore_admin` is true for a user, this is ignored. | `./deploy-arm.bin -a` |
| `--exclude <file>` | Excludes a package from installation. | `./deploy-arm.bin --exclude "Chrome"` |
| `--include "<file/file_name_1/file_name_2>"` | Include a package to install. | `./deploy-arm.bin --include "zoomUSInstaller/zoom.us"` |

`--exclude <file>` is used to prevent packages defined in the YAML config file from being installed on device.

`--include <file>` is used to *install packages in the package folder*, but *not defined in the YAML config*. 
This is intended to be used to keep the defined packages in the YAML config as default applications 
to install on all devices.
- Strings after the delimiter (`/`) are used as an installation check in the given search directories. 
- If the delimiter is omitted, then the deployment will attempt to install without checking for previous
installs.

## Zipping

Upon generation, the ZIP file is placed inside the root directory.

The files that are *zipped* for deployment are located inside the `dist` directory.
To add files that is used on the client, place the files inside the `dist` directory.
Directory structure does not matter .

It is *important to update the ZIP file after any changes*, running `bash scripts/go_zip.sh` will update it properly.
Additionally, there is a Docker container (`cronner`) that periodically runs a cron job every 10 minutes by
accessing the token-based endpoint to update the ZIP file.

## Logging

The log file is created in the temporary folder `/tmp` by default on the client. 
The log name follows the format: `2006-01-02T-15-04-05.<SERIAL>.log`.
- The permissions are 0600 by default.

If the **log file fails to send to the server**, ensure to save this log file if the FileVault key was generated.
**Any subsequent reruns** will not include the key in new logs if FileVault was successfully enabled. 
- **Disabling** FileVault is recommended to ensure the key is generated in the new log.

Logs from the client and server are found in the `logs` folder in the root directory. The server logs are located in the 
subdirectory `server-logs`.

# YAML Configuration File

```yaml
# sample config
accounts:
  account_one:
    user_name: "EXAMPLE.NAME"
    password: "PASSWORD"
    ignore_admin: true
  account_two:
    password: "PASSWORD"
    change_password: true
admin: # REQUIRED
  user_name: "USERNAME"
  password: "PASSWORD"
packages:
  pkg_one_name:
    - "pkg_one_app_name_one"
    - "pkg_one_app_name_two"
  pkg_two_name:
    - "pkg_two_app_name_one"
  pkg_three_name:
    -
search_directories:
  - "/search_dir_one" 
  - "/search_dir_two" 
server_host: "https://127.0.0.1:5000" # REQUIRED
filevault: false
firewall: false
always_cleanup: false
```

The YAML configuration file is used for configuration of the binary.
The ***YAML should be configured prior to building the binary***.
It is <u>embedded into the binary</u>, and *any changes will require an update to the binary* 
via `bash scripts/go_zip.sh`.

Only the fields marked as **required** are needed. If a non-required field is omitted, then that section
will be skipped during the deployment process.

A sample config can be found in the repository or by looking at the top of this section.

***IMPORTANT***: If special characters are used inside a string field, **it must be quoted**.

## YAML Reference

`accounts`: Creates the default users on the client device.
- `account_name`: Groups info for a user, it can be named anything but *must be unique*.
    - `user_name`: The username of the user, this value *must be unique*. If omitted, the binary
    will prompt for an input to create the user.
    - `password` (REQUIRED): The password of the user used to login. Required if a user is being made.
    - `change_password`: Prompts for a password reset upon login of the account.
    It is **highly recommended** to enable this for users with default passwords.
    - `ignore_admin`: Ignores creating the user as admin if the `-a` flag is used. This is only used for
    default accounts in the YAML config.

`admin` (REQUIRED): The user info for the main admin/first account of the device. Used for automation.
  - `user_name` (REQUIRED): The username of the admin account.
  - `password` (REQUIRED): The password of the admin account.

`packages`: Package file names that are being installed from the `pkg-files` directory on the client device.
  - `package_name`: The package file. The `dist` folder will be read to find any files ending in `.pkg`. The file names
  can be an exact match or fuzzy matched, but it is recommended to put the full package name in, e.g. `teamviewer.pkg`. 
  Quotes are required if there are spaces in the package name, e.g. `"Office 2016.pkg"`.
    - `installed_file_name`: This is the file that is installed when a `.pkg` is successfully installed. It can
    be either a `.app` file or a directory. It is *not case sensitive*, and should match the file name in the
    given search directory. For example, `Microsoft Word.app` is searched by `"microsoft word"` or `"Microsoft Word.app"`.
    Can be omitted in the config but must pass an empty value `-` or `- ""`. Spaces or special characters requires quotes.

`search_directories`: Array of paths that are used for `installed_file_name` to search for applications.

`server_host` (REQUIRED): The URL of the server, this is required for communications and must be in HTTPS. 
By default it is the private IP of the server on port 5000. 

`filevault`: Enable or disable FileVault activation in the deployment.

`firewall`: Enable or disable Firewall activation in the deployment.

`always_cleanup`: Enable or disable the removal of the deployment files from the device. If the server is not reachable, 
then the cleanup will not occur regardless of value.

# Limitations and Security

## Security

**Do not run this** publicly, which will cause the endpoints to be accessible to everyone.

The deployment process is expected to be ran ***on a private network***, and therefore its security is at a level where it 
protects the bare minimum. 

The endpoints do not have proper safeguards in place, although only the updating ZIP endpoint has basic authentication.
Exposing these endpoints can cause unintended consequences.

## curl

The `curl` command uses the `--insecure` option to bypass the verify check (used in the Go code too).
Although this is not recommended, it is used in this case due to the nature of device deployment- 
or in other words the devices are fully wiped prior to deployment.

# License

MacDeploy is available under the [MIT License](https://opensource.org/license/MIT).