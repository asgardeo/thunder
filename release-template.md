# Release Template for Thunder Documentation

This template can be used to update the README.md file when creating new releases.

## Update Instructions

When creating a new release (e.g., v0.9.0), update the following in README.md:

### 1. Download and Run WSO2 Thunder section

Replace the version number `0.8.0` with the new version throughout the section:

```markdown
1. **Download the distribution from the X.X.X release**

    Download the appropriate distribution for your operating system and architecture:
    
    - **Linux x64**: [thunder-X.X.X-linux-x64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/thunder-X.X.X-linux-x64.zip)
    - **Linux ARM64**: [thunder-X.X.X-linux-arm64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/thunder-X.X.X-linux-arm64.zip)
    - **macOS x64**: [thunder-X.X.X-macos-x64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/thunder-X.X.X-macos-x64.zip)
    - **macOS ARM64**: [thunder-X.X.X-macos-arm64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/thunder-X.X.X-macos-arm64.zip)
    - **Windows x64**: [thunder-X.X.X-win-x64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/thunder-X.X.X-win-x64.zip)
    
    For other architectures, visit the [latest release page](https://github.com/asgardeo/thunder/releases/latest).
```

### 2. Docker images section

Update Docker tag from `0.8.0` to the new version:

```bash
docker pull ghcr.io/asgardeo/thunder:X.X.X
```

### 3. Sample App section

Replace the version number `0.8.0` with the new version:

```markdown
1. Download the sample app from the X.X.X release

    Download the appropriate sample app for your operating system and architecture:
    
    - **Linux x64**: [sample-app-X.X.X-linux-x64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/sample-app-X.X.X-linux-x64.zip)
    - **Linux ARM64**: [sample-app-X.X.X-linux-arm64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/sample-app-X.X.X-linux-arm64.zip)
    - **macOS x64**: [sample-app-X.X.X-macos-x64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/sample-app-X.X.X-macos-x64.zip)
    - **macOS ARM64**: [sample-app-X.X.X-macos-arm64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/sample-app-X.X.X-macos-arm64.zip)
    - **Windows x64**: [sample-app-X.X.X-win-x64.zip](https://github.com/asgardeo/thunder/releases/download/vX.X.X/sample-app-X.X.X-win-x64.zip)
    
    For other architectures, visit the [latest release page](https://github.com/asgardeo/thunder/releases/latest).
```

### 4. Unzip commands

Update unzip and directory navigation commands:

```bash
unzip thunder-X.X.X-<os>-<arch>.zip
cd thunder-X.X.X-<os>-<arch>/
```

```bash
unzip sample-app-X.X.X-<os>-<arch>.zip
cd sample-app-X.X.X-<os>-<arch>/
```

### 5. Configuration paths

Update the sample app configuration path:

```
sample-app-X.X.X-<os>-<arch>/app
```

## Search and Replace Strategy

Use your editor's search and replace function to update all occurrences:

1. Replace `0.8.0` with the new version number
2. Replace `v0.8.0` with `vX.X.X` (where X.X.X is the new version)
3. Verify all download links are correct by checking the actual release assets

## Verification

After updating, verify that:
1. All download links are accessible (return 302 redirects)
2. Version numbers are consistent throughout the document
3. Docker image tags match the release version
4. File paths and directory names match the new version