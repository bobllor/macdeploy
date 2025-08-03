# About

This is the TGS Hardware's macOS deployment guide.

For more detailed information on how to use, configure, and manage the deployment files visit the TEKsystems Quick Survival Guide or KBXXX on ServiceNow.
<br />
This README contains short run downs of the server, and below you can find the section to get the process started.

Features:
- Automatic SSH key generation and authorization to the server.
- Automatic installation of the required software.
- Automatic user creation.
- Automatic enabling of FileVault and FireWall.
- Generates the key of the FileVault and stores it in the server.
- Generates logs and stores it in the server.

**IMPORTANT**: Although a majority of the manual interaction with the macOS deployment has been eliminated, there are still some manual interactions required. These can be found at:
1. Copying of the SSH key to the server (2 inputs).
2. The first root access requirement (1 input).
3. User creation (1 input).
4. FileVault activation (2 inputs).

# Quick Start

## Before you Read

There is a QR code that runs the commands for you when scanned. The below is the way to do it manually.

## Initial Setup

Before we start, the deployment script is required on the device. Run the following one liner in terminal:

```shell
ssh-keygen -f ~/.ssh/id_rsa -q -N "" -C "$(date +"%Y-%m-%dT%T:%M%S)"; ssh-copy-id donotmigrate@10.142.46.165; scp -rq donotmigrate@10.142.46.165://Users/donotmigrate/mac-deployment/client-files ~
```

This command does two things:
1. It generates the SSH key and copies it to the server to prevent multiple password inputs later in the script.
2. It installs the required files from the server onto the client device.

You can chain the start up script afterward the last command if needed.

## Run the Script

The script contains **two types of flags** that can be used with the script:
- `T`: TeamViewer flag, used to install TeamViewer.
- `A`: Admin flag, used to make the user admin.
<br />
By default:
- All users are created as standard accounts.
- TeamViewer is not installed.

The commands to run the script for each use case:

Standard: `bash deploy.sh -T`
Admin: `bash deploy.sh -T -A` or `bash deploy.sh -TA`
No TeamViewer: `bash deploy.sh`
No TeamViewer and Admin: `bash deploy.sh -A`