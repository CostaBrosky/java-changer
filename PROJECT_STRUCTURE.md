# Project Structure

Complete overview of the Java Version Switcher project.

## Directory Structure

```
java-changer/
│
├── .github/                          # GitHub configuration
│   ├── workflows/
│   │   └── release.yml               # CI/CD pipeline for automated releases
│
├── cmd/                              # Application entry points
│   └── jv/
│       └── main.go                   # Main CLI with all commands
│
├── internal/                         # Internal code (not exportable)
│   ├── java/                         # Java installation management
│   │   ├── version.go                # Java version data structures
│   │   └── detector.go               # Auto-detection of installations
│   ├── config/                       # Configuration management
│   │   └── config.go                 # Load/Save config, custom paths
│   └── env/                          # Environment variable management
│       └── windows.go                # Windows Registry API integration
│
├── go.mod                            # Go module definition
├── go.sum                            # Go dependencies checksums
│
├── build.bat                         # Windows build script
├── .gitignore                        # Git ignore rules
│
├── README.md                         # Main documentation
├── QUICKSTART.md                     # Quick start guide
├── INSTALL.md                        # Detailed installation guide
├── CHANGELOG.md                      # Version history
└── PROJECT_STRUCTURE.md              # This file
```

## Main Components

### 1. CLI (`cmd/jv/main.go`)

Application entry point. Handles:
- Command parsing
- Routing to handlers
- User-friendly output
- Error handling

**Implemented commands:**
- `list`, `use`, `current` - Version management
- `add`, `remove` - Custom installations
- `add-path`, `remove-path`, `list-paths` - Search paths
- `version`, `help` - Utilities

### 2. Java Detector (`internal/java/`)

Responsible for finding and identifying Java installations.

**`detector.go`:**
- `FindAll()` - Finds all installations (standard + custom)
- `GetVersion()` - Extracts version from an installation
- `IsValidJavaPath()` - Verifies if a path contains Java
- `IsValidSearchPath()` - Verifies if a directory exists

**Detection strategy:**
1. Scans standard paths (Program Files, etc.)
2. Scans custom search paths from config
3. Adds specific custom installations
4. Extracts version from `java -version` or directory name

### 3. Configuration Manager (`internal/config/`)

Manages persistent configuration in `%USERPROFILE%\.javarc`.

**Config structure:**
```json
{
  "custom_paths": [
    "C:\\custom\\jdk-specific"
  ],
  "search_paths": [
    "C:\\DevTools\\Java"
  ]
}
```

**Main methods:**
- `Load()` / `Save()` - Config file management
- `AddCustomPath()` / `RemoveCustomPath()` - Specific installations
- `AddSearchPath()` / `RemoveSearchPath()` - Search directories

### 4. Environment Manager (`internal/env/`)

Manages Windows environment variables via Registry API.

**Key functions:**
- `SetJavaHome()` - Modifies JAVA_HOME and PATH
- `updatePath()` - Removes old Java paths, adds new
- `broadcastSettingChange()` - Notifies Windows of changes

**Technical details:**
- Uses `golang.org/x/sys/windows/registry`
- Modifies `HKLM\System\CurrentControlSet\Control\Session Manager\Environment`
- Requires administrator privileges
- Broadcasts `WM_SETTINGCHANGE` for real-time updates

## Typical Workflow

### User runs: `jv use 17`

1. **main.go** receives the command
2. **detector.FindAll()** finds all Java installations
3. **main.go** searches for version matching "17"
4. **env.SetJavaHome()** modifies registry:
   - Updates `JAVA_HOME`
   - Removes old Java paths from `PATH`
   - Adds `%JAVA_HOME%\bin` to PATH
   - Broadcasts changes
5. Success output to user

## Dependencies

### Runtime
- No runtime dependencies - everything is embedded

### Build Time
- Go 1.21+
- `golang.org/x/sys/windows` - Windows APIs

## Security

- **Administrator privileges**: Required only for `jv use`
- **Registry access**: Read for JAVA_HOME, Write for SetJavaHome
- **No network**: No network calls
- **No telemetry**: Zero data collection
- **Open source**: Fully auditable code

## Performance

- **Binary size**: ~2-3 MB (with `-s -w` optimizations)
- **Startup time**: < 100ms
- **Scan time**: ~50-200ms (depends on Java installations)
- **Memory usage**: < 10 MB

## Compatibility

- **OS**: Windows 10, Windows 11
- **Architecture**: AMD64 (x86_64)
- **Java versions**: All (1.8+, 11, 17, 21, etc.)
- **Java distributions**:
  - Oracle JDK
  - OpenJDK
  - Eclipse Adoptium (Temurin)
  - Zulu
  - Amazon Corretto
  - Microsoft OpenJDK
  - Others

## Testing

### Manual Testing
```bash
# Build
go build -o jv.exe ./cmd/jv

# Test basic commands
jv.exe help
jv.exe version
jv.exe list
jv.exe list-paths

# Test as administrator
jv.exe use 17
```

## Contributing

To contribute to the project:

1. Fork the repository
2. Create a branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See also templates for [Bug Reports](.github/ISSUE_TEMPLATE/bug_report.md) and [Feature Requests](.github/ISSUE_TEMPLATE/feature_request.md).

## License

MIT License - See LICENSE file for details.

---

**Documentation updated:** 2025
