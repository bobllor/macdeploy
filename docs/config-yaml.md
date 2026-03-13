# YAML Reference

`accounts`: Creates the users on the client device. Can be omitted if no users need to be created.
- `account_name`: It can be named anything but *must be unique*.
    - `username`: This value *must be unique*. If omitted, an input prompt for a
    username will be displayed.
    - `password`: Can be omitted, a password input prompt will appear.
    - `apply_policy`: Apply password policies to the user from the given values.
    - `ignore_admin`: Ignores giving admin to the user if the *admin flag* is used. 
    This applies only for accounts defined in the YAML config.

`admin`: A user info map for the main account. 
  - `username`: It must match the same internal username during the initial account creation.
  For example, if the display name is `Admin User` the *internal username* is `adminuser`. Can be omitted.
  - `password`: It can be omitted, but will prompt for the password. If it fails to validate then the program will exit.
  - `apply_policy`: Apply password policies to the admin account. Must be `true` if the admin account requires
  policies applied.

`packages`: Packages that are being installed from the distribution directory.
  - `package_name`: The package name ending in `.pkg`. Used to execute scripts found in the distribution directory. 
  It is *case insensitive* and *looks for a name match*. 
    - `installed_file_name`: The directory added after installing a `.pkg` file. It is not case sensitive, 
    and matches the file name in the search directories. 
    Example: `Microsoft Word.app` can be found by `"microsoft word"` or `"Word.app"`.
    If omitted then the package will be installed on every attempt.

`scripts`: Scripts to be executed on the client device. It executes in three deployment stages: before, during, and after.
The script files *must have the correct permission* prior to being compressed into the ZIP file. 
Each section is an array of script names, it is *case insensitive* and *looks for a name match*.
  - `pre`: Scripts to be executed before deployment.
  - `inter`: Scripts to be executed during deployment, this is executed after installation of packages. 
  - `post`: Scripts to be executed after deployment.
  
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

`filevault`: Enable or disable FileVault activation in the deployment.

`firewall`: Enable or disable Firewall activation in the deployment.
