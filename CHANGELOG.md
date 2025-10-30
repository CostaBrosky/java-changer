# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-10-30

### Added
- Initial public release of Java Version Switcher (jv)
- Interactive TUI powered by Charm (Huh prompts, Lip Gloss styles)
- Auto-detection of Java installations from standard directories and custom search paths
- Persistent configuration (custom installations, search paths)
- System environment management: sets `JAVA_HOME` and ensures `%JAVA_HOME%\bin` in `PATH`
- Commands:
  - `list`, `use`, `switch`, `current`
  - `install`, `doctor`, `repair`
  - `add`, `remove`, `add-path`, `remove-path`, `list-paths`
  - `help`, `version`
- PowerShell installer script: user-level install, PATH update, config init, autocomplete setup

### Security
- Administrator privileges required for system-wide environment changes (`use`, `switch`, `install`, `repair`)

---

[1.0.0]: https://github.com/CostaBrosky/jv/releases/tag/v1.0.0
