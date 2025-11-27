# Release Process

This repository uses semantic versioning and automated releases via GitHub Actions.

## Semantic Versioning

We follow [Semantic Versioning 2.0.0](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

## Creating a Release

1. **Update version documentation** (if needed):
   - Update `IMPLEMENTATION_COMPLETE.md` with the new version
   - Update any version references in README.md

2. **Create and push a git tag**:
   ```bash
   # For a new minor version
   git tag -a v1.1.0 -m "Release v1.1.0: Add new features"
   
   # For a patch version
   git tag -a v1.0.1 -m "Release v1.0.1: Bug fixes"
   
   # For a major version (breaking changes)
   git tag -a v2.0.0 -m "Release v2.0.0: Breaking changes"
   
   # Push the tag
   git push origin v1.1.0
   ```

3. **Automated release workflow** will:
   - Verify Go module integrity
   - Run all tests and linters
   - Build the package
   - Generate changelog from commits
   - Create GitHub release
   - Notify Go module proxy

## Release Workflow

The release workflow (`.github/workflows/release.yml`) runs on tag push and:
- ✅ Validates Go module
- ✅ Runs full test suite
- ✅ Runs golangci-lint
- ✅ Builds all packages
- ✅ Generates changelog
- ✅ Creates GitHub release
- ✅ Notifies Go proxy

## Version Compatibility

- **v1.x.x**: Current stable version
  - Go 1.21+ required
  - No breaking changes within v1.x series
  
- **v2.x.x** (future): Would indicate breaking API changes

## Go Module Usage

Users can reference specific versions:
```go
// In go.mod
require github.com/creiche/confluent-go v1.0.0

// Or latest
require github.com/creiche/confluent-go latest
```

## Best Practices

1. **Tag from main branch** after PR merge
2. **Use annotated tags** (`-a` flag) with descriptive messages
3. **Test thoroughly** before tagging
4. **Document breaking changes** in commit messages and release notes
5. **Follow conventional commits** for automatic changelog generation:
   - `feat:` for new features (minor version)
   - `fix:` for bug fixes (patch version)
   - `BREAKING CHANGE:` in footer for major version

## Manual Release Steps (if workflow fails)

1. Create tag: `git tag -a v1.0.1 -m "Release message"`
2. Push tag: `git push origin v1.0.1`
3. Create GitHub release manually at https://github.com/creiche/confluent-go/releases/new
4. Notify Go proxy: `curl https://proxy.golang.org/github.com/creiche/confluent-go/@v/v1.0.1.info`
