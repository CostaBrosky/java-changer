# Java Version Switcher (jv)

Un tool CLI semplice e veloce per cambiare versione di Java su Windows con un singolo comando.

## Caratteristiche

- **Auto-rilevamento**: Trova automaticamente le installazioni Java in path standard
- **Path personalizzati**: Supporta l'aggiunta di installazioni Java custom
- **Switch permanente**: Modifica le variabili d'ambiente di sistema (JAVA_HOME e PATH)
- **Facile da usare**: Comandi semplici e intuitivi
- **Nessuna dipendenza**: Eseguibile standalone

## Prerequisiti

- Windows 10/11
- Go 1.21+ (solo per la compilazione)
- Privilegi di amministratore (per modificare variabili d'ambiente di sistema)

## Installazione

### 1. Compilare da sorgente

```bash
# Clone o scarica questo repository
cd D:\Work\java-changer

# Scarica le dipendenze
go mod download

# Compila l'eseguibile
go build -o jv.exe ./cmd/jv

# (Opzionale) Compila con ottimizzazioni per ridurre dimensione
go build -ldflags="-s -w" -o jv.exe ./cmd/jv
```

### 2. Aggiungere al PATH

Per usare `jv` da qualsiasi directory:

1. Copia `jv.exe` in una directory nel tuo PATH (es: `C:\Windows\System32`) oppure
2. Aggiungi la directory corrente al PATH di sistema:
   - Apri "Variabili d'ambiente" dal Pannello di Controllo
   - Modifica la variabile PATH
   - Aggiungi il path dove si trova `jv.exe`

## Utilizzo

### Visualizzare tutte le versioni Java disponibili

```bash
jv list
```

Output esempio:
```
Available Java versions:

* 17.0.1          C:\Program Files\Java\jdk-17 (auto)
  1.8.0_322       C:\Program Files\Java\jdk1.8.0_322 (auto)
  21.0.0          C:\custom\jdk-21 (custom)
```

La versione corrente è contrassegnata con `*`.

### Cambiare versione Java

```bash
jv use 17
```

Il comando cercherà la versione che contiene "17" e la attiverà. Non serve specificare la versione completa.

```bash
jv use 1.8
jv use 21
```

### Visualizzare versione corrente

```bash
jv current
```

Output:
```
Current Java version: 17.0.1
JAVA_HOME: C:\Program Files\Java\jdk-17
```

### Aggiungere installazione Java custom

Se hai Java installato in una directory non standard:

```bash
jv add C:\custom\jdk-21
```

Il tool verificherà che sia una installazione Java valida (presenza di `bin\java.exe`).

### Rimuovere un path custom

```bash
jv remove C:\custom\jdk-21
```

### Gestione Search Paths (Nuovo!)

Se hai Java installato in directory non standard (es: `C:\DevTools\Java`), invece di aggiungere ogni versione manualmente, puoi aggiungere la directory base e il detector la scannerà automaticamente.

#### Aggiungere un search path

```bash
jv add-path C:\DevTools\Java
```

Questo dirà al detector di cercare automaticamente tutte le installazioni Java in `C:\DevTools\Java`. Se hai `C:\DevTools\Java\jdk-17` e `C:\DevTools\Java\jdk-21`, entrambe verranno rilevate automaticamente.

#### Visualizzare tutti i search paths

```bash
jv list-paths
```

Output esempio:
```
Java search paths:

Standard paths (built-in):
  C:\Program Files\Java [exists]
  C:\Program Files (x86)\Java
  C:\Program Files\Eclipse Adoptium
  ...

Custom search paths:
  C:\DevTools\Java [exists]
```

#### Rimuovere un search path

```bash
jv remove-path C:\DevTools\Java
```

**Differenza tra `add` e `add-path`:**
- `jv add C:\custom\jdk-17` → Aggiunge UNA installazione specifica
- `jv add-path C:\custom` → Scansiona la directory per TUTTE le installazioni Java

### Aiuto

```bash
jv help
```

## Come funziona

1. **Auto-rilevamento**: Il tool scansiona automaticamente queste directory:
   - `C:\Program Files\Java`
   - `C:\Program Files (x86)\Java`
   - `C:\Program Files\Eclipse Adoptium`
   - `C:\Program Files\Eclipse Foundation`
   - `C:\Program Files\Zulu`
   - `C:\Program Files\Amazon Corretto`
   - `C:\Program Files\Microsoft`
   - `C:\DevTools\Java`
   - Più eventuali search paths custom aggiunti con `jv add-path`

2. **Configurazione persistente**: Salvata in `%USERPROFILE%\.javarc` (file JSON)
   - **Search paths custom**: Directory da scansionare automaticamente
   - **Installazioni custom**: Path specifici a singole installazioni Java

3. **Modifica variabili d'ambiente**:
   - Modifica `JAVA_HOME` nel registry di sistema
   - Aggiorna `PATH` rimuovendo vecchi riferimenti Java e aggiungendo `%JAVA_HOME%\bin`
   - Invia un broadcast WM_SETTINGCHANGE per notificare il sistema

4. **Privilegi amministratore**: Necessari per modificare le variabili d'ambiente di sistema (`HKEY_LOCAL_MACHINE`)

## Note importanti

- **Eseguire come amministratore**: Il comando `jv use` richiede privilegi amministratore. Apri il terminale come amministratore prima di usarlo.
- **Riavvio applicazioni**: Dopo aver cambiato versione, potrebbe essere necessario riavviare il terminale o le applicazioni per vedere le modifiche.
- **Backup**: Il tool rimuove automaticamente i vecchi path Java dal PATH per evitare conflitti.

## Risoluzione problemi

### "Error: failed to open registry key (run as administrator)"

Apri il terminale (CMD o PowerShell) con privilegi di amministratore:
- Cerca "cmd" o "PowerShell" nel menu Start
- Click destro → "Esegui come amministratore"

### "No Java installations found"

Se hai Java installato ma non viene rilevato:
1. Verifica che Java sia in una delle directory standard
2. Oppure aggiungi il path manualmente: `jv add C:\path\to\java`

### "Invalid Java installation path"

Assicurati che il path contenga la directory con `bin\java.exe`:
- ✅ Corretto: `C:\Program Files\Java\jdk-17`
- ❌ Sbagliato: `C:\Program Files\Java\jdk-17\bin`

## Struttura del progetto

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

## Contribuire

Pull request e segnalazioni di bug sono benvenute!

## Licenza

MIT License - Usa liberamente per progetti personali e commerciali.

---

**Fatto con ❤️ per semplificare lo sviluppo Java su Windows**
