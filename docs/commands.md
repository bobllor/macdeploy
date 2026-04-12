## About

The binary used to start the deployment process has numerous flags and supports three sub-commands:
1. `user`: Creates a local user on the device.
2. `install`: Installs a package found in the `dist` folder.
3. `filevault`: FileVault related operations.

Running `macdeploy -h`, `macdeploy install -h`, or `macdeploy user -h`, will provide information on the flags
and how to use it found here.

## Normal Flags

| Options | Description |
| ---- | ---- |
| `--admin`, `-a` | Gives admin to a created user. If `ignore_admin` is true in the YAML, this is ignored. |
| `--skiplocal` | Skips the creation of the local user account, if configured in the YAML. |
| `--createlocal`, `-c` | Enables the local user account creation process. Skips YAML account creation if true. |
| `--cleanup` | Removes deployment files upon successful completion. |
| `--verbose`, `-v` | Output log levels INFO or above to the terminal. |
| `--debug` | Include debug logging to the terminal. |
| `--skipsend` | Prevents the log from being sent to the server. |
| `--skipfilevault` | Skips the FileVault process. |
| `--plist "/path/to/plist"` | Apply password policies using a plist path. |
| `--exclude "<file>"` | Excludes a package defined in the YAML from installing. |
| `--include "<file>[,<installed_file_1>,<installed_file_2>...]"` | Include a package to install. |

The `--include` flag value is expected to be a string or a CSV string. The command will do
different things depending on which style is used. This can be used multiple times
to include as many files to install as needed.
- `<file>`: The `.pkg` file to install. This can be the *file name* or a *substring*.
The program will install from *substring matching*, it is recommended to type the *file name*.
- `<installed_file>`: Comma-separate values that is the *installation file name* containing the
files after the package has been installed. This is used for conditional installations.
This is often found in the `/Applications` folder and ends with the extension `.app`, but it can be any installation files.

The `--exclude` flag is *only used* for excluded packages that are *added in the YAML file*.
This is uses *substring matching* to remove packages, but it is recommended to use the full name of
the file to remove.
- If the package `antivirus.pkg` has an entry in the YAML, then `--exclude "antivirus"` will prevent
the package from being installed.

There are some caveats to installations:
- All included files are *expected to be in the `dist` folder*.
- If only a *single file* is added with `--include`, then the binary will attempt to install
the program every time without checking if it exists.
- Any of the installed files are expected to be found in the *search directories*, which is given inside
the YAML file. If not given, then this *will always attempt to install the files.*

## User Creation

When `macdeploy user` is ran without any flags, it will prompt for a `username` entry
and a `password` entry.
Afterwards, the created user account will be added to FileVault.
- If the secure token fails, **do not keep the user and recreate it.** This will cause issues
with logging in once FileVault is encrypting the device.
- This process is the same as how it is done through the normal usage of `macdeploy`.

The entered username *is the display name of the user*. The internal username is automatically 
formatted to how Apple formats the username.
- `John Doe` -> `johndoe` internal / `John Doe` display.

The password is hidden by default. An argument can be passed with the `-p`/`--pasword` flag, but it is *not 
recommended to use* as it reveal the password in terminal.

### `user` Flags

| Options | Description |
| ----- | ----- |
| `-a`, `--admin` | Grants admin to the user |
| `--applypolicy` | Applies a password policy on login, requires options in config defined |
| `--debug` | Enables debug logging |
| `-u <string>`, `--username <string>` | The username of the user |
| `-p <string>`, `--password <string>` | The password of the user |
| `-v`, `--verbose` | Enables info logging |

## Package Installation

`macdeploy install` requires *positional arguments*, which represents the file name
to install. This works similarly to *how package installation works in the normal process*,
it is expected to be *located in the `dist` folder* and it *matches the package via substring*.
- Example: `macdeploy install chrome.pkg antivirus.pkg "some pkg here"` will install packages
containing the strings `chrome.pkg`, `antivirus.pkg`, or `some pkg here`.
- Ensure *quotations are used* when a file has spaces in its name.
- It is *recommended to type the full name out* regardless, as it can match incorrectly.

The `install` subcommand also supports DMG extraction with the flag `--mountdmg`.
This will automatically mount, extract, and unmount the data into the `dist` folder.

It *does not support* installation file name conditions, it will *always attempt to install* if used. 

### `install` flags

| Options | Description |
| ----- | ----- |
| `--debug` | Enables debug logging |
| `-v`, `--verbose` | Enables info logging |
| `--mountdmg` | Mounts and extracts the contents of DMG files |

## FileVault Operations

`macdeploy filevault <command>` is used for FileVault related operations on a MacBook device.
These are the available commands for the operation:
- `disable`: Disables FileVault
- `enable`: Enable FileVault
- `list`: Lists the users that are added to FileVault (users who can unlock the file disk)
- `status`: Checks the status of FileVault

The two commands `disable` and `enable` both have two flags that allow for quicker automation:
1. `--username`/`-u`: The admin username. If left blank, it will assume that the current 
logged in user is admin with `whoami`.
2. `--password`/`-p`: The admin password. It left blank, it will use a hidden input prompt with
confirmation. It is recommended this to be left blank.

> IMPORTANT
>
> Operations regarding FileVault requires `sudo` permission, regardless of if its a status check.
> Enabling and disabling FileVault does require an admin account in order to perform.

### `filevault` flags

| Options | Description |
| ----- | ----- |
| `--debug` | Enables debug logging |
| `-v`, `--verbose` | Enables info logging |