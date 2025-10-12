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
Windows is not supported, but WSL works.

Below are the tools and software required on the server before starting the deployment process.
- `Go`
- `docker`
- `docker compose`
- `git`
- `zip`
- `unzip`

`zip`, `unzip`, and `curl` are required on the clients. 
- MacBook devices have these installed by default, as of Sequioa and up.

### Installation and Setup

It is recommended to use the latest version:
```shell
git checkout $(git describe --tags $(git rev-list --tags --max-count=1))
```

If a specific version is needed: 
```shell
git checkout VERSION_TAG
```

Server setup:
```shell
git clone REPLACE_ME_HERE && \
cd macos-deployment && \
bash init.sh && \
git checkout $(git describe --tags $(git rev-list --tags --max-count=1)) && \
docker compose build && docker compose up -d
```

**IMPORTANT**: Before usage, the *YAML* configuration is required to be created in order
for the deployment process to work.
- Click [here](#yaml-configuration) to get started on the YAML configuration file.
- A sample configuration is given inside the project's root directory.

`go_zip.sh` located in the `scripts` folder is required to be ran after the YAML configured.
***Two binaries are generated*** in the `dist` folder and ZIP file after running `go_zip.sh`: 
`macdeploy` and `x86_64-macdeploy`.
- `macdeploy` is used for *Apple Silicon* MacBooks.
- `x86_64-macdeploy` is used for *Intel* MacBooks.

As Intel MacBooks are being phased out, the following examples will be using `macdeploy`, however the 
commands will be the same if the *Intel* version is used.

### Usage

The client devices *must be connected to the same network* as the server.

***Two binaries are generated*** in the `dist` folder and ZIP file after running `go_zip.sh`: 
`macdeploy` and `x86_64-macdeploy`.
- `macdeploy` is used for *Apple Silicon* MacBooks.
- `x86_64-macdeploy` is used for *Intel* MacBooks.

As Intel MacBooks are being phased out, the following examples will be using `macdeploy`, however the 
commands will be the same if the *Intel* version is used.

Replace `SERVER_IP_DOMAIN` with the IP or domain name to the server.
```shell
curl https://SERVER_IP_DOMAIN:5000/api/packages/deploy.zip -o deploy.zip --insecure && \
unzip deploy.zip
```
This downloads the ZIP file from the server and unzips the contents into the working directory of the client.

To start the deployment process:
```shell
./dist/macdeploy
```

The binary supports *flags*, which can be found [here](#deployment-options).

### Deployment Options

| Options | Description |
| ---- | ---- |
| `--admin`, `-a` | Gives admin to a created user. If `ignore_admin` is true in the YAML, this is ignored. |
| `--mount` | Mounts DMGs and extracts contents into the distribution directory on the client. |
| `--remove-files` | Removes deployment files upon successful completion. |
| `--verbose`, `-v` | Output debug logging to the terminal. |
| `--no-send` | Prevents the log from being sent to the server. |
| `--plist "./path/to/plist"` | Apply password policies using a plist path. |
| `--exclude "file"` | Excludes a package from installation. |
| `--include "<file/installed_file_1/installed_file_2>"` | Include a package to install. |

The `installed_file_1/installed_file_2` arguments of the`--include` flag is the installed file name, 
i.e. the files on the device after installing the package.
- For example, if `Chrome.pkg` is installed a file will be created named `Google Chrome.app` 
found inside `/Applications`. 
- To install the package and check if it is already installed: 
`--include "chrome.pkg/Google Chrome"`.

## YAML Configuration

```yaml
# sample config
accounts:
  default_account_one:
    username: "TEST.USERNAME"
    password: "PASSWORD"
    ignore_admin: true
  default_account_two:
    apply_policy: true
packages:
  package_1.pkg:
    - "installed file name.app"
  package 2:
    - "fuzzy installed app"
  excluded package.pkg:
search_directories:
  - "/Applications" 
  - "/Library/Application Support" 
scripts:
  - "example script 1.sh"
  - "example script 2.sh"
policies:
  reuse_password: 1
  require_alpha: true
  require_numeric: false
  min_characters: 5
  max_characters: 15
  change_on_login: true # REQUIRED true for policies to be applied
admin:
  username: "ADMIN_USERNAME"
  password: "ADMIN_PASSWORD"
  apply_policy: true # applies the policies above on the admin account
server_host: "https://127.0.0.1:5000"
log: "/path/to/log"
filevault: true
firewall: true
```

The YAML configuration file is used for configuration of the binary, and must be 
***configured prior to building the binary***.
It is *embedded into the binary*, meaning any new updates will require a new binary to be generated
via `bash scripts/go_zip.sh`. 

The YAML file is not case sensitive, but *must be named `config`* and can end in `.yaml`, `.yml`, `.YAML`, or `.YML`. 

### YAML Reference

`accounts`: Creates the default users on the client device.
- `account_name`: It can be named anything but *must be unique*. *This is not the admin account*.
    - `username`: This value *must be unique*. If omitted, an input prompt for a
    username will be displayed.
    - `password`: Can be omitted, a password input prompt will appear.
    - `apply_policy`: Apply password policies to the user from the given values.
    - `ignore_admin`: Ignores giving admin to the user if the *admin flag* is used. 
    This is only used for default accounts in the YAML config.

`admin`: A user info map for the main root account of the device for automation. 
It can be omitted for security purposes. 
  - `username`: It must match the same internal username during the initial account creation. It can be omitted.
  For example, if the display name is `Admin User` the *internal username* is `adminuser`.
  - `password`: If the password fails to validate then the program will exit. It can be omitted, but will prompt
  for the password.
  - `apply_policy`: Apply password policies to the admin. Must be `true` if the admin account requires
  policies applied.

`packages`: Packages that are being installed from the distribution directory.
  - `package_name`: All files ending in ending in `.pkg` are retrieved, and will be matched to the given `package_name`. 
  It is not case sensitive and fuzzy finds, but it is *recommended to match the name in the folder*.
    - `installed_file_name`: The folder added after installing a `.pkg` file. It is not case sensitive, 
    and matches the file name in the search directories. 
    Example: `Microsoft Word.app` can be found by `"microsoft word"` or `"Word.app"`.
    If omitted then the package will be installed on every attempt.

`scripts`: Array of script names that are to be executed upon the device. These files are *expected to be in the
distribution folder. All files ending in `.sh` will be obtained, and this array is used to execute matching scripts.
Prior to creating the script, ensure that it has the correct permissions.
  
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

`log`: The path to the folder of the log files. By default, it is stored at `~/logs/macdeploy`.

`filevault`: Enable or disable FileVault activation in the deployment.

`firewall`: Enable or disable Firewall activation in the deployment.

## Server and Deployment

### Zipping

Upon generation, the ZIP file is placed inside the `build` directory in the project's root directory.

The files that are *zipped for deployment* are located inside the `dist` directory.
- This is the location for all packages, scripts, DMGs, and any other files that need to be on the 
client device.

Directory structure does not matter in `dist`, but it is recommended to create separate directories in order
to keep them organized and prevent naming conflicts. Additionally, some *packages may require config files* along
side the installer.
- During deployment files paths are obtained recursively.

It is *important to update the ZIP file after any changes*, running `bash scripts/go_zip.sh` will update it properly.
Additionally, there is another Docker container that periodically runs a loop by default every 30 minutes.

### Logging

The log output location can be defined inside the YAML configuration with the field `log`.

The value to the log path is expected to be *a directory*, and if a `.log` extension is attached to the
the file will be dropped, taking the parent directory instead.
- All parent directories are created.

In the event of a failure, the default log output will be set to `~/logs/macdeploy`.

The log name follows the format: `2006-01-02T-15-04-05.<SERIAL>.log`.
- The permissions are 0600 by default.

Logs from the client and server are found in the `logs` folder in the project's root directory. 
The server logs are located in the subdirectory `server-logs`.

## Limitations and Security

### Security

**Do not run this** publicly, which will cause the endpoints to be accessible to everyone.

The deployment process is expected to be ran ***on a private network***, and therefore its security is at a level where 
it protects the bare minimum. 

The endpoints do not have proper safeguards in place, although only the updating ZIP endpoint has basic authentication.
Exposing these endpoints can cause unintended consequences.

### curl

The `curl` command uses the `--insecure` option to bypass the verification check.
It is used in this case due to the nature of device deployment, or in other words the devices are fully 
wiped prior to deployment.

## License

MacDeploy is available under the [MIT License](https://opensource.org/license/MIT).

## Acknowledgements

Special thanks to these resources:

- [Cobra CLI](https://github.com/spf13/cobra)
- [Go YAML](https://github.com/goccy/go-yaml)
- [MD Badges](https://github.com/inttter/md-badges)