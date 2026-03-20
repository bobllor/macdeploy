## About

The `config` YAML is used to configure the binary, allowing for customization of how the binary
will do the process. It is *embeded into the binary* for the program to run properly. 

This is *required* in order to compile the binary. However, the only required field
is *`server_host`*, as this is how the clients *retrieve the binary and deployment files*
and to start the process.

When the script `go_zip.sh` is ran, the file will be ran through a validation check. If it fails to validate,
then the ZIP process will be canceled and an error will be displayed for a fix.

## YAML Reference

These are uncategorized fields of the YAML file. The only *required field* is the *`server_host` field*, although
it is recommended to have `filevault` and `firewall` to be *`true`* for security reasons.
- By default `filevault` and `firewall` are false.

Values:
- `install_directories`: An array of paths that are the directories where the install files of a `.pkg` installation
will be located in.
- `server_host`: The IP or domain of the server, used for client-server communication. *This is required* for the
deployment to work and it *must be a URL* (`https` or `http`).
- `filevault`: Enable or disable FileVault activation in the deployment.
- `firewall`: Enable or disable Firewall activation in the deployment.

```yaml
install_directories: # when pkg files are installed, the files will be installed into these directories
  - /Applications
  - /Library/Application Support
server_host: "https://169.254.1.5:5000" # the server host, can be a domain or a local IP, this must be a URL
filevault: true # enables filevault process for the binary
firewall: false # disables firewall process for the binary
```

### `Accounts`

It is a dictionary that contains a dictionary which contains the *user info*
that is used used to *automate local account creation* on the device.

This can be omitted if no users need to be created. If omitted, then the account creation
will prompt for a creation.

Values:
- `account_name`: The dictionary key, it is only used to hold the *user info*.
This can be named anything but it *must be a unique value*.
  - `username`: This value *must be unique* otherwise it will fail if the user exists
  in the device. If omitted, an input prompt for a username will be displayed.
  - `password`: If given, it will use this password as the password for the user.
  Can be omitted, a password input prompt will appear.
  - `apply_policy`: Apply password policies to the user.
  - `ignore_admin`: Ignores granting admin to the user if the *admin flag* is used. 
  This applies only for accounts defined in the YAML config.

```yaml
accounts:
  account_one: # creates a user with no prompt, and ignores the admin flag if given.
    username: "defaultuser0"
    password: "PASSWORD"
    ignore_admin: true
  account_two: # prompts during the user creation for username and password.
    apply_policy: true # applies a password policy to this user
```

### Admin

A dictionary used to store the information of the admin account. This is similar to
the `accounts` dictionary, but does not use a separate dictionary to hold user information.

Due to this being in plain text, *it is recommended not to include the password* and type it in
manually on the deploying device instead.
- It can be included for automation purposes, as this will skip any `sudo` prompt, assuming it is correct.

This dictionary can be omitted entirely, but will require a password input for `sudo` elevation. The
`username` will also take the current logged in user via `whoami` by default.

Values:
- `username`: It must match the same internal username during the initial account creation.
For example, if the display name is `Admin User` the *internal username* is `adminuser`. Can be omitted.
- `password`: It can be omitted, but will prompt for the password. If it fails to validate then the program will exit.
- `apply_policy`: Apply password policies to the admin account. Must be `true` if the admin account requires
policies applied.

```yaml
admin: # username and password can be omitted.
  username: "ADMIN_USERNAME"
  password: "ADMIN_PASSWORD"
  apply_policy: true # applies the policies above on the admin account
```

### Cleanup

The `cleanup` field is used to get user confirmation before removing the deployment files
at the end of the deployment process.

It has two valid values:
1. `warn`: Prompts for user confirmation before removing files.
2. `force`: It will remove the files with no user prompt.

By default `cleanup` has the value `warn`, it will prompt every time before the deployment files
are removed.

If an invalid value is used for `cleanup`, the binary will refuse to run and the `go_zip.sh` will
fail to create the binary during validation.

### Packages

A dictionary containing dictionaries that have a *string key* and an *array value*.
The key represents the *`.pkg` file name to install*, while the value is an 
*array of strings that are the installed file names*.
- All keys must have a colon (`:`) at the end, even if no array is used.

This is used to install applications *by default* on program run. The files are
*expected to be in the `dist` folder* of the deployment files.

This is optional, if omitted then no packages will be installed.

Values:
- `package_name`: The package name ending in `.pkg`. Used to execute scripts found in the distribution directory. 
It is *case insensitive* and *does a substring match*. 
  - `<installed_file_name>`: The installation files added after a `.pkg` installation. It is not case sensitive, 
  and does a substring match with the *files found in the install folders*. 

The `<installed_file_name>` is the value used in an array and has no limit to the entries. 
It can be empty, but the package will be installed on every attempt no matter if it is 
installed or not.

```yaml
packages:
  package_1.pkg: # install a pkg file containing `package_1.pkg` with installation files named `package 1.app`
    - "package 1.app"
  package 2: # install a pkg file containing `package_2` with installation files containing `package_2`
    - "package_2"
  package 3.pkg: # install a pkg file containing `package 3.pkg` with no installation files
```

### Policies

A dictionary that contains the basic password policy applications for a user. This is used to force
change passwords for the local user on login. It is *recommended* for security reasons.

If defined, `change_on_login` is *required*. This enables the password policies to be used on the account
login.

Values:
- `policies`: A map of password policies applied to chosen accounts in the config.
  - `reuse_password`: Determines if the user can reuse a password. Ranges from 0 to 15, with 1 being the default. 
  - `require_alpha`: Requires the password to have at least one letter.
  - `require_numeric`: Requires the password to have at least one number.
  - `min_characters`: Minimum characters for the password.
  - `max_characters`: Maxmimum characters for the password. 
  - `change_on_login`: Requires a password change before logging in. This is **required** in order 
  to apply the password policies.

```yaml
policies:
  reuse_password: 1 # cannot reuse the last password within 1 from the current
  require_alpha: true # password must contain a letter
  require_numeric: false # password can include or not include numbers
  min_characters: 5
  max_characters: 15
  change_on_login: true # REQUIRED true for policies to be applied
```

### Scripts

The scripts dictionary is used to inject script execution during certain stages of the process lifecycle.
There are *three deployment stages* where it executes:
1. `pre`: Before the deployment process starts
2. `mid`: During the deployment process, right after package installation
3. `post`: After the deployment process ends

All three categories are expected to be *an array of strings*, and the files *must be inside the `dist` folder*.
In other words, the files are packaged into the ZIP file for deployment.

> IMPORTANT
>
> The script files *must be executable* in order to be ran on the client device.
> The script execution will always be logged, all status codes and output will always be logged.

Values:
- `scripts`: The dictionary start field for the scripts.
  - `pre`: Runs during after the initialization of the deployment, before 
  - `mid`: Scripts to be executed during deployment, this is executed after installation of packages. 
  - `post`: Scripts to be executed after deployment.

```yaml
scripts:
  pre: # run before the main deployment start, often used for initializations
    - change_hostname.sh
  mid: # run during the deployment, after package installation
    - add_to_desktop.sh
  post: # runs after the deployment, often used for cleanups or finishing touches
    - clean_up.sh
```