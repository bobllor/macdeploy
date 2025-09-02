# <p align="center">MacDeploy</p>

Looking to automate MacBook deployments? No MDM? No JAMF? No problem! 

*MacDeploy* is a light-weight server and automation framework used to deploy MacBooks with minimal manual interactions needed.
It features:
- Automation of user creation, SecureToken handling, package installations, FileVault key handling, logging, and more.
- FileVault key generation automatic storage to the server.
- Password change on login similar to that of Windows.
- A lightweight file server to facilitate client-server communication and file distributing, bypassing SSH.
- Easy deployment of the server and scripts anywhere, on any device.
- Uses a self-signed certificate to enable HTTPS for encryption.
- Customizable YAML configuration.

It is powered by Go, Python, Bash, and Docker.

***Security warning***: This server was built with the intention to be running on a *secure, private network*.
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

Alternatively, you can run `docker compose` to create and start the containers.
```shell
docker compose up
```

**IMPORTANT**: Before starting the deployment process, it is required to configure the *YAML* configuration file in order
for the deployment process to work.
Click [here](#yaml-configuration-file) to get started on the YAML configuration file.

# Usage

## Deployment

The macOS devices must be connected to the same network as the server.

The files on the client device is located in the `dist` directory upon unzipping.

You must have a **YAML configuration file** set up prior to deploying, otherwise there will be issues running the
deployment process.
- Run `bash scripts/go_zip.sh` after configuring to setup the ZIP file for deployment.

There are ***two binaries generated*** when running the script: `deploy-arm.bin` and `deploy-x86_64.bin`.
- `deploy-arm.bin` is used on *Apple Silicon*/`ARM64` MacBooks.
- `deploy-x86_64.bin` is used on *Intel*/`x86_64` MacBooks.
As Intel MacBooks are being phased out, the most often use case would be `deploy-arm.bin`. In any case,
`deploy-x86_64.bin` will remain for scenarios with Intel MacBooks.

Both binaries has three flags that can be used with arguments.

The command below is an example one liner for `ARM` MacBooks. It installs all packages and creates a standard user.
Replace `<YOUR_DOMAIN>` with your domain (by default the server's private IP):
```shell
curl https://<YOUR_DOMAIN>:5000/api/packages/deploy.zip --insecure -o deploy.zip && \
unzip deploy.zip && \
./dist/deploy-arm.bin
```

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

To only unzip the ZIP file and use the `deploy` binary with flags:
```shell
curl https://<YOUR_DOMAIN>:5000/api/packages/deploy.zip --insecure -o deploy.zip && \
unzip deploy.zip
```

**DISCLAIMER**: It is not possible to fully automate macOS deployments due to Apple's policies.
Some processes will still require manual interactions.

## Deploy Flags

| Flag | Usage | Example |
| ---- | ---- | ---- |
| `-a` | Gives admin to the user. If `ignore_admin` is true for a user, this is ignored. | `./deploy-arm.bin -a` |
| `--exclude <file>` | Excludes a package from installation. | `./deploy-arm.bin --exclude "Chrome"` |
| `--include "<file/file_name_1/file_name_2>"` | Include a package to install. | `./deploy-arm.bin --include "zoomUSInstaller/zoom.us"` |

`--exclude <file>` is used to prevent packages defined in the YAML config file from 
being installed on a device.

`--include <file>` is used to *download a package found in the package folder*, but *not in the YAML config*. 
This is intended to be used to separate the packages in the YAML config as default applications 
to install on all devices.
- Strings past the delimiter (`/`) are used as search values in the given search directories. 
These files usually end with the `.app` extension.
- If the delimiter is omitted, then the deployment will attempt to install without checking for previous
installs.
- If including files with spaces or special characters, wrap them in **quotes**.

## Zipping

Upon generation, the ZIP file is placed inside the root directory.

The files that are to be *zipped* are located inside the `dist` directory. The entire directory will be zipped
and placed into the ZIP file.

To add a file that is downloaded to the client, place the file inside the `dist` directory. 
Directory structure does not affect the zipping process.

It is *important to update the ZIP file after any changes*, running `bash scripts/go_zip.sh` will update it properly.
Additionally, there is a Docker container (`cronner`) that periodically runs a cron job every 10 minutes by
accessing the token-based endpoint to update the ZIP file.

## Logging

The log file is created in the temporary folder `/tmp` by default on the client. 
The log name follows the format: `2006-01-02T-15-04-05.<SERIAL>.log`.
- The permissions are 0600 by default, but can be removed safely.

If the **log file fails to send to the server**, ensure to save this log file if the FileVault key was generated.

By default, the file removal is not activated. If it is enabled, an unresponsive server will not remove the files.
This is done to allow you to troubleshoot the server and rerun the binary to properly add the logs to the server.

The **FileVault key will not be in subsequent logs** assuming it succeeded on the first attempt. 

Logs from the client and server goes to the `logs` folder in the repository. The server logs are located in the 
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

***IMPORTANT***: If special characters are used inside a string field, **it must be quoted**.

The YAML configuration file is used for **default options** of the final binary build. The binary uses
the configuration to setup the deployment process for the clients.

The ***YAML should be configured prior to building the binary*** or *before the deployment process begins*.
It is <u>embedded into the binary</u>, and *any changes will require an update to the binary* 
via `bash scripts/go_zip.sh`.

A sample config can be found in the repository or by looking at the top of this section.

If certain values are omitted, the script functionality *will be skipped*.
- For example, if no `packages` are given, then no attempts are made to install any packages.

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
  - `package_name`: The package file, *it is case sensitive and must have the same name* as
  the `.pkg` files in the package directory. Do not include the extension `.pkg`.
    - `installed_file_name`: The application or directory of the package files after installation, *it is
    case sensitive and must have the same name* as the files as they are in the directories. 
    Do not include the extension `.app` if applicable. Can be omitted but must pass an 
    empty value `-` or `- ""`.

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