# Quick Start Guide

## Prerequisiti

1. **Installa Go** (se non lo hai già):
   - Scarica da: https://go.dev/dl/
   - Installa e verifica: `go version`

## Build & Installazione (3 passi)

### 1. Compila il progetto

```bash
# Doppio click su build.bat
# OPPURE da terminale:
build.bat
```

### 2. Installa il tool

Scegli una delle opzioni:

**Opzione A - System-wide (consigliato)**
```cmd
copy jv.exe C:\Windows\System32\
```

**Opzione B - Directory custom nel PATH**
```cmd
# Crea una directory per i tuoi tool
mkdir C:\tools
copy jv.exe C:\tools\

# Aggiungi C:\tools al PATH:
# 1. Cerca "variabili d'ambiente" nel menu Start
# 2. Click su "Modifica variabili d'ambiente di sistema"
# 3. Click "Variabili d'ambiente"
# 4. Sotto "Variabili di sistema", seleziona "Path" e click "Modifica"
# 5. Click "Nuovo" e aggiungi: C:\tools
# 6. Click OK su tutte le finestre
```

### 3. Usa il tool

```bash
# Lista versioni Java disponibili
jv list

# Cambia a Java 17 (ESEGUI COME AMMINISTRATORE!)
jv use 17

# Verifica versione corrente
jv current
java -version
```

## Comandi Principali

```bash
# Gestione versioni
jv list                     # Lista tutte le versioni disponibili
jv use 17                   # Cambia a Java 17
jv current                  # Mostra versione corrente

# Installazioni custom (specifiche)
jv add C:\my\jdk-17         # Aggiungi UNA installazione specifica

# Search paths (scansione directory)
jv add-path C:\DevTools\Java    # Aggiungi directory da scansionare
jv list-paths                   # Mostra tutti i search paths
jv remove-path C:\DevTools\Java # Rimuovi search path

# Aiuto
jv help                     # Mostra aiuto completo
```

**Differenza tra `add` e `add-path`:**
- `add`: Per una singola installazione (es: `C:\custom\jdk-17`)
- `add-path`: Per una directory contenente multiple versioni (es: `C:\DevTools\Java` che contiene jdk-17, jdk-21, ecc.)

## IMPORTANTE: Privilegi Amministratore

Il comando `jv use` richiede privilegi amministratore perché modifica le variabili d'ambiente di sistema.

**Come eseguire come amministratore:**
- Click destro su "CMD" o "PowerShell" nel menu Start
- Seleziona "Esegui come amministratore"
- Esegui `jv use <version>`

## Verifica Installazione

Dopo aver installato jv.exe:

```bash
# Apri un NUOVO terminale e prova:
jv help
```

Se vedi il messaggio di aiuto, l'installazione è riuscita!

## Problemi comuni

**"jv non è riconosciuto..."**
- Assicurati di aver aggiunto jv.exe al PATH
- Riavvia il terminale dopo aver modificato il PATH

**"failed to open registry key"**
- Esegui il terminale come amministratore

**"No Java installations found"**
- Aggiungi manualmente: `jv add C:\path\to\jdk`

---

Per dettagli completi, vedi [README.md](README.md)
