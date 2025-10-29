# GitHub Configuration

Questa directory contiene la configurazione per GitHub Actions e Issue Templates.

## Struttura

```
.github/
├── workflows/
│   └── release.yml          # Pipeline automatica per releases
├── ISSUE_TEMPLATE/
│   ├── bug_report.md        # Template per segnalazioni bug
│   └── feature_request.md   # Template per richieste funzionalità
└── README.md                # Questo file
```

## GitHub Actions Workflows

### release.yml

Pipeline automatica che:
- Si attiva quando crei un tag `v*` (es: `v1.0.0`)
- Compila `jv.exe` per Windows (amd64)
- Crea checksums SHA256
- Crea automaticamente una GitHub Release con gli asset

**Come usare:**
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

Vedi [RELEASE.md](../RELEASE.md) per dettagli completi.

## Issue Templates

Quando gli utenti creano una issue, possono scegliere tra:
- **Bug Report**: Per segnalare problemi
- **Feature Request**: Per suggerire nuove funzionalità

I template guidano gli utenti a fornire tutte le informazioni necessarie.

## Permessi Richiesti

Il workflow di release richiede:
- `contents: write` - Per creare releases e upload assets

Questo è già configurato nel file `release.yml`.

## Secrets

Il workflow usa `GITHUB_TOKEN` che è automaticamente fornito da GitHub Actions.
Non servono secret aggiuntivi.
