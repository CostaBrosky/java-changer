# Struttura del Progetto

Panoramica completa del progetto Java Version Switcher.

## Struttura Directory

```
java-changer/
│
├── .github/                          # Configurazione GitHub
│   ├── workflows/
│   │   └── release.yml               # Pipeline CI/CD per releases automatiche
│   └── README.md                     # Documentazione GitHub config
│
├── cmd/                              # Entry points dell'applicazione
│   └── jv/
│       └── main.go                   # CLI principale con tutti i comandi
│
├── internal/                         # Codice interno (non esportabile)
│   ├── java/                         # Gestione Java installations
│   │   ├── version.go                # Strutture dati versioni Java
│   │   └── detector.go               # Auto-rilevamento installazioni
│   ├── config/                       # Gestione configurazione
│   │   └── config.go                 # Load/Save config, custom paths
│   └── env/                          # Gestione environment variables
│       └── windows.go                # Windows Registry API integration
│
├── go.mod                            # Go module definition
├── go.sum                            # Go dependencies checksums
│
├── build.bat                         # Script build per Windows
├── .gitignore                        # Git ignore rules
│
├── README.md                         # Documentazione principale
├── QUICKSTART.md                     # Guida rapida per iniziare
├── INSTALL.md                        # Guida installazione dettagliata
├── RELEASE.md                        # Guida per creare releases
├── CHANGELOG.md                      # Storia delle versioni
└── PROJECT_STRUCTURE.md              # Questo file
```

## Componenti Principali

### 1. CLI (`cmd/jv/main.go`)

Entry point dell'applicazione. Gestisce:
- Parsing dei comandi
- Routing ai vari handler
- Output user-friendly
- Gestione errori

**Comandi implementati:**
- `list`, `use`, `current` - Gestione versioni
- `add`, `remove` - Installazioni custom
- `add-path`, `remove-path`, `list-paths` - Search paths
- `version`, `help` - Utilità

### 2. Java Detector (`internal/java/`)

Responsabile di trovare e identificare installazioni Java.

**`detector.go`:**
- `FindAll()` - Trova tutte le installazioni (standard + custom)
- `GetVersion()` - Estrae versione da una installazione
- `IsValidJavaPath()` - Verifica se un path contiene Java
- `IsValidSearchPath()` - Verifica se una directory esiste

**Strategia di rilevamento:**
1. Scansiona path standard (Program Files, etc.)
2. Scansiona custom search paths da config
3. Aggiunge installazioni custom specifiche
4. Estrae versione da `java -version` o nome directory

### 3. Configuration Manager (`internal/config/`)

Gestisce la configurazione persistente in `%USERPROFILE%\.javarc`.

**Struttura Config:**
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

**Metodi principali:**
- `Load()` / `Save()` - Gestione file config
- `AddCustomPath()` / `RemoveCustomPath()` - Installazioni specifiche
- `AddSearchPath()` / `RemoveSearchPath()` - Directory di ricerca

### 4. Environment Manager (`internal/env/`)

Gestisce le variabili d'ambiente Windows tramite Registry API.

**Funzioni chiave:**
- `SetJavaHome()` - Modifica JAVA_HOME e PATH
- `updatePath()` - Rimuove vecchi path Java, aggiunge nuovo
- `broadcastSettingChange()` - Notifica Windows delle modifiche

**Dettagli tecnici:**
- Usa `golang.org/x/sys/windows/registry`
- Modifica `HKLM\System\CurrentControlSet\Control\Session Manager\Environment`
- Richiede privilegi amministratore
- Broadcast `WM_SETTINGCHANGE` per aggiornamento real-time

### 5. CI/CD Pipeline (`.github/workflows/release.yml`)

Pipeline automatica per releases.

**Trigger:**
- Push di tag `v*` (es: `v1.0.0`)
- Trigger manuale

**Processo:**
1. Checkout codice
2. Setup Go 1.21
3. Download dipendenze
4. Build con versioning (`-ldflags`)
5. Test eseguibile
6. Crea ZIP con docs
7. Genera checksums SHA256
8. Crea GitHub Release automaticamente

**Output:**
- `jv.exe` - Eseguibile standalone
- `jv-vX.Y.Z-windows-amd64.zip` - Package completo
- `checksums.txt` - SHA256 checksums

## Flusso di Lavoro Tipico

### Utente esegue: `jv use 17`

1. **main.go** riceve il comando
2. **detector.FindAll()** trova tutte le installazioni Java
3. **main.go** cerca versione matching "17"
4. **env.SetJavaHome()** modifica registry:
   - Aggiorna `JAVA_HOME`
   - Rimuove vecchi path Java da `PATH`
   - Aggiunge `%JAVA_HOME%\bin` al PATH
   - Broadcast modifiche
5. Output di successo all'utente

### Sviluppatore crea release: `git tag v1.0.0`

1. Tag pushato su GitHub
2. **GitHub Actions** si attiva automaticamente
3. **release.yml** workflow:
   - Compila `jv.exe` per Windows
   - Crea package ZIP
   - Genera checksums
   - Crea GitHub Release
4. Release disponibile per il download

## Dipendenze

### Runtime
- Nessuna dipendenza runtime - tutto embedded

### Build Time
- Go 1.21+
- `golang.org/x/sys/windows` - API Windows

### GitHub Actions
- `actions/checkout@v4`
- `actions/setup-go@v5`
- `actions/upload-artifact@v4`
- `softprops/action-gh-release@v1`

## Sicurezza

- **Privilegi amministratore**: Richiesti solo per `jv use`
- **Registry access**: Read per JAVA_HOME, Write per SetJavaHome
- **No network**: Nessuna chiamata di rete
- **No telemetry**: Zero raccolta dati
- **Open source**: Codice completamente auditable

## Performance

- **Binary size**: ~2-3 MB (con ottimizzazioni `-s -w`)
- **Startup time**: < 100ms
- **Scan time**: ~50-200ms (dipende da installazioni Java)
- **Memory usage**: < 10 MB

## Compatibilità

- **OS**: Windows 10, Windows 11
- **Architecture**: AMD64 (x86_64)
- **Java versions**: Tutte (1.8+, 11, 17, 21, etc.)
- **Java distributions**:
  - Oracle JDK
  - OpenJDK
  - Eclipse Adoptium (Temurin)
  - Zulu
  - Amazon Corretto
  - Microsoft OpenJDK
  - Altri

## Testing

### Test Manuali
```bash
# Build
go build -o jv.exe ./cmd/jv

# Test comandi base
jv.exe help
jv.exe version
jv.exe list
jv.exe list-paths

# Test come amministratore
jv.exe use 17
```

### Test Pipeline
```bash
# Simula build pipeline
go build -ldflags="-s -w -X main.Version=v1.0.0-test" -o jv.exe ./cmd/jv
```

## Contribuire

Per contribuire al progetto:

1. Fork del repository
2. Crea un branch (`git checkout -b feature/amazing-feature`)
3. Commit modifiche (`git commit -m 'Add amazing feature'`)
4. Push al branch (`git push origin feature/amazing-feature`)
5. Apri una Pull Request

Vedi anche i template per [Bug Reports](.github/ISSUE_TEMPLATE/bug_report.md) e [Feature Requests](.github/ISSUE_TEMPLATE/feature_request.md).

## License

MIT License - Vedi LICENSE file per dettagli.

---

**Documentazione aggiornata al:** 2025
