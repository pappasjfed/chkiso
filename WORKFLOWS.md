# Quick Start Guide - CI/CD Workflows

This guide helps you quickly get started with the chkiso CI/CD system.

## For Contributors

### Making Changes

1. **Create a branch** from main
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes** to the code

3. **Push your branch**
   ```bash
   git push origin feature/my-feature
   ```

4. **Create a Pull Request**
   - GitHub Actions will automatically:
     - ✓ Lint your PowerShell code
     - ✓ Build and test the executable
     - ✓ Run security scans
     - ✓ Validate documentation

5. **Review and merge**
   - Address any issues found by CI
   - Once approved, merge to main

## For Maintainers

### Creating a Release

#### Step 1: Trigger the Release Workflow

1. Go to [GitHub Actions](https://github.com/pappasjfed/chkiso/actions)
2. Click on **"Create Release"** workflow
3. Click **"Run workflow"** button
4. Fill in the form:
   - **Version**: Enter semantic version (e.g., `1.0.0`, `1.1.0`, `2.0.0-beta`)
   - **Pre-release**: Check if this is a pre-release

#### Step 2: Automatic Build

Once the release is created, the **Build and Release** workflow automatically:
- ✓ Extracts version from the git tag
- ✓ Compiles `chkiso.ps1` to `chkiso.exe` with version number
- ✓ Signs the executable (if certificates configured)
- ✓ Generates SHA256 checksums
- ✓ Attaches all files to the GitHub release

#### Step 3: Publishing

The **Publish Package** workflow then:
- ✓ Verifies the release artifacts
- ✓ Publishes to GitHub Releases
- 📋 (Future) Can publish to PowerShell Gallery
- 📋 (Future) Can publish to Chocolatey

### Version Numbers

Use [Semantic Versioning](https://semver.org/):

- `1.0.0` - First stable release
- `1.0.1` - Patch (bug fixes)
- `1.1.0` - Minor (new features, backward compatible)
- `2.0.0` - Major (breaking changes)
- `1.1.0-beta` - Pre-release

### Managing Releases

#### View All Releases
```bash
gh release list
```

#### View Specific Release
```bash
gh release view v1.0.0
```

#### Download Release Assets
```bash
gh release download v1.0.0
```

## Workflow Status

Check workflow status with badges in README.md:

- 🟢 Green: All checks passing
- 🔴 Red: Failed (click to see details)
- 🟡 Yellow: In progress

## Common Tasks

### Run Tests Locally

```powershell
# Install dependencies
Install-Module -Name ps2exe -Force

# Build executable
ps2exe -inputFile chkiso.ps1 -outputFile chkiso.exe `
  -noConsole:$false -title "chkiso" -version "1.0.0.0"

# Run tests
.\chkiso.exe test\test.iso
```

### Lint Code Locally

```powershell
# Install PSScriptAnalyzer
Install-Module -Name PSScriptAnalyzer -Force

# Run linter
Invoke-ScriptAnalyzer -Path chkiso.ps1 -Recurse
```

### Generate Documentation Locally

```powershell
# View help
Get-Help .\chkiso.ps1 -Full
```

## Security

### Code Signing Setup

If you have a code signing certificate:

1. Set up secrets in GitHub:
   - `CERTIFICATE` - Base64-encoded PFX
   - `CERT_PASSWORD` - Certificate password

2. Set up variables:
   - `CERTHASH` - Certificate SHA1 thumbprint
   - `CERTNAME` - Certificate subject name

See [CODE_SIGNING.md](../CODE_SIGNING.md) for details.

### Security Scans

Security scans run:
- ✓ On every pull request
- ✓ On every push to main
- ✓ Daily at 2 AM UTC
- ✓ Manually via workflow dispatch

## Troubleshooting

### Workflow Failed?

1. Click on the failed workflow badge
2. Click on the failed job
3. Expand the failed step
4. Read the error message
5. Fix the issue and push again

### Release Creation Failed?

- **Version already exists**: Use a new version number
- **Invalid version format**: Use semantic versioning (X.Y.Z)
- **Tag already exists**: Delete the tag first (if needed)

### Build Failed?

- Check PowerShell syntax errors
- Verify ps2exe can compile the script
- Review test failures

## Getting Help

- 📖 [Full Workflow Documentation](.github/workflows/README.md)
- 🔐 [Code Signing Guide](CODE_SIGNING.md)
- 📚 [Main README](README.md)
- 🐛 [Report Issues](https://github.com/pappasjfed/chkiso/issues)

## Advanced

### Manual Workflow Triggers

All workflows can be triggered manually:

1. Go to **Actions** tab
2. Select the workflow from the left sidebar
3. Click **"Run workflow"**
4. Fill in any required inputs
5. Click **"Run workflow"** button

### Workflow Files

- `ci.yml` - Continuous integration (lint + test)
- `release.yml` - Create releases with versioning
- `build-release.yml` - Build and attach to releases
- `publish.yml` - Publish to package managers
- `documentation.yml` - Documentation validation and generation
- `security.yml` - Security scanning
- `test.yml` - Legacy test workflow

## Best Practices

1. ✅ Always create PRs for changes
2. ✅ Wait for CI to pass before merging
3. ✅ Use semantic versioning for releases
4. ✅ Write descriptive commit messages
5. ✅ Update documentation with code changes
6. ✅ Review security scan results
7. ✅ Test locally before pushing

## Next Steps

- [ ] Configure code signing (optional)
- [ ] Set up PowerShell Gallery publishing (optional)
- [ ] Set up Chocolatey publishing (optional)
- [ ] Enable GitHub Pages for documentation (optional)

---

**Questions?** Open an issue or check the workflow documentation!
