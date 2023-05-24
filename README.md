# Bay CLI

A CLI tool for common tasks relating to the Single Digital Presence hosting platform, Bay.

# Usage

## Encryption

These commands simplify common cryptographic processes.

*Encrypt secrets to store in the project codebase.
```
# If piping stdout from another command, use "-" as the input .
bay kms encrypt \   
    --project content-foo-vic-gov-au \
    --key production \
    --input path/to/original.txt \
    --filename original.txt
```
This will store the encrypted file at `keys/production/original.txt.asc`, which will be decrypted to the same path without the `.asc` extension.