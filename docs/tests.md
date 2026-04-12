## About

This section contains manual tests that are performed manually on MacBook devices.

Due to the binary requiring `sudo` and performing system-level/destructive actions, an Action Runner
cannot be used for testing the actual deployment.

Test cases will be under the version or future version (if applicable) and its descriptions.
Do note that the test cases are only available for *versions 3* and above.

An example of a test case:
```md
### Version 5.1.60

Test: Normal User Creation
- Expected Behavior: 
    - User is created via information in the YAML file.
    - User is created via the flag `-c` when running the binary.
- Result: Pass (2099-06-12T13:23:44.214+00:00)

Test: Server Log Payload
- Expected Behavior:
    - Log storage is skipped.
- Result: Fail (2099-06-12T13:23:44.214+00:00)
- Notes:
    - Forgot to add the flag.
```

## Version 3.0.0

> Version 3.0.0:
> - Overhauled FileVault process behavior
> - New flag and flag name changes
> - Subcommand `filevault` added to `macdeploy`

### Test 1

Normal FileVault process
- Baseline:
    - FileVault is disabled on the device.
    - The server does not have an entry of the device.
- Expected Behavior:
    - FileVault is enabled and the info is stored on the server.
- Result: Pass (2026-04-12T19:56:52.535+00:00)

### Test 2

FileVault with device entry with no key and disabled status
- Baseline:
    - Server has a device entry but does not have the key stored.
    - FileVault is disabled on the device.
- Expected Behavior:
    - The FileVault key is added into the server.
    - Server logs has an entry informing there is no key found for the entry.
- Result: Pass (2026-04-12T19:56:52.535+00:00)

### Test 3

FileVault with device entry with no key and enabled status
- Baseline:
    - Server has a device entry but does not have the key stored.
    - FileVault is enabled on the device.
- Expected Behavior:
    - FileVault is automatically disabled during the process after the query.
    - The FileVault key is added into the server.
    - Server logs has an entry informing there is no key found for the entry.
- Result: Pass (2026-04-12T20:05:22.024+00:00)

### Test 4

FileVault with no device entry and enabled status
- Baseline:
    - Server does not have a device entry.
    - FileVault is enabled on the device.
- Expected Behavior:
    - FileVault is automatically disabled during the process after the query.
    - The full device entry is added into the server.
- Result: Pass (2026-04-12T20:09:44.457+00:00)

### Test 5

Disable FileVault with `macdeploy filevault`
- Baseline:
    - FileVault is enabled.
- Expected Behavior:
    - FileVault is disabled.
- Result: Pass (2026-04-12T20:13:20.436+00:00)

### Test 6

Enable FileVault with `macdeploy filevault`
- Baseline:
    - FileVault is disabled.
- Expected Behavior:
    - FileVault is enabled.
    - Key is stored on the server.
- Result: Pass (2026-04-12T20:14:54.321+00:00)

### Test 7

Normal `macdeploy` with FileVault enabled
- Baseline:
    - FileVault is enabled.
    - Server has full info stored.
    - Condition for key replacement is not met (<= 1.5 hours).
- Expected Behavior:
    - FileVault process is skipped.
    - Output is issued to use `--forcefilevault` flag.
- Result: Pass (2026-04-12T20:17:34.761+00:00)

### Test 8

Flag `--forcefilevault`
- Baseline:
    - FileVault is enabled.
    - Server has full info stored.
    - Condition for key replacement is not met (<= 1.5 hours).
- Expected Behavior:
    - FileVault is disabled.
    - FileVault process is started.
    - Stored key is replaced.
- Result: Pass (2026-04-12T20:31:52.710+00:00)