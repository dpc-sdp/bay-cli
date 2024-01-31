# Bay CLI

A CLI tool for common tasks relating to the Single Digital Presence hosting platform, Bay.

# Usage

## Encryption

These commands simplify common cryptographic processes.

*Encrypt secrets to store in the project codebase.*
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

## Project Map

These commands provide a simple way to show how projects relate to eachother.

*Show all frontends connected a specific backends*

```sh
bay project-map by-backend content-vic
bay project-map by-backend content-health-vic-gov-au content-legalaid-vic-gov-au 
bay project-map by-backend --all --output=json
```

*Show which backend a specific frontend connects to*

```sh
bay project-map by-frontend vic-gov-au 
bay project-map by-frontend --all
```

### Outputs

These commands accept a `--output` flag which can be one of the following values:

* `json` - a json object
* `table` - a cli table
* `go-template-file` - accepts another flag `--go-template-file` which indicates a go template to use for output. See `examples/templates` for ways to use this.