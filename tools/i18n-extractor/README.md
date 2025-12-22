# i18n-extractor

`i18n-extractor` is a tool that scans Go source files for `core.I18nMessage` literals and generates a `defaults.go` file containing all extracted keys and their default values.

## Usage

The tool is integrated into the project's build process via the `Makefile`.

To generate i18n defaults:

```bash
make generate_i18n
```

This command will:
1. Build the extractor tool and run unit tests (only if not already installed).
2. Extract i18n messages from the backend source code.

## Development

The source code of the tool is located in `tools/i18n-extractor`.

### Running Unit Tests

To run the unit tests manually:

```bash
cd tools/i18n-extractor
go test -v .
```

### Force Rebuild/Install

If you have modified the code of the `i18n-extractor` itself, you may need to force a rebuild of the tool binary to ensure your changes are picked up. You can do this using the `-B` flag with `make`:

```bash
make -B install-i18n-extractor
```

Or to force everything including generation:

```bash
make -B generate_i18n
```
