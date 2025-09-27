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

Looking to automate MacBook deployments? No MDM? No JAMF? No problem! 

*MacDeploy* is a light-weight server and CLI automation tool used to deploy MacBooks with minimal manual interactions 
needed. It features:
- Automation of package installation, DMG extraction, user creation, admin tools, logging, and more.
- A lightweight file server to facilitate client-server communication and file distributing.
- Automated storage of the FileVault key to the server upon generation.
- Password policies for user created accounts.
- Portability of server deployment on any Linux or MacBook device.
- Uses a self-signed certificate to enable HTTPS for encryption.
- Customizable YAML configuration.

***DISCLAIMER***: The HTTPS file server was built with the intention to be running on a *secure, private network*.
There is no additional security implemented to handle a public facing server.

### Planned Updates

- [ ❌ ] Add additional password policies.

### Powered By

[![Go](https://img.shields.io/badge/Go-%2300ADD8.svg?&logo=go&logoColor=white)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3776AB?logo=python&logoColor=fff)](https://www.python.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](https://www.docker.com/)
![Bash](https://img.shields.io/badge/Bash-4EAA25?logo=gnubash&logoColor=fff)

## Table of Contents

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

## Getting Started

### Prerequisites

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

### Installation

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

### Usage

The MacBook devices must be connected to the same network as the server.

You must have a **YAML configuration file** set up prior to deploying, otherwise there will be issues running the
deployment process.
- Run `bash scripts/go_zip.sh` after configuring to setup the ZIP file for deployment.

There are ***two binaries generated*** when running the script: `macdeploy` and `deploy-x86_64.bin`.
- `macdeploy` is used on *Apple Silicon* MacBooks.
- `deploy-x86_64.bin` is used on *Intel* MacBooks.

As Intel MacBooks are being phased out, the most often use case would be `macdeploy`.

The binary has three flags that can be used, `-a`, `--exclude`, and `--include`.

### Deployment

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
./dist/macdeploy
```

**DISCLAIMER**: It is not possible to fully automate macOS deployments due to Apple's policies.
Some processes will still require manual interactions.

### Deploy Flags

| Flag | Usage | Example |
| ---- | ---- | ---- |
| `--admin`, `-a` | Gives admin to the user. If `ignore_admin` is true for a user, this is ignored. | `./dist/macdeploy -a` |
| `--mount` | Auto mount, unmount, and extraction of DMGs. | `./dist/macdeploy --mount` |
| `--remove-files` | Cleans up deployment files upon successful completion. | `./dist/macdeploy --remove-files` |
| `--verbose`, `-v` | Output debug logging to the terminal. | `./dist/macdeploy -v` |
| `--no-send` | Prevents the log from being sent to the server. | `./dist/macdeploy --no-send` |
| `--apply-policy` | Applies password policy to the created user. | `./dist/macdeploy --apply-policy` |
| `--exclude <file>` | Excludes a package from installation. | `./dist/macdeploy --exclude "Chrome"` |
| `--include "<file/installed_file_1/installed_file_2>"` | Include a package to install. | `./dist/macdeploy --include "zoomUSInstaller/zoom.us"` |

The `installed_file_1/installed_file_2` of the flag `--include` is the installed file name, i.e. the files
on the device after installing the package.
- For example, if `Chrome.pkg` is installed a file will be created named `Google Chrome.app` 
found inside `/Applications`. To install it and check if it is already installed: 
`--include "chrome.pkg/google chrome"`.

## Zipping

Upon generation, the ZIP file is placed inside the root directory.

The files that are *zipped* for deployment are located inside the `dist` directory.
To add files that is used on the client, place the files inside the `dist` directory.
Directory structure does not matter .

It is *important to update the ZIP file after any changes*, running `bash scripts/go_zip.sh` will update it properly.
Additionally, there is a Docker container (`cronner`) that periodically runs a cron job every 10 minutes by
accessing the token-based endpoint to update the ZIP file.

## Logging

The log output location can be defined inside the YAML configuration.
The value to the log path is expected to be *a directory*, and if a `.log` extension is attached to the
the file will be dropped, taking the parent.
- All directories will be created for the log path.

In the event of a failure, the default log output will be set to `~/.macdeploy`.

The log name follows the format: `2006-01-02T-15-04-05.<SERIAL>.log`.
- The permissions are 0600 by default.

If the **log file fails to send to the server**, ensure to save this log file if the FileVault key was generated.
**Any subsequent reruns** will not include the key in new logs if FileVault was successfully enabled. 

Logs from the client and server are found in the `logs` folder in the root directory. 
The server logs are located in the subdirectory `server-logs`.

## YAML Configuration

```yaml
# sample config
accounts:
  account_one:
    username: "EXAMPLE.NAME"
    password: "PASSWORD"
    ignore_admin: true
  account_two:
    password: "PASSWORD"
    change_password: true
admin:
  username: "USERNAME"
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
log: "/path/to/log"
filevault: false
firewall: false
```

The YAML configuration file is used for configuration of the binary.
The ***YAML should be configured prior to building the binary***.
It is <u>embedded into the binary</u>, and *any changes will require an update to the binary* 
via `bash scripts/go_zip.sh`.

The YAML file is not case sensitive, but *must be named `config.*`*. 
The extension can be any valid YAML extension.

A sample config can be found in the repository or by looking at the top of this section.

***IMPORTANT***: If special characters are used inside a string field, **it must be quoted**.

### YAML Reference

`accounts`: Creates the default users on the client device.
- `account_name`: A user info map, it can be named anything but *must be unique*. 
*This is not the admin account*.
    - `username` (string): The username of the user, this value *must be unique*. If omitted, an input prompt for a
    username will be displayed.
    - `password` (string): The password of the user used to login. If omitted, then a password prompt will appear for 
    input.
    - `apply_policy` (boolean): Apply password policies to the user from the given values.
    - `ignore_admin` (boolean): Ignores creating the user as admin if the *admin flag* is used. 
    This is only used for default accounts in the YAML config.

`admin`: A user info map for the main admin/first account of the device, used to automate majority of the workflow.
It can be omitted for security purposes. 
  - `username` (string): The username of the admin account. Can be omitted, but it must be the same as the *internal 
  username* of the MacBook during creation. 
  For example, if the display name is `Admin User` the *internal username* is `adminuser`.
  - `password` (string): The password of the admin account. If omitted, then a prompt for the password is displayed. If 
  the password fails to validate then the program will not continue.
  - `apply_policy` (boolean): Apply password policies to the user from the given values.

`packages`: Package file names that are being installed from the `pkg-files` directory on the client device.
  - `package_name` (string): The package file. The `dist` folder will be read to find any files ending in `.pkg`. 
  The file names can be an exact match or fuzzy matched, but it is recommended to put the full 
  package name in, e.g. `teamviewer.pkg`. Quotes are required if there are spaces in the package name, 
  e.g. `"Office 2016.pkg"`.
    - `installed_file_name` (string): This is the file that is installed when a `.pkg` is successfully installed. 
    It is *not case sensitive*, and should match the file name in the given search directory. 
    For example, `Microsoft Word.app` can be found by `"microsoft word"` or `"Word.app"`.
    Can be omitted in the config but must pass an empty value `-` or `- ""`.
  
`policies`: A map of password policies applied to chosen accounts in the config.
  - `reuse_password` (number): Determines if the user can reuse a password. The number ranges from 0 to 15, with 1
  being the default. 0 means the current password can be reused, 1 means the current password cannot be reused,
  and numbers between 2-15 means the user cannot reuse the last N passwords. If more than 15 is given, it will reduce
  back down to 15.
  - `alpha` (boolean): Requires the password to have at least one letter.
  - `numeric` (boolean): Requires te password to have at least one number.
  - `min_characters` (number): Minimum characters for the password.
  - `max_characters` (number): Maxmimum characters for the password. 
  - `change_on_login` (boolean): Prevents the user from logging in without changing their password. This is required
  in order to apply the other passwords.

`search_directories`: Array of paths that are used for `installed_file_name` to search for applications.

`server_host`: The URL of the server, used for client-server communication in HTTPS. 
By default it is the private IP of the server on port 5000. Any CURL requests must use `--insecure`.

`filevault`: Enable or disable FileVault activation in the deployment.

`firewall`: Enable or disable Firewall activation in the deployment.

## Limitations and Security

### Security

**Do not run this** publicly, which will cause the endpoints to be accessible to everyone.

The deployment process is expected to be ran ***on a private network***, and therefore its security is at a level where it 
protects the bare minimum. 

The endpoints do not have proper safeguards in place, although only the updating ZIP endpoint has basic authentication.
Exposing these endpoints can cause unintended consequences.

### curl

The `curl` command uses the `--insecure` option to bypass the verify check (used in the Go code too).
Although this is not recommended, it is used in this case due to the nature of device deployment- 
or in other words the devices are fully wiped prior to deployment.

## License

MacDeploy is available under the [MIT License](https://opensource.org/license/MIT).

## Acknowledgements

Special thanks to these resources:

- [Cobra CLI](https://github.com/spf13/cobra)
- [Go YAML](https://github.com/goccy/go-yaml)