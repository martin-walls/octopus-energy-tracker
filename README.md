# Octopus Energy Tracker - Watt's Occurring?

## Prerequisites

- Golang
- [just](https://github.com/casey/just)

## Environment variables

| Name                     | Value                                     |
| ------------------------ | ----------------------------------------- |
| `OCTOPUS_API_KEY`        | Your API key from the Octopus dashboard.  |
| `OCTOPUS_ACCOUNT_NUMBER` | Your Octopus account number (A-xxxxxxxx). |

## Building

Build the project with

```sh
just build
```

This will generate TypeScript types from the Go code (into `ts/types/`),
then transpile the TypeScript into JavaScript (into `static/dist/`).

### Generating TypeScript types

TypeScript types are generated from Go code using [Tygo](https://github.com/gzuidhof/tygo).
The generated output will be in the `ts-types/` directory.

To run just this step:

```sh
just generate-ts
```

If you have Tygo installed locally, the local binary will be used.
If not, it will use the `Dockerfile.generate` image.

Install Tygo locally with

```sh
go install github.com/gzuidhof/tygo@latest
```

## Running

Run the Go server with

```sh
just run
```

If you make changes to the TypeScript code, run `just build` to update the JavaScript code.
You don't need to restart the Go server.
