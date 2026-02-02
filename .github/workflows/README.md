# GitHub Actions Workflows

This directory contains the GitHub Actions workflows for the chkiso Go project.

## Workflows Overview

### 🏗️ Build Go Binary (`build-go.yml`)
**Trigger:** Pull requests, pushes to main, manual dispatch

Builds and tests the Go binary:
- **Build**: Compiles the Go binary for testing
- **Testing**: Runs comprehensive tests with multiple verification methods
- **Cross-platform Build**: Builds for all supported platforms (Windows, Linux, macOS, FreeBSD)

### 🚀 Release (`release.yml`)
**Trigger:** Release creation, manual dispatch

Builds multi-platform binaries and attaches to releases:
- Builds for Windows (amd64, arm64)
- Builds for Linux (amd64, 386, arm64, arm)
- Builds for macOS (amd64, arm64)
- Builds for FreeBSD (amd64)
- Generates SHA256 checksums for each binary
- Automatically attaches all binaries to GitHub releases

### 📦 Publish Package (`publish.yml`)
**Trigger:** Release published, manual dispatch

Handles package publishing to:
- ✅ GitHub Releases (active)
- 📋 Go package registries (future - Docker Hub, Homebrew, etc.)
- 📋 Chocolatey (future - Windows package manager)

### 📚 Documentation (`documentation.yml`)
**Trigger:** Changes to markdown files or Go code, manual dispatch

Documentation pipeline that:
- Validates markdown links
- Checks README structure
- Validates Go code documentation
- Generates usage documentation
- Creates documentation index

### 🔒 Security Scan (`security.yml`)
**Trigger:** Pushes to main, pull requests, daily schedule, manual dispatch

Security scanning including:
- CodeQL analysis for Go
- Dependency review (PRs only)
- Secret scanning with TruffleHog
- Go vet security analysis

### 🧪 Test (`test.yml`)
**Trigger:** Pull requests, pushes to main, manual dispatch

Legacy test workflow (now replaced by CI workflow):
- Compiles executable
- Runs comprehensive test suite
- Tests multiple verification methods

## Workflow Relationships

```
┌─────────────────────────────────────────────────┐
│                                                 │
│  Developer Creates PR                           │
│         │                                       │
│         ├──> CI (lint + test)                   │
│         ├──> Security Scan                      │
│         └──> Documentation Validation           │
│                                                 │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│                                                 │
│  Ready to Release                               │
│         │                                       │
│         ├──> Create Release (manual)            │
│         │        │                              │
│         │        └──> Build and Release         │
│         │                    │                  │
│         │                    └──> Publish       │
│         │                                       │
│         └──> Documentation (on main)            │
│                                                 │
└─────────────────────────────────────────────────┘
```

## Creating a Release

To create a new release:

1. **Trigger the Create Release workflow**:
   - Go to Actions → Create Release → Run workflow
   - Enter version (e.g., `1.0.0`, `1.1.0-beta`)
   - Choose if it's a pre-release
   - Click "Run workflow"

2. **The workflow will**:
   - Validate the version format
   - Generate a changelog
   - Create and push a git tag
   - Create a GitHub release

3. **Build automatically triggers**:
   - The Build and Release workflow automatically runs
   - Compiles the executable with the version number
   - Signs it (if configured)
   - Attaches artifacts to the release

4. **Publish runs automatically**:
   - Verifies the release
   - Publishes to configured targets

## Version Numbering

This project uses **Semantic Versioning** (semver):

- **MAJOR.MINOR.PATCH** (e.g., `1.0.0`)
- **MAJOR.MINOR.PATCH-prerelease** (e.g., `1.0.0-beta`, `2.0.0-rc.1`)

Examples:
- `1.0.0` - First stable release
- `1.1.0` - New features, backward compatible
- `1.1.1` - Bug fixes
- `2.0.0` - Breaking changes
- `1.2.0-beta` - Beta pre-release

## Code Signing

Code signing is optional. See [CODE_SIGNING.md](../../CODE_SIGNING.md) for setup instructions.

If signing certificates are not configured, the workflow will build unsigned executables.

## Secrets and Variables

### Required for Code Signing
- `CERTIFICATE` (secret) - Base64-encoded PFX certificate
- `CERT_PASSWORD` (secret) - Certificate password
- `CERTHASH` (variable) - Certificate SHA1 thumbprint
- `CERTNAME` (variable) - Certificate subject name

### Optional
- `GH_TOKEN` - GitHub token (automatically provided)

## Manual Workflow Triggers

All workflows can be manually triggered from the Actions tab:

1. Go to **Actions** in GitHub
2. Select the workflow
3. Click **Run workflow**
4. Fill in any required inputs
5. Click **Run workflow**

## Troubleshooting

### Build Failures
- Check if ps2exe module is installed correctly
- Verify PowerShell script syntax
- Review test failures in the CI workflow

### Release Failures
- Ensure version doesn't already exist
- Check tag naming format (must start with 'v')
- Verify GitHub token permissions

### Security Scan Failures
- Review CodeQL alerts in Security tab
- Check for secrets in commits
- Address PSScriptAnalyzer warnings

## Contributing

When contributing:
1. All PRs trigger CI and security scans
2. Ensure tests pass before merging
3. Follow semantic versioning for releases
4. Update documentation as needed

## Maintenance

- Security scans run daily at 2 AM UTC
- Review dependabot alerts regularly
- Keep GitHub Actions updated
- Monitor workflow run times and optimize as needed
