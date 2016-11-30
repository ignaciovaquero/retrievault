# retrievault
Retrieve [Vault](https://vaultproject.io) secrets and expose them into files or environment variables.

## Introduction

**retrievault** is a tool that allow you to fetch [Vault](https://vaultproject.io) secrets and store them into files, in an easy way. Some of the features of retrievault are listed below:

- **Easy JSON configuration file**: You make a descriptive list of secrets you want to fetch and where you want to store them in a **config.json** file.

- **Support for both [Generic](https://www.vaultproject.io/docs/secrets/generic/index.html) and [PKI](https://www.vaultproject.io/docs/secrets/pki/index.html) secret backends**: Supports retrieving key-value secrets and certificates from Generic and PKI backends.

- **Easy destination handling**: You can specify the destination path per secret, or even per element in a secret (e.g.: if you have multiple key-value pairs inside a single Generic secret, you can specify different file locations for each key-value pair).

- **Easy permission handling**: Just like before, you can specify permission bits per secret or per element in a secret. Those permission bits are specified in octal notation, event if we know that JSON doesn't support them. They are parsed as strings and converted inside retrievault.
---
## Usage

You can use **retrievault** both as a standalone script as well as a [Docker](https://www.docker.com/) container.

In either way, the first thing to do is to create a config.json file in which we specify the secrets we want to fetch, the location where we want to store them and some Vault configurations like the Vault Token and the Vault Address. Note that the token must be **able to fetch all the secrets specified**. In the future we will add the ability to specify one token per secret.

Below we show an example of such a file:

```json
{
  "log_level": "debug",
  "log_file": "/var/log/retrievault.log",
  "secrets": [
    {
      "type": "certs",
      "path": "/etc/retrievault/certs",
      "vault_path": "pki/issue/rolename",
      "parameters": {
        "common_name": "common-name.yourdomain.com",
        "ttl": "24h",
        "alt_names": [
          "common.yourdomain.com",
          "localhost"
        ],
        "ip_sans": [
          "10.142.0.1"
        ],
        "key": {
          "path": "common.key",
          "perm": "0600"
        },
        "cert": {
          "path": "common.crt",
          "perm": "0644"
        },
        "ca_cert": {
          "path": "/etc/retrievault/ca-cert/common-ca.crt",
          "perm": "0644"
        }
      }
    },
    {
      "type": "generic",
      "path": "/etc/retrievault/generic",
      "vault_path": "generic/github_ssh_keys",
      "parameters": {
        "keys": {
          "private": {
            "path": "id_rsa_github",
            "perm": "0600"
          },
          "public": {
            "path": "id_rsa_github.pub",
            "perm": "0644"
          }
        }
      }
    }
  ],
  "ca_cert_path": "/usr/local/share/ca-certificates/ca.crt",
  "insecure": false,
  "vault_addr": "https://vault.yourdomain.es:8200",
  "vault_token": "token"
}
```

Let's explain those options one by one:
- **log_level**: Sets the log level for retrievault. This can be set to one of "debug", "info", "warn", "error", "fatal" and "panic". Defaults to `info`.
- **log_file**: The output file for logging. This can also be set to "stderr" or "stdout". Defaults to `/var/log/retrievault.log`.
- **ca_cert_path**: The path to a PEM-encoded CA cert file to use to verify the Vault server SSL certificate.
- **insecure**: Enables or disables SSL verification. Defaults to `false`.
- **vault_addr**: The Vault Address. This can also be set via the environment variable **VAULT_ADDR**.
- **vault_token**: The Vault Token for fetching all of the secrets. This can also be set via the environment variable **VAULT_TOKEN**.
- **secrets**: An array of secrets to fetch. All secret types have common properties like:
  - **type**: The type of the secret. Currently, we support only "generic" and "certs". This is mandatory.
  - **path**: This is optional and can be set to an absolute or relative directory. If the destination directory doesn't exist it will be created. By setting the path here we set this as the base path for all the components of the secret (keys or certs, depending on the secret type). If we take a look to the example above, the keys fetched at the secret of type "generic" will be stored at `/etc/retrievault/generic/id_rsa_github` and `/etc/retrievault/generic/id_rsa_github.pub` respectively.
  - **vault_path**: The Vault path to fetch the secret. This is mandatory.
  - **parameters**: Parameters specific to the secret type. See the corresponding secret type to find out more about this.

### Standalone script
