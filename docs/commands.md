## About

The binary used to start the deployment process has numerous flags and supports two sub-commands:
1. `user`: Creates a local user on the device.
2. `install`: Installs a package found in the `dist` folder.

Running `macdeploy -h`, `macdeploy install -h`, or `macdeploy user -h`, will provide information on the flags
and how to use it found here.

## Flags

| Options | Description |
| ---- | ---- |
| `--admin`, `-a` | Gives admin to a created user. If `ignore_admin` is true in the YAML, this is ignored. |
| `--skiplocal`, `-s` | Skips the creation of the local user account, if configured in the YAML. |
| `--createlocal`, `-c` | Enables the local user account creation process. Skips YAML account creation if true. |
| `--cleanup` | Removes deployment files upon successful completion. |
| `--verbose`, `-v` | Output log levels INFO or above to the terminal. |
| `--debug` | Include debug logging to the terminal. |
| `--nosend` | Prevents the log from being sent to the server. |
| `--pwlist "/path/to/plist"` | Apply password policies using a plist path. |
| `--exclude "file"` | Excludes a package from installation. |
| `--include "<file,installed_file_1,installed_file_2>"` | Include a package to install. |

The `installed_file_1,installed_file_2` arguments of the`--include` flag is the installed file name, 
i.e. the files on the device after installing the package.
- For example, if `Chrome.pkg` is installed a file will be created named `Google Chrome.app` 
found inside `/Applications`. 
- To install the package and check if it is already installed: 
`--include "chrome.pkg,Google Chrome"`.

# User Creation

When `macdeploy user` is ran without any flags, it will prompt for a `username` entry
and a `password` entry.
Afterwards, the created user account will be added to FileVault.
- If the secure token fails, **do not keep the user and recreate it.** This will cause issues
with logging in once FileVault is encrypting the device.
- This process is the same as how it is done through the normal usage of `macdeploy`.

The entered username *is the display name of the user*. The internal username is automatically 
formatted to how Apple formats the username.
- `John Doe` -> `johndoe` internal / `John Doe` display.

The password is hidden by default. This can also be passed with the `-p` flag, but it is *not 
recommended to use* as it reveal the password in terminal.