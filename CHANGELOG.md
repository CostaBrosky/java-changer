# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - TBD

### Added
- Initial release of Java Version Switcher (jv)
- Auto-detection of Java installations from standard paths:
  - `C:\Program Files\Java`
  - `C:\Program Files (x86)\Java`
  - `C:\Program Files\Eclipse Adoptium`
  - `C:\Program Files\Eclipse Foundation`
  - `C:\Program Files\Zulu`
  - `C:\Program Files\Amazon Corretto`
  - `C:\Program Files\Microsoft`
  - `C:\DevTools\Java`
- Custom search paths support via `add-path` command
- Permanent system environment variable modification (JAVA_HOME and PATH)
- Commands:
  - `jv list` - List all available Java versions
  - `jv use <version>` - Switch to specified Java version
  - `jv current` - Show current Java version
  - `jv add <path>` - Add specific Java installation
  - `jv remove <path>` - Remove custom installation
  - `jv add-path <dir>` - Add directory to search for Java installations
  - `jv remove-path <dir>` - Remove directory from search paths
  - `jv list-paths` - Show all search paths (standard + custom)
  - `jv version` - Show version information
  - `jv help` - Show help message
- Persistent configuration in `%USERPROFILE%\js.json`
- Windows Registry integration for system-wide changes
- Comprehensive documentation (README, QUICKSTART, INSTALL guides)

### Features
- Single executable with no external dependencies
- Automatic broadcast of environment variable changes to Windows
- Intelligent PATH management (removes old Java paths, adds new)
- Support for multiple Java distributions (Oracle, OpenJDK, Adoptium, Zulu, Corretto, etc.)
- Version detection from both `java -version` and directory names

### Security
- Requires administrator privileges for system-wide changes
- No network requests or telemetry
- Open source and auditable code

---

## Version History Legend

- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** in case of vulnerabilities

[Unreleased]: https://github.com/CostaBrosky/jv/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/CostaBrosky/jv/releases/tag/v1.0.0
