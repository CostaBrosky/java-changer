# Release Guide

Questa guida spiega come rilasciare una nuova versione del Java Version Switcher.

## Prerequisiti

1. Hai accesso al repository GitHub
2. Hai i permessi per creare release e tags
3. Il codice è stato testato e funziona correttamente

## Processo di Release

### 1. Preparare il codice

Assicurati che tutto il codice sia committato e pushato su `main`:

```bash
git add .
git commit -m "Prepare for release vX.Y.Z"
git push origin main
```

### 2. Creare un tag di versione

Usa il formato [Semantic Versioning](https://semver.org/): `vMAJOR.MINOR.PATCH`

- **MAJOR**: Cambiamenti incompatibili con versioni precedenti
- **MINOR**: Nuove funzionalità, backward compatible
- **PATCH**: Bug fixes, backward compatible

Esempi:
- `v1.0.0` - Prima release stabile
- `v1.1.0` - Aggiunta funzionalità search paths
- `v1.0.1` - Bug fix

```bash
# Crea il tag localmente
git tag -a v1.0.0 -m "Release v1.0.0: Initial stable release"

# Pusha il tag su GitHub
git push origin v1.0.0
```

### 3. GitHub Actions si attiva automaticamente

Quando pushhi il tag, GitHub Actions:

1. ✅ Compila `jv.exe` per Windows (amd64)
2. ✅ Testa l'eseguibile
3. ✅ Crea un archivio ZIP con exe + documentazione
4. ✅ Genera checksums SHA256
5. ✅ Crea una GitHub Release con tutti gli assets

Il processo richiede circa 2-5 minuti.

### 4. Verifica la Release

1. Vai su: `https://github.com/USERNAME/java-changer/releases`
2. Verifica che la release sia stata creata con:
   - `jv.exe` - Eseguibile standalone
   - `jv-vX.Y.Z-windows-amd64.zip` - Archivio completo
   - `checksums.txt` - Checksums SHA256
3. Scarica `jv.exe` e testalo

### 5. (Opzionale) Aggiorna le Release Notes

Se necessario, modifica la descrizione della release su GitHub per aggiungere:
- Changelog dettagliato
- Breaking changes
- Note di migrazione
- Screenshot o esempi

## Release Manuale (senza tag)

Se vuoi creare una release senza creare un tag, puoi attivare manualmente il workflow:

1. Vai su GitHub → Actions
2. Seleziona "Release" workflow
3. Click "Run workflow"
4. Seleziona il branch e click "Run workflow"

**Nota:** Le release manuali avranno versione `v0.0.0-dev` e non creeranno una GitHub Release automatica.

## Rollback di una Release

Se devi rimuovere una release:

### Rimuovere il tag

```bash
# Rimuovi il tag localmente
git tag -d v1.0.0

# Rimuovi il tag da GitHub
git push origin :refs/tags/v1.0.0
```

### Rimuovere la Release su GitHub

1. Vai su Releases
2. Click sulla release da rimuovere
3. Click "Delete release"

## Build Locale per Test

Se vuoi testare il build prima di rilasciare:

```bash
# Build con versione specifica
go build -ldflags="-s -w -X main.Version=v1.0.0-test" -o jv.exe ./cmd/jv

# Testa
.\jv.exe version
.\jv.exe help
```

## Checklist Pre-Release

Prima di creare un tag di release, verifica:

- [ ] Tutti i test passano
- [ ] La documentazione è aggiornata (README.md, QUICKSTART.md)
- [ ] Il CHANGELOG è aggiornato (se presente)
- [ ] Non ci sono file sensibili committati (.env, credentials, ecc.)
- [ ] Il codice compila senza errori o warning
- [ ] Hai testato manualmente le funzionalità principali
- [ ] Il numero di versione segue Semantic Versioning

## Esempio: Prima Release (v1.0.0)

```bash
# 1. Verifica lo stato
git status
git log --oneline -5

# 2. Crea e pusha il tag
git tag -a v1.0.0 -m "Release v1.0.0

- Auto-detection of Java installations
- Custom search paths support
- Permanent system environment variable changes
- Easy version switching with single command
"

git push origin v1.0.0

# 3. Attendi la pipeline (2-5 minuti)
# 4. Verifica su GitHub Releases
# 5. Scarica e testa jv.exe
```

## Troubleshooting

### La pipeline fallisce

1. Controlla i log su GitHub Actions
2. Verifica che `go.mod` sia corretto
3. Assicurati che tutti i file necessari esistano

### Il build non contiene la versione corretta

La versione viene iniettata tramite ldflags. Verifica che:
- Il tag inizi con `v` (es: `v1.0.0`, non `1.0.0`)
- La variabile `Version` sia definita in `cmd/jv/main.go`

### La release non viene creata automaticamente

La release automatica funziona solo per tag che iniziano con `v`. Verifica:
- Il tag è nel formato `v*` (es: `v1.0.0`)
- Hai i permessi per creare release
- Il workflow ha `permissions: contents: write`

---

**Buon release! 🚀**
