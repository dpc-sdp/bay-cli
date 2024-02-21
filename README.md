# Bay CLI

A CLI tool for common tasks relating to the Single Digital Presence hosting platform, Bay.

# Usage

## Encryption

These commands simplify common cryptographic processes.

*Encrypt secrets to store in the project codebase.
```
cat /tmp/oauth.pem | bay kms encrypt \   
    --project content-foo-vic-gov-au \
    --key production > /keys/production/oauth.pem.asc
```
This will store the encrypted file at `keys/production/oauth.pem.asc`.

*Decrypt secrets stored in the codebase with this tool*
```
cat oauth.pen.asc | bay kms decrypt > oauth.pem
```

## Elastic Cloud
Commands for querying and interacting with the Elastic Cloud API.

#### Required inputs

> [!CAUTION]
> Variables are deployment specific - make sure the deployment you are targeting is not a production deployment.

* `EC_DEPLOYMENT_API_KEY` (environment variable) - Generated from the deployments Kibana settings
* `EC_DEPLOYMENT_CLOUD_ID` (command line flag) - Found on the deployments Elastic Cloud 'manage' page

#### Usage
`delete-stale` Delete indices that are greater than 30 days old

```
bay elastic-cloud delete-stale --EC_DEPLOYMENT_CLOUD_ID 'string'
```

# Installation

## Homebrew (OSX)

```
brew tap dpc-sdp/bay-cli
brew install bay-cli
```

# Binaries

Download the binaries for your OS / platform from the releases page - https://github.com/dpc-sdp/bay-cli/releases
