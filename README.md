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
