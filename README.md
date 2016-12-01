# retrievault
Retrieve [Vault](https://vaultproject.io) secrets and save them into files.

## Table of contents

- [Introduction](#introduction)
- [Usage](#usage)
  - [Type "generic"](#type-generic)
    - [Example](#example)
  - [Type "certs"](#type-certs)
- [Deployment](#deployment)  
  - [Standalone script](#standalone-script)
    - [Download](#download)
    - [Run it!](#standalone-run-it)
  - [Docker](#docker)
    - [Build the image](#build-image)
    - [Run it! :)](#docker-run-it)
- [TODO](#todo)

---

## Introduction

**retrievault** is a tool that allow you to fetch [Vault](https://vaultproject.io) secrets and store them into files, in an easy way. Some of the features of retrievault are listed below:

- **Easy JSON configuration file**: You make a descriptive list of secrets you want to fetch and where you want to store them in a **config.json** file.

- **Support for both [Generic](https://www.vaultproject.io/docs/secrets/generic/index.html) and [PKI](https://www.vaultproject.io/docs/secrets/pki/index.html) secret backends**: Supports retrieving key-value secrets and certificates from Generic and PKI backends.

- **Easy destination handling**: You can specify the destination path per secret, or even per element in a secret (e.g.: if you have multiple key-value pairs inside a single Generic secret, you can specify different file locations for each key-value pair).

- **Easy permission handling**: Just like before, you can specify permission bits per secret or per element in a secret. Those permission bits are specified in octal notation, event if we know that JSON doesn't support them. They are parsed as strings and converted inside retrievault.

---

## Usage

You can use **retrievault** both as a standalone script as well as a [Docker](https://www.docker.com/) container.

In either way, the first thing to do is to create a config.json file in which we specify the secrets we want to fetch, the location where we want to store them and some Vault configurations like the Vault Token and the Vault Address. Note that the token must be **able to fetch all the secrets specified**. In the future we plan to add the ability to specify one token per secret.

Below we show a full example of such a file:

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

### Type "generic"<a name=type-generic></a>

The "generic" type accepts the following `parameters`:
- **keys**: This is a map where the keys are strings which match the key of the component of a secret, and the value are some parameteres related to the destination file. Let's have a closer look to this by means of an example.

#### Example

Suppose you write a generic secret into vault:
```
vault write generic/some_secret first_component="hello" second_component="world"
```

Then you can fetch this secret into two different files. We want both files to be stored at the same directory location, but with different file names. We can use the following JSON configuration file for retrievault to fetch them:

```json
{
  "log_level": "info",
  "log_file": "/var/log/retrievault.log",
  "secrets": [
    {
      "type": "generic",
      "path": "/etc/some/path",
      "vault_path": "generic/some_secret",
      "parameters": {
        "keys": {
          "first_component": {
            "path": "my_secret_1",
            "perm": "0600"
          },
          "second_component": {
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

As we can see, in the *keys* map we specify a `path` and a `perm` for the key `first_component`, so it will be stored with permissions `0600` at path `/etc/some/path/my_secret_1`.

However, for the `second_component`, we just specify the `perm`. **retrievault** will use the key of the secret component as the name of the file in the destination path. So the `second_component` of the secret will be stored at: `/etc/some/path/second_component`.

We could also have specified an absolute `path` for one of the components. This will override the `path` top-level key in the secret. For example, if we have specified the following configuration for the `second_component`:

```json
"second_component": {
  "path": "/etc/another/path/my_secret_2",
  "perm": "0644"
}
```

Then the component will be stored at `/etc/another/path/my_secret_2`, overriding the previous `/etc/some/path`.

### Type "certs"<a name=type-certs></a>

The "certs" type accepts the following `parameters`:

- **common_name**: The requested CN for the certificate. If the CN is allowed by role policy, it will be issued.
- **ttl**: Requested Time To Live. Cannot be greater than the role's `max_ttl` value. If not provided, the role's `ttl` value will be used. Note that the role values default to system values if not explicitly set.
- **alt_names**: Requested Subject Alternative Names. These can be host names or email addresses; they will be parsed into their respective fields. If any requested names do not match role policy, the entire request will be denied.
- **ip_sans**: Requested IP Subject Alternative Names. Only valid if the role allows IP SANs (which is the default).
- **key**: This allows you to define specific destination options for the key issued. It behaves similar to the *keys* parameter of the "generic" backend (see the previous section).
- **cert**: Same as the one before, but concerning the public certificate issued.
- **ca_cert**: Same as the one before but concerning the certificate issued.

It is worth noting that **retrievault** will automatically handle certificates' chain of trust by appending the certificates in the chain to the public certificate issued. For more details, please dive into the source code.

We suggest you to have a look at the full example above, for an example of using the "certs" secret type with **retrievaukt**.

## Deployment

As we mentioned before, we can use **retrievault** as a standalone script or as a Docker container.

### Standalone script<a name=standalone-script></a>

#### Download

You can download **retrievault** by executing the following command.

```
wget -O /usr/local/bin/retrievault https://github.com/DatioBD/retrievault/raw/master/docker/retrievault
```

#### Run it!<a name=standalone-run-it></a>

Run **retrievault** by specifying the path to the config file:

```
retrievault --config /path/to/config.json
```

You can optionally specify the log level or the log output file directly from the cli. This will override the value from the `config.json` file:

```
retrievault --config /path/to/config.json --log-level debug --log-file stdout
```

### Docker

#### Build the image<a name=build-image></a>

Run the following commands in order to build the image:

```
git clone https://github.com/DatioBD/retrievault.git
cd retrievault
make
```

#### Run it! :)<a name=docker-run-it></a>

Run the following command in order to run the Docker container:

```
docker run -d -v path/to/config.json:/etc/retrievault/config/config.json -e VAULT_ADDR="https://vault.address:8200" VAULT_TOKEN="vault_token" some-repo/retrievault:latest
```

It should be advise to you that all secrets are stored inside the Docker container. In order to make them available to other docker containers, you should **mount a volume at the destination path of each secret fetched**, and share this volume across all the containers which must get this secret.

## TODO

- Improve logging
- TESTS!!
- Handle renewal of certificates and generic secrets
- Compatibility with more Vault secret backends
