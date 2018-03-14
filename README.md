# Docker Machine CloudSigma Driver

Create Docker machines on [CloudSigma](https://www.cloudsigma.com/).

You need to use your e-mail address as username and password and pass that to
`docker-machine create` with `--cloudsigma-username` and `--cloudsigma-password` options.


## Usage

    $ docker-machine create --driver cloudsigma \
        --cloudsigma-username <YOUR-EMAIL> \
        --cloudsigma-password <YOUR-PASSWORD> \
        MY_COMPUTE_INSTANCE

If you encounter any troubles, activate the debug mode with `docker-machine --debug create ...`.

### When explicitly passing environment variables

    $ export CLOUDSIGMA_USERNAME=<YOUR-EMAIL>; export CLOUDSIGMA_PASSWORD=<YOUR-PASSWORD>
    $ docker-machine create --driver cloudsigma MY_COMPUTE_INSTANCE


## Options

- `--cloudsigma-cpu`: CPU clock speed for the host in MHz.
- `--cloudsigma-drive-size`: Drive size for the host in GiB.
- `--cloudsigma-drive-uuid`: CloudSigma drive uuid.
- `--cloudsigma-memory`: Size of memory for the host in MB.
- `--cloudsigma-password`: **required** Your CloudSigma password.
- `--cloudsigma-ssh-port`: SSH port to connect.
- `--cloudsigma-ssh-user`: SSH username to connect.
- `--cloudsigma-static-ip`: CloudSigma network adapter’s static IP address.
- `--cloudsigma-username`: **required** Your CloudSigma user email.

#### Environment variables and default values

| CLI option                  | Environment variable    | Default                                |
| --------------------------- | ----------------------- | -------------------------------------- |
| `--cloudsigma-cpu`          | `CLOUDSIGMA_CPU`        | `2000`                                 |
| `--cloudsigma-drive-size`   | `CLOUDSIGMA_DRIVE_SIZE` | `20`                                   |
| `--cloudsigma-drive-uuid`   | `CLOUDSIGMA_DRIVE_UUID` | `6fe24a6b-b5c5-40ba-8860-771044d2500d` |
| `--cloudsigma-memory`       | `CLOUDSIGMA_MEMORY`     | `1024`                                 |
| **`--cloudsigma-password`** | `CLOUDSIGMA_PASSWORD`   | -                                      |
| `--cloudsigma-ssh-port`     | `CLOUDSIGMA_SSH_PORT`   | `22`                                   |
| `--cloudsigma-ssh-user`     | `CLOUDSIGMA_SSH_USER`   | `cloudsigma`                           |
| `--cloudsigma-static-ip`    | `CLOUDSIGMA_STATIC_IP`  | -                                      |
| **`--cloudsigma-username`** | `CLOUDSIGMA_USERNAME`   | -                                      |


## Contributing

We hope you'll get involved! Read our [Contributors' Guide](.github/CONTRIBUTING.md) for details.
