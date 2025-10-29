# Java Version Switcher (jv)

> Un tool CLI semplice e veloce per cambiare versione di Java su Windows con un singolo comando.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Windows](https://img.shields.io/badge/Platform-Windows-0078D6?logo=windows)](https://www.microsoft.com/windows)

## 🚀 Caratteristiche

- ✅ **Auto-rilevamento** automatico delle installazioni Java
- ✅ **Switch permanente** modifica le variabili d'ambiente di sistema (JAVA_HOME e PATH)
- ✅ **Search paths personalizzati** per directory custom
- ✅ **Configurazione persistente** salvata automaticamente
- ✅ **Nessuna dipendenza** eseguibile standalone
- ✅ **Supporto per tutte le distribuzioni** Oracle JDK, OpenJDK, Adoptium, Zulu, Corretto, Microsoft

## 📋 Indice

- [Prerequisiti](#-prerequisiti)
- [Installazione](#-installazione)
- [Utilizzo Rapido](#-utilizzo-rapido)
- [Comandi](#-comandi)
- [Esempi](#-esempi)
- [Come Funziona](#-come-funziona)
- [Configurazione](#-configurazione)
- [FAQ](#-faq)
- [Troubleshooting](#-troubleshooting)
- [Contribuire](#-contribuire)
- [Licenza](#-licenza)

## 🔧 Prerequisiti

- **Sistema Operativo**: Windows 10 o Windows 11
- **Go**: 1.21+ (solo per compilazione da sorgente)
- **Privilegi**: Amministratore (per modificare variabili d'ambiente di sistema)

## 📦 Installazione

### Metodo 1: Download Eseguibile (Consigliato)

1. Scarica l'ultima versione di `jv.exe` dalla pagina [Releases](https://github.com/USERNAME/java-changer/releases)
2. Copia `jv.exe` in una directory nel tuo PATH:
   - **Opzione A**: `C:\Windows\System32\` (richiede privilegi admin)
   - **Opzione B**: Crea una directory (es: `C:\tools\`) e aggiungila al PATH
3. Apri un nuovo terminale e verifica: `jv version`

### Metodo 2: Compilazione da Sorgente

```bash
# Clone del repository
git clone https://github.com/USERNAME/java-changer.git
cd java-changer

# Download dipendenze
go mod download

# Compila
go build -ldflags="-s -w" -o jv.exe ./cmd/jv

# (Opzionale) Copia nel PATH
copy jv.exe C:\Windows\System32\
```

### Aggiungere al PATH (se necessario)

1. Premi `Win + X` → Seleziona "Sistema"
2. Click su "Impostazioni di sistema avanzate"
3. Click su "Variabili d'ambiente"
4. Sotto "Variabili di sistema", seleziona "Path" → "Modifica"
5. Click "Nuovo" e aggiungi il path della directory contenente `jv.exe`
6. Click "OK" su tutte le finestre
7. **Riavvia il terminale**

## ⚡ Utilizzo Rapido

```bash
# 1. Lista tutte le versioni Java disponibili
jv list

# 2. Cambia a Java 17 (ESEGUI COME AMMINISTRATORE!)
jv use 17

# 3. Verifica la versione corrente
jv current
java -version
```

**IMPORTANTE**: Il comando `jv use` richiede privilegi amministratore. Click destro su CMD/PowerShell → "Esegui come amministratore"

## 📚 Comandi

### Gestione Versioni

| Comando | Descrizione | Esempio |
|---------|-------------|---------|
| `jv list` | Lista tutte le versioni Java disponibili | `jv list` |
| `jv use <version>` | Cambia alla versione specificata | `jv use 17` |
| `jv current` | Mostra la versione Java corrente | `jv current` |

### Installazioni Custom

| Comando | Descrizione | Esempio |
|---------|-------------|---------|
| `jv add <path>` | Aggiungi una installazione Java specifica | `jv add C:\custom\jdk-21` |
| `jv remove <path>` | Rimuovi una installazione custom | `jv remove C:\custom\jdk-21` |

### Search Paths

| Comando | Descrizione | Esempio |
|---------|-------------|---------|
| `jv add-path <dir>` | Aggiungi directory da scansionare per Java | `jv add-path C:\DevTools\Java` |
| `jv remove-path <dir>` | Rimuovi una search path | `jv remove-path C:\DevTools\Java` |
| `jv list-paths` | Mostra tutti i search paths | `jv list-paths` |

### Utilità

| Comando | Descrizione |
|---------|-------------|
| `jv version` | Mostra versione di jv |
| `jv help` | Mostra messaggio di aiuto |

## 💡 Esempi

### Scenario 1: Switch tra versioni standard

```bash
# Lista versioni disponibili
jv list

# Output:
# Available Java versions:
#
# * 17.0.1          C:\Program Files\Java\jdk-17 (auto)
#   11.0.12         C:\Program Files\Java\jdk-11 (auto)
#   1.8.0_322       C:\Program Files\Java\jdk1.8.0_322 (auto)

# Cambia a Java 11
jv use 11

# Verifica
jv current
java -version
```

### Scenario 2: Aggiungere directory custom

Se hai Java in una directory non standard (es: `C:\DevTools\Java\` con multiple versioni):

```bash
# Aggiungi la directory base
jv add-path C:\DevTools\Java

# Il detector troverà automaticamente tutte le versioni in quella directory
jv list

# Output:
# Available Java versions:
#
#   17.0.1          C:\Program Files\Java\jdk-17 (auto)
#   21.0.0          C:\DevTools\Java\jdk-21 (auto)
#   19.0.2          C:\DevTools\Java\jdk-19 (auto)
```

### Scenario 3: Aggiungere installazione specifica

```bash
# Aggiungi UNA installazione specifica
jv add D:\Projects\special-jdk-17

# Usa quella versione
jv use special
```

### Differenza tra `add` e `add-path`

**`jv add <path>`**: Aggiungi UNA installazione Java specifica
```bash
jv add C:\custom\jdk-17
# Aggiunge SOLO questa installazione
```

**`jv add-path <directory>`**: Scansiona una directory per TUTTE le installazioni Java
```bash
jv add-path C:\DevTools\Java
# Se contiene jdk-17, jdk-19, jdk-21, trova tutte e tre
```

## 🔍 Come Funziona

### 1. Auto-rilevamento

Il tool scansiona automaticamente queste directory standard:

```
C:\Program Files\Java
C:\Program Files (x86)\Java
C:\Program Files\Eclipse Adoptium
C:\Program Files\Eclipse Foundation
C:\Program Files\Zulu
C:\Program Files\Amazon Corretto
C:\Program Files\Microsoft
C:\DevTools\Java
```

Più eventuali search paths custom aggiunti con `jv add-path`.

### 2. Configurazione Persistente

La configurazione viene salvata in `%USERPROFILE%\.javarc` (file JSON):

```json
{
  "custom_paths": [
    "C:\\custom\\jdk-special"
  ],
  "search_paths": [
    "C:\\DevTools\\Java",
    "D:\\Work\\java-installations"
  ]
}
```

### 3. Modifica Variabili d'Ambiente

Quando esegui `jv use <version>`, il tool:

1. Modifica `JAVA_HOME` nel Registry di sistema
2. Aggiorna `PATH`:
   - Rimuove vecchi riferimenti Java (es: vecchio `%JAVA_HOME%\bin`)
   - Aggiunge `%JAVA_HOME%\bin` all'inizio del PATH
3. Invia un broadcast `WM_SETTINGCHANGE` per notificare Windows delle modifiche

**Tecnicamente**:
- Usa le API Windows Registry (`HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment`)
- Richiede privilegi amministratore per scrivere nel Registry di sistema
- Le modifiche sono permanenti e sopravvivono ai riavvii

### 4. Estrazione Versione

Il tool identifica la versione Java in due modi:

1. **Esegue `java -version`** e parsifica l'output
2. **Fallback**: estrae dal nome della directory (es: `jdk-17`, `jdk1.8.0_322`)

## ⚙️ Configurazione

### File di Configurazione

Posizione: `%USERPROFILE%\.javarc`

Esempio:
```json
{
  "custom_paths": [
    "C:\\MyJava\\jdk-17-custom",
    "D:\\Projects\\special-jdk"
  ],
  "search_paths": [
    "C:\\DevTools\\Java",
    "D:\\JavaInstalls"
  ]
}
```

### Modifica Manuale (Avanzato)

Puoi modificare manualmente il file `.javarc` con un editor di testo, poi esegui `jv list` per vedere le modifiche.

## ❓ FAQ

<details>
<summary><b>Devo eseguire sempre come amministratore?</b></summary>

No, solo il comando `jv use` richiede privilegi amministratore perché modifica le variabili d'ambiente di sistema. Gli altri comandi (`list`, `current`, `add-path`, ecc.) funzionano normalmente.
</details>

<details>
<summary><b>Le modifiche sono permanenti?</b></summary>

Sì! `jv use` modifica le variabili d'ambiente di sistema in modo permanente. Le modifiche sopravvivono ai riavvii e sono visibili a tutte le applicazioni.
</details>

<details>
<summary><b>Funziona con tutte le distribuzioni Java?</b></summary>

Sì! Funziona con:
- Oracle JDK
- OpenJDK
- Eclipse Adoptium (Temurin)
- Azul Zulu
- Amazon Corretto
- Microsoft OpenJDK
- Liberica JDK
- Qualsiasi altra distribuzione con la struttura standard `bin/java.exe`
</details>

<details>
<summary><b>Posso usarlo con Java 8, 11, 17, 21?</b></summary>

Sì, tutte le versioni Java sono supportate (da Java 1.6 in poi).
</details>

<details>
<summary><b>Cosa succede al PATH quando cambio versione?</b></summary>

Il tool:
1. Rimuove automaticamente vecchi path Java dal PATH
2. Aggiunge `%JAVA_HOME%\bin` all'inizio del PATH
3. Questo garantisce che la versione corretta sia sempre usata
</details>

## 🔧 Troubleshooting

### "jv non è riconosciuto come comando"

**Causa**: `jv.exe` non è nel PATH

**Soluzione**:
```bash
# Verifica dove si trova jv.exe
where jv

# Se non viene trovato, aggiungilo al PATH (vedi sezione Installazione)
```

### "failed to open registry key (run as administrator)"

**Causa**: Stai eseguendo `jv use` senza privilegi amministratore

**Soluzione**:
1. Click destro su "CMD" o "PowerShell"
2. Seleziona "Esegui come amministratore"
3. Riesegui il comando

### "No Java installations found"

**Causa**: Java non è in una directory standard o non è installato

**Soluzione**:
```bash
# Aggiungi la directory dove hai installato Java
jv add-path C:\path\to\java\directory

# Oppure aggiungi l'installazione specifica
jv add C:\path\to\jdk
```

### Le modifiche non si applicano subito

**Causa**: Il terminale o le applicazioni non hanno ricaricato le variabili d'ambiente

**Soluzione**:
1. Chiudi e riapri il terminale
2. Riavvia le applicazioni (IDE, ecc.)
3. In casi estremi, riavvia Windows

### Windows Defender blocca l'eseguibile

**Causa**: Windows potrebbe bloccare eseguibili scaricati da internet

**Soluzione**:
1. Verifica la fonte (GitHub releases ufficiale)
2. Compila da sorgente (Metodo 2)
3. Aggiungi un'eccezione in Windows Defender

### "Invalid Java installation path"

**Causa**: Il path specificato non contiene `bin\java.exe`

**Soluzione**:
```bash
# Assicurati di specificare la directory ROOT del JDK
# ✅ Corretto:
jv add C:\Program Files\Java\jdk-17

# ❌ Sbagliato:
jv add C:\Program Files\Java\jdk-17\bin
```

## 🤝 Contribuire

I contributi sono benvenuti! Se hai idee, bug reports o feature requests:

1. Apri una [Issue](https://github.com/USERNAME/java-changer/issues)
2. Fai un Fork del progetto
3. Crea un branch (`git checkout -b feature/amazing-feature`)
4. Committa le modifiche (`git commit -m 'Add amazing feature'`)
5. Pusha il branch (`git push origin feature/amazing-feature`)
6. Apri una Pull Request

## 📄 Licenza

Questo progetto è rilasciato sotto licenza MIT. Vedi il file [LICENSE](LICENSE) per dettagli.

---

## 🌟 Extra

### Struttura del Progetto

```
java-changer/
├── cmd/jv/              # Entry point CLI
├── internal/
│   ├── java/            # Rilevamento versioni Java
│   ├── config/          # Gestione configurazione
│   └── env/             # Modifica variabili d'ambiente Windows
├── go.mod
└── README.md
```

### Link Utili

- 📖 [Guida Rapida (QUICKSTART.md)](QUICKSTART.md)
- 📝 [Guida Installazione (INSTALL.md)](INSTALL.md)
- 🏗️ [Struttura Progetto (PROJECT_STRUCTURE.md)](PROJECT_STRUCTURE.md)
- 📋 [Changelog (CHANGELOG.md)](CHANGELOG.md)

---

<div align="center">

**Fatto con ❤️ per semplificare lo sviluppo Java su Windows**

[⬆ Torna su](#java-version-switcher-jv)

</div>
