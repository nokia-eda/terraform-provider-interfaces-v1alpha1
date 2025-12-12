# terraform-provider-interfaces-v1alpha1

terraform-provider-interfaces-v1alpha1 is a terraform provider plugin for the `interfaces` resource in Nokia EDA.

## Installation and Usage

Go to <https://docs.eda.dev/latest/development/terraform/>

## Development

To install the provider from the source code, run from the root of the repo:

```bash
make install
```

This make target builds provider plugin binary, stores it under `./build` relative directory and adds the `dev_overrides` block in the `${HOME}/.terraform.rc` file to point to the local binary.

This instructs terraform not to download the provider from the registry and instead look for the provider locally.

To remove the locally built provider run:

```bash
make uninstall
```

This removes the binary from the `./build` directory and removes the corresponding provider key in the `dev_overrides` block from the `${HOME}/.terraform.rc` file.

## Provider configuration variables

| TF variable              | OS env variable          | Default     | Description              |
| ------------------------ | ------------------------ | ----------- | ------------------------ |
| base_url                 | BASE_URL                 |             | Base URL                 |
| keycloak_master_realm    | KEYCLOAK_MASTER_REALM    | "master"    | Keycloak Master Realm    |
| keycloak_admin_client_id | KEYCLOAK_ADMIN_CLIENT_ID | "admin-cli" | Keycloak Admin Client ID |
| keycloak_admin_username  | KEYCLOAK_ADMIN_USERNAME  | "admin"     | Keycloak Admin Username  |
| keycloak_admin_password  | KEYCLOAK_ADMIN_PASSWORD  | "admin"     | Keycloak Admin Password  |
| client_id                | CLIENT_ID                | "eda"       | EDA Client ID            |
| client_secret            | CLIENT_SECRET            |             | EDA Client Secret        |
| realm                    | REALM                    | "eda"       | EDA Realm                |
| username                 | USERNAME                 | "admin"     | EDA Username             |
| password                 | PASSWORD                 | "admin"     | EDA Password             |
| tls_skip_verify          | TLS_SKIP_VERIFY          | false       | TLS skip verify          |
| rest_debug               | REST_DEBUG               | false       | REST Debug               |
| rest_timeout             | REST_TIMEOUT             | "15s"       | REST Timeout             |
| rest_retries             | REST_RETRIES             | 3           | REST Retries             |
| rest_retry_interval      | REST_RETRY_INTERVAL      | "5s"        | REST Retry Interval      |