---
title: Overview
page_title: "PingOne Provisioning Provider"
description: "The PingOne Provisioning provider manages PingOne provisioning stores, plans, rules, and user attributes."
slug: provider_overview
category:
  uri: PingOne Provisioning Terraform Provider
position: 0
privacy:
  view: public
---
# PingOne Provisioning Provider

The PingOne Provisioning provider manages PingOne provisioning stores, plans, rules, and user attributes.

## Warning

This entire provider was vibe-coded and intended for our own internal testing purposes only. We are in no way affiliated
with PingOne, and take no responsibility for the correctness of the implementation, or damage this provider may cause
within your PingOne environment.

## Usage

```terraform
terraform {
  required_providers {
    pingoneprovisioning = {
      source = "easytofu/pingoneprovisioning"
    }
  }
}
```

```terraform
provider "pingoneprovisioning" {
  client_id      = var.pingone_client_id
  client_secret  = var.pingone_client_secret
  environment_id = var.pingone_environment_id
  region         = "NA"
}
```

## Authentication

The PingOne API requires `client_id`, `client_secret`, and `environment_id`. These can be provided in the provider block or via environment variables:

```shell
export PINGONE_CLIENT_ID="..."
export PINGONE_CLIENT_SECRET="..."
export PINGONE_ENVIRONMENT_ID="..."
```

The provider also supports GitHub SCIM operations. Configure a GitHub token when using those data sources:

```shell
export GITHUB_TOKEN="..."
```

## Schema

### Optional

- `client_id` (String) The Client ID for the worker application. Can also be set with the `PINGONE_CLIENT_ID` environment variable.
- `client_secret` (String) The Client Secret for the worker application. Can also be set with the `PINGONE_CLIENT_SECRET` environment variable.
- `environment_id` (String) The Environment ID where the worker application is defined. Can also be set with the `PINGONE_ENVIRONMENT_ID` environment variable.
- `region` (String) The PingOne region to use. Short codes: `NA`, `EU`, `AP`, `CA`, `AU`, `SG`. Long codes: `NorthAmerica`, `Europe`, `AsiaPacific`, `Australia-AsiaPacific`, `Canada`, `Singapore`. Default: `NA`. Can also be set with the `PINGONE_REGION` environment variable.
- `oauth_token_url` (String) Optional override for the OAuth token URL (example: `https://auth.pingone.com/<env_id>/as/token`). If unset, derived from `region` and `environment_id`.
- `api_base_url` (String) Optional override for the Management API base URL (example: `https://api.pingone.com/v1`). If unset, derived from `region`.
- `github_token` (String) GitHub classic personal access token for enterprise team APIs. Can also be set with the `GITHUB_TOKEN` environment variable.
- `github_api_base_url` (String) Optional override for the GitHub API base URL (default: `https://api.github.com`). Can also be set with the `GITHUB_API_BASE_URL` environment variable.
- `github_api_version` (String) Optional override for the GitHub API version header (default: `2022-11-28`). Can also be set with the `GITHUB_API_VERSION` environment variable.
