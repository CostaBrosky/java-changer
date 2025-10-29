# Guida di Installazione

Guida passo-passo per installare Java Version Switcher (jv) su Windows.

## Metodo 1: Download da GitHub Releases (Consigliato)

### Passo 1: Download

1. Apri il browser e vai su: https://github.com/USERNAME/java-changer/releases
2. Clicca sull'ultima release (quella in alto)
3. Sotto "Assets", clicca su **`jv.exe`** per scaricare l'eseguibile

### Passo 2: Verifica (Opzionale ma Consigliato)

Verifica che il file scaricato sia autentico usando SHA256:

```powershell
# In PowerShell, vai nella cartella Downloads
cd ~\Downloads

# Calcola il checksum
Get-FileHash jv.exe -Algorithm SHA256

# Confronta l'output con il checksum in checksums.txt nella release
```

### Passo 3: Installazione

**Opzione A - System-wide (Consigliato):**

1. Apri File Explorer come Amministratore
2. Copia `jv.exe` in `C:\Windows\System32\`
3. Fatto! Ora puoi usare `jv` da qualsiasi terminale

**Opzione B - Directory Custom:**

1. Crea una directory per i tuoi tool, per esempio:
   ```cmd
   mkdir C:\tools
   ```

2. Copia `jv.exe` in `C:\tools\`

3. Aggiungi `C:\tools` al PATH:
   - Premi `Win + X` → Seleziona "Sistema"
   - Click su "Impostazioni di sistema avanzate"
   - Click su "Variabili d'ambiente"
   - Sotto "Variabili di sistema", seleziona "Path" e click "Modifica"
   - Click "Nuovo" e aggiungi: `C:\tools`
   - Click "OK" su tutte le finestre
   - **Riavvia il terminale** per applicare le modifiche

### Passo 4: Verifica Installazione

Apri un nuovo terminale (CMD o PowerShell) e prova:

```cmd
jv version
jv help
```

Se vedi l'output, l'installazione è riuscita! 🎉

## Metodo 2: Compilazione da Sorgente

### Prerequisiti

- Go 1.21 o superiore installato ([Download](https://go.dev/dl/))
- Git installato ([Download](https://git-scm.com/download/win))

### Passi

1. **Clone del repository:**
   ```bash
   git clone https://github.com/USERNAME/java-changer.git
   cd java-changer
   ```

2. **Download dipendenze:**
   ```bash
   go mod download
   ```

3. **Compila:**
   ```bash
   go build -ldflags="-s -w" -o jv.exe ./cmd/jv
   ```

4. **Verifica:**
   ```bash
   .\jv.exe version
   ```

5. **Installa** (copia in una directory nel PATH, vedi Metodo 1, Passo 3)

## Utilizzo Base

Dopo l'installazione, ecco i comandi base:

```bash
# Lista versioni Java disponibili
jv list

# Cambia a Java 17 (ESEGUI TERMINALE COME AMMINISTRATORE!)
jv use 17

# Mostra versione corrente
jv current

# Aggiungi directory custom da scansionare
jv add-path C:\DevTools\Java
```

**IMPORTANTE:** Il comando `jv use` richiede privilegi amministratore perché modifica le variabili d'ambiente di sistema.

## Come Eseguire come Amministratore

### CMD/PowerShell
1. Cerca "cmd" o "PowerShell" nel menu Start
2. **Click destro** → "Esegui come amministratore"
3. Esegui i comandi `jv`

### Windows Terminal
1. Apri Windows Terminal
2. Click sulla freccia ▼ vicino al tab
3. Tieni premuto `Ctrl` e click sul profilo (CMD o PowerShell)
4. Si aprirà con privilegi amministratore

## Troubleshooting

### "jv non è riconosciuto come comando interno o esterno"

**Soluzione:**
- Assicurati di aver aggiunto `jv.exe` al PATH
- Riavvia il terminale dopo aver modificato il PATH
- Verifica con: `where jv` (dovrebbe mostrare il path a jv.exe)

### "failed to open registry key (run as administrator)"

**Soluzione:**
- Il comando `jv use` richiede privilegi amministratore
- Esegui il terminale come amministratore (vedi sopra)

### "No Java installations found"

**Soluzione:**
- Verifica che Java sia installato sul tuo sistema
- Se Java è in una directory non standard, aggiungila:
  ```bash
  jv add C:\path\to\jdk
  # oppure
  jv add-path C:\path\to\java-directory
  ```

### Windows Defender blocca l'eseguibile

**Soluzione:**
- Questo può succedere con eseguibili scaricati da internet
- Verifica il checksum SHA256 per assicurarti che sia autentico
- Aggiungi un'eccezione in Windows Defender se necessario
- In alternativa, compila da sorgente (Metodo 2)

### Le modifiche alle variabili d'ambiente non si applicano

**Soluzione:**
- Dopo aver eseguito `jv use`, riavvia:
  - Il terminale corrente (chiudi e riapri)
  - Le applicazioni che devono usare Java (es: IDE)
- In casi estremi, riavvia Windows

## Disinstallazione

Per rimuovere jv:

1. Elimina `jv.exe` da dove l'hai installato:
   ```cmd
   # Se installato in System32
   del C:\Windows\System32\jv.exe

   # Se installato in directory custom
   del C:\tools\jv.exe
   ```

2. (Opzionale) Rimuovi la configurazione:
   ```cmd
   del %USERPROFILE%\.javarc
   ```

## Prossimi Passi

Una volta installato, leggi la [Guida Rapida](QUICKSTART.md) per imparare ad usare jv efficacemente.

Per documentazione completa, vedi il [README](README.md).

---

**Buon switching! ☕**
