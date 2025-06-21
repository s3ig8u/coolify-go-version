# Version Management for Coolify Go

This document describes the version management system for Coolify Go, which ensures that every change is properly versioned and tracked.

## Overview

The version management system automatically updates the version on every commit and provides tools for manual version control. It follows [Semantic Versioning](https://semver.org/) (SemVer) principles:

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

## Automatic Version Management

### Pre-commit Hook

Every commit automatically bumps the **patch version** unless:

- The version file is already being committed
- It's a merge commit
- It's a revert commit

The pre-commit hook is located at `.git/hooks/pre-commit` and runs automatically.

### Version File

The version information is stored in `version.go`:

```go
const (
    Version   = "v1.4.0"
    BuildTime = "2025-06-20T11:38:00Z"
    GitCommit = "abc123def"
)
```

## Manual Version Management

### Using the Version Script

The `scripts/version.sh` script provides comprehensive version management:

```bash
# Show current version information
./scripts/version.sh info

# Bump versions
./scripts/version.sh patch    # v1.4.0 → v1.4.1
./scripts/version.sh minor    # v1.4.0 → v1.5.0
./scripts/version.sh major    # v1.4.0 → v2.0.0

# Set specific version
./scripts/version.sh set v2.0.0

# Create git tags
./scripts/version.sh tag "Release with new features"
./scripts/version.sh push-tag
./scripts/version.sh release "Major release"
```

### Using Makefile Commands

The Makefile provides convenient shortcuts:

```bash
# Version management
make version-info      # Show current version
make version-bump      # Bump patch version
make version-minor     # Bump minor version
make version-major     # Bump major version
make version-set       # Set specific version (VERSION=v1.2.3)
make version-tag       # Create git tag
make version-release   # Create and push release
```

## Version Bumping Guidelines

### When to Bump Major Version (X.0.0)

- Breaking changes to the API
- Incompatible database schema changes
- Major architectural changes
- Removal of deprecated features

### When to Bump Minor Version (0.X.0)

- New features added
- New API endpoints
- New configuration options
- Backwards-compatible enhancements

### When to Bump Patch Version (0.0.X)

- Bug fixes
- Security patches
- Documentation updates
- Performance improvements
- **Every commit** (automatic)

## Release Workflow

### Development Workflow

1. **Make changes** to the codebase
2. **Commit changes** - patch version is automatically bumped
3. **Test thoroughly** before release
4. **Create release** when ready:

```bash
# For patch releases (automatic)
git commit -m "Fix bug in authentication"
# Version automatically bumped by pre-commit hook

# For minor releases
make version-minor
git commit -m "Add new user management features"
make version-tag "New user management features"
make version-release

# For major releases
make version-major
git commit -m "Breaking: New API design"
make version-tag "Major API redesign"
make version-release
```

### Production Release Process

1. **Ensure all tests pass**:

   ```bash
   make test
   make quick-test
   ```

2. **Check migration status**:

   ```bash
   make migrate-status
   ```

3. **Create release**:

   ```bash
   make version-release "Production release v1.4.0"
   ```

4. **Build and deploy**:
   ```bash
   make docker
   # Deploy to production
   ```

## Version Information in Application

The application exposes version information through:

### Health Check Endpoint

```bash
curl http://localhost:8080/health
```

Response:

```json
{
  "status": "healthy",
  "version": "v1.4.0",
  "buildTime": "2025-06-20T11:38:00Z",
  "commit": "abc123def",
  "timestamp": "2025-06-20T11:38:30.123Z",
  "database": "connected"
}
```

### Version Endpoint

```bash
curl http://localhost:8080/version
```

Response:

```json
{
  "version": "v1.4.0",
  "buildTime": "2025-06-20T11:38:00Z",
  "commit": "abc123def"
}
```

### Command Line

```bash
./coolify-go -version
```

Output:

```
Coolify Go v1.4.0 (built 2025-06-20T11:38:00Z, commit abc123def)
```

## Database Schema Versioning

The database schema is versioned using a hash-based system:

```bash
# Check schema version
make migrate-schema-info

# View migration status
make migrate-status
```

The schema hash ensures consistency across deployments and helps detect schema drift.

## Best Practices

### ✅ DO

- Let the pre-commit hook handle patch version bumps automatically
- Use semantic versioning for manual bumps
- Create git tags for releases
- Include meaningful commit messages
- Test thoroughly before major/minor releases
- Document breaking changes in release notes

### ❌ DON'T

- Manually edit `version.go` directly
- Skip version bumps for important changes
- Create releases without testing
- Use non-semantic version numbers
- Forget to push git tags

## Troubleshooting

### Pre-commit Hook Not Running

```bash
# Check if hook is executable
ls -la .git/hooks/pre-commit

# Make executable if needed
chmod +x .git/hooks/pre-commit
```

### Version Script Not Found

```bash
# Check if script exists and is executable
ls -la scripts/version.sh

# Make executable if needed
chmod +x scripts/version.sh
```

### Git Tag Already Exists

The version script will warn you if a tag already exists. You can:

- Use a different version number
- Delete the existing tag: `git tag -d v1.4.0`
- Force push: `git push origin :refs/tags/v1.4.0`

## Integration with CI/CD

The version management system integrates with CI/CD pipelines:

### GitHub Actions

```yaml
- name: Build with version info
  run: |
    VERSION=$(git describe --tags --always)
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    GIT_COMMIT=$(git rev-parse --short HEAD)
    go build -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
```

### Docker Build

```bash
docker build \
  --build-arg VERSION="$(git describe --tags --always)" \
  --build-arg BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ)" \
  --build-arg GIT_COMMIT="$(git rev-parse --short HEAD)" \
  -t coolify-go:latest .
```

## Migration from Manual Versioning

If you're migrating from manual version management:

1. **Backup current version**:

   ```bash
   cp version.go version.go.backup
   ```

2. **Initialize version management**:

   ```bash
   ./scripts/version.sh set v1.4.0
   ```

3. **Install pre-commit hook**:

   ```bash
   chmod +x .git/hooks/pre-commit
   ```

4. **Test the system**:
   ```bash
   make version-info
   # Make a small change and commit to test auto-bump
   ```

## Support

For issues with version management:

1. Check this documentation
2. Run `./scripts/version.sh help`
3. Check the pre-commit hook logs
4. Verify git configuration
