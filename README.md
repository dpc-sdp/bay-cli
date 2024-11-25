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

## Project mapping
### by-backend
List FE projects that a supplied Tide CMS (Lagoon metadata `backend-project`) connect to.

### by-frontend
Get the BE project that the supplied FE projects connect to.

## Deployment metadata
The deployment metadata command returns a json encoded string with project metadata that is not generally available to pods during the build process.

```
> bay deployment metadata
{"deployment":{"sha":"58d27fb5218e4703bd7c1a696471f69f259226cd","authorName":"Guy Owen","when":"2024-11-25 12:36:16 +1100 +1100","tag":"tag not found","msg":"[SD-323] Added deployment cmd."}}
```

# Installation

## Homebrew (OSX)

```
brew tap dpc-sdp/bay-cli
brew install bay-cli
```

# Binaries

Download the binaries for your OS / platform from the releases page - https://github.com/dpc-sdp/bay-cli/releases
