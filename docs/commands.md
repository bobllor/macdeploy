## About

The binary used to start the deployment process has numerous flags and supports three subcommands:
1. `user`: Local user related operations
2. `install`: Installs packages found in the `dist` folder
3. `filevault`: FileVault related operations

Nearly all subcommands *requires sudo privileges* due to it being system/device level actions.
Using these commands will *prompt for admin passwords* every time it is used.
The username is *automatically retrieved*, however in case of a failure- it wil prompt for manual input.

> It is expected that when `macdeploy` is ran, the current logged in user
> has admin privileges.

## Normal Flags

| Options | Description |
| ---- | ---- |
| `--admin`, `-a` | Gives admin to a created user. If `ignore_admin` is true in the YAML, this is ignored. |
| `--cleanup` | Removes deployment files upon successful completion. |
| `--createlocal`, `-c` | Enables the local user account creation process. Skips YAML account creation if true. |
| `--debug` | Include debug logging to the terminal. |
| `--exclude "<file>"` | Excludes a package defined in the YAML from installing. |
| `--forcefilevault` | Forces the FileVault process to overwrite existing keys with no warnings. |
| `--include "<file>[,<installed_file_1>,<installed_file_2>...]"` | Include a package to install. |
| `--plist "/path/to/plist"` | Apply password policies using a plist path. |
| `--skipfilevault` | Skips the FileVault process. |
| `--skiplocal` | Skips the creation of the local user account, if configured in the YAML. |
| `--skipsend` | Prevents the log from being sent to the server. |
| `--verbose`, `-v` | Output log levels INFO or above to the terminal. |

### About `--include` and `--exclude` Flags

The `--include` flag value is expected to be a string or a comma-separated string. The command will do
different things depending on which style is used. This can be used multiple times
to include as many files to install as needed.
- `<file>`: The `.pkg` file to install. This can be the *file name* or a *substring*.
The program will install from *substring matching*, it is recommended to type the full *file name*.
- `<installed_file>`: String value that is the *installation file name* containing the
files after the package has been installed. Values past the first comma are used to indicate the installed folder/files
of the package. This *prevents reinstalls* if the package is already installed.

There are some caveats to installations:
- All included files are *expected to be in the `dist` folder*.
- If only a *single file* is added with `--include`, then the binary will attempt to install
the program every time without checking if it exists.
- Any of the installed files are expected to be found in the *search directories*, which is given inside
the YAML file. If not given, then this *will always attempt to install the files.*

The `--exclude` flag is *only used* for excluded packages that are *added in the YAML file*.
The program will uninstall from *substring matching*, it is recommended to type the full *file name*.
Example with a YAML package entry of `antivirus.pkg`:
- `--exclude "antivirus"` will prevent the package from being installed
- `--exclude "some_pkgname_here"` will do nothing as the package does not have an entry in the YAML

## User Command

The subcommand `macdeploy user` enables operations for local users outside of the main application loop.
It provides the following:
- Creation/deletion of local users
- Granting/revoking admin privileges

User operations are at the *system level*, meaning the subcommand will require `sudo`.
- `macdeploy user list` is the only subcommand that does not require `sudo`.

There are 5 subcommands available for `macdeploy user`:
1. `macdeploy user create`: Create a new user
2. `macdeploy user delete`: Delete users from the device
3. `macdeploy user grantadmin`: Grants admin to users
4. `macdeploy user revokeadmin`: Revoke admin from users
5. `macdeploy user list`: List the current local users on the device (internal usernames)

The following flags are available to all subcommands:

| Options | Description |
| ---- | ---- |
| `--debug` | Debug level or higher logging output to `stdout` |
| `--verbose` | Info level or higher logging output to `stdout` |

### User Creation

When `create` is used, it will prompt for the username and password if empty.
Arguments are supported with the subcommand, however only the first argument is used for the username.

The username input of the username prompt *is the display name of the user*. 
The internal username is automatically generated from the given name and formatted
to match Apple's naming convention. As an example:
- (Input: `John Doe`) Display Name: `John Doe` | Internal Name: `johndoe`
- (Input: `john.doe`) Display Name: `john.doe` | Internal Name: `john.doe`

The password is hidden by default. An argument can be passed with the `-p`/`--pasword` flag for automation, 
but it is *not recommended to use* as the password will be in plain text.

> IMPORTANT
>
> Upon successful creation, the user will be *added to the list of SecureToken users*.
> This is **important to allow the user to unlock the device if the device is encrypted**.
>
> If the SecureToken process *fails*, then the *user will be deleted*. In case of a failure on the user
> deletion, *ensure to remove the user* manually- otherwise they will be unable to unlock an encrypted device.

Available flags:

| Options | Description |
| ---- | ---- |
| `--admin`, `-a` | Grants admin to the user |
| `--applypolicy` | applies the password policy for the user, this requires the `change_on_login` field to be true in the YAML config |
| `--password`, `-p` | The password string of the user, optional and recommended to not use |

### User Deletion

When deletion is required, `delete` will take arguments of users to delete from the device.
The user *must exist*, otherwise the deletion attempt will fail.

The deletion is performed based on the *internal username* of the user. However, any input given with the command
will automatically generate the internal username for deletion.
In other words, both *display name* and the *internal username* can be used to delete a user from the device.

The delete command supports multiple arguments, each argument will have a deletion attempt.

### User Admin Privileges

Both `grantadmin` and `revokeadmin` are used to grant/revoke admin privileges of a user respectively.
The user must exist and also must be the correct account type depending on the command used.
- `grantadmin` -> user is expected to not be an admin
- `revokeadmin` -> user is expected to already be an admin

Both commands support *multiple user* operations by processing any amount of arguments.

> If the `System Settings` window is open prior to running either command, the account type will not
> show its updated type. 
>
> The window must be restarted to refresh the account type in the `Users & Groups` tab.

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