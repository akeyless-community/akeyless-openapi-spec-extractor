# Akeyless OpenAPI Spec Extractor

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The Akeyless OpenAPI Spec Extractor is a command-line interface (CLI) tool that allows you to extract specific endpoints from an OpenAPI specification (Swagger) along with all the relevant details of the endpoint. It provides a convenient way to fetch OpenAPI specs from a URL or process local OpenAPI spec files.

## Project Genesis and Purpose

The Akeyless OpenAPI Spec Extractor project was born out of the need to extract comprehensive information about a specific endpoint from a large list of endpoints in an OpenAPI specification. The goal was to make it easier to send this information to a large language model to assist with code generation for interacting with that endpoint using tools for agentic workflows.

By extracting all the relevant details of an endpoint, including inputs and outputs, the Akeyless OpenAPI Spec Extractor simplifies the process of integrating with APIs and facilitates the development of intelligent agents that can understand and interact with specific endpoints based on the extracted information.

## Features

- Fetch OpenAPI specs from a URL
- Process local OpenAPI spec files
- Process OpenAPI specs from stdin
- Extract specific endpoints based on the provided path
- Output the extracted data in JSON or YAML format
- Customize logging levels for better visibility and debugging

## Installation

To install the Akeyless OpenAPI Spec Extractor, make sure you have [Node.js](https://nodejs.org) installed on your system. Then, run the following command:

```bash
npm install -g akeyless-openapi-spec-extractor
```

This will install the CLI tool globally on your system.

## Usage

The Akeyless OpenAPI Spec Extractor provides three commands: `fetch`, `local`, and `stdin`.

### Fetch Command

The `fetch` command allows you to fetch an OpenAPI spec from a URL and extract a specific endpoint.

```bash
openapi-extractor fetch -u <url> -p <path> [options]
```

- `-u, --url <url>`: URL of the OpenAPI spec (required)
- `-p, --path <path>`: The path to the desired endpoint within the OpenAPI spec (including the leading slash), for example, "/auth" (required)
- `-o, --output <output>`: Output type (default: "json"): "json" or "yaml"
- `-l, --loglevel <level>`: Logging level (default: "error"): "error", "warn", "info", "debug"

### Local Command

The `local` command allows you to process a local OpenAPI spec file and extract a specific endpoint.

```bash
openapi-extractor local -f <file> -p <path> [options]
```

- `-f, --file <file>`: Path to the local OpenAPI spec file (required)
- `-p, --path <path>`: The path to the desired endpoint within the OpenAPI spec (including the leading slash), for example, "/auth" (required)
- `-o, --output <output>`: Output type (default: "json"): "json" or "yaml"
- `-l, --loglevel <level>`: Logging level (default: "error"): "error", "warn", "info", "debug"

### Stdin Command

The `stdin` command allows you to process an OpenAPI spec from stdin and extract a specific endpoint.

- `-p, --path <path>`: The path to the desired endpoint within the OpenAPI spec (including the leading slash), for example, "/auth" (required)
- `-o, --output <output>`: Output type (default: "json"): "json" or "yaml"
- `-l, --loglevel <level>`: Logging level (default: "error"): "error", "warn", "info", "debug"

## Examples

Fetch an OpenAPI spec from a URL and extract the "/auth" endpoint:

```bash
openapi-extractor fetch -u https://api.example.com/openapi.json -p "/auth"
```

Process a local OpenAPI spec file and extract the "/users" endpoint:

```bash
openapi-extractor local -f ./openapi.yaml -p "/users"
```

Process an OpenAPI spec from stdin and extract the "/products" endpoint:

```bash
cat openapi.yaml | openapi-extractor stdin -p "/products"
```

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for more details.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvement, please open an issue or submit a pull request on the [GitHub repository](https://github.com/akeyless-community/akeyless-openapi-spec-extractor).

## Acknowledgements

This CLI tool is built using the following open-source libraries:

- [openapi-extract](https://www.npmjs.com/package/openapi-extract)
- [commander](https://www.npmjs.com/package/commander)
- [axios](https://www.npmjs.com/package/axios)
- [winston](https://www.npmjs.com/package/winston)
- [js-yaml](https://www.npmjs.com/package/js-yaml)

We would like to express our gratitude to the maintainers and contributors of these libraries for their excellent work.
