# sekret

Secure your API keys in OS keychain, load them as env vars. No more plaintext secrets in `.zshrc`.

## Why sekret?

Most developers store API keys as plaintext `export` statements in `.zshrc` or `.env` files. With the growing number of AI tools (Claude Code, Cursor, Aider, Gemini CLI, etc.), managing these keys securely becomes increasingly important.

sekret stores your keys in the OS keychain (macOS Keychain, GNOME Keyring) and loads them as environment variables — with zero change to your daily workflow.

## Installation

```bash
go install github.com/eazyhozy/sekret@latest
```

## Quick Start

```bash
# Register your keys
sekret add openai
sekret add anthropic

# Add to .zshrc (replace existing export statements)
echo 'eval "$(sekret env)"' >> ~/.zshrc
source ~/.zshrc

# Done. Everything works as before.
```

## Commands

| Command | Description |
|---------|-------------|
| `sekret add <name>` | Register a new API key (interactive input) |
| `sekret list` | List registered keys (values are masked) |
| `sekret set <name>` | Update an existing key |
| `sekret remove <name>` | Remove a key (with confirmation) |
| `sekret env` | Output all keys as `export` statements |

## Built-in Key Registry

Keys are automatically mapped to environment variables:

| Name | Env Variable | Key Prefix |
|------|-------------|------------|
| `openai` | `OPENAI_API_KEY` | `sk-` / `sk-proj-` |
| `anthropic` | `ANTHROPIC_API_KEY` | `sk-ant-` |
| `gemini` | `GEMINI_API_KEY` | `AIza` |
| `github` | `GITHUB_TOKEN` | `ghp_` / `github_pat_` |
| `groq` | `GROQ_API_KEY` | `gsk_` |

For custom keys, specify the env var name:

```bash
sekret add my-service --env MY_SERVICE_KEY
```

## How It Works

- **Key values** are stored in the OS keychain (OS-level encryption)
- **Metadata** (name-to-env-var mapping) is stored in `~/.config/sekret/config.json`
- Key values are **never written to any file**
- Key input is always interactive (never accepted as CLI arguments, protecting shell history)

## Platform Support

| OS | Backend | Status |
|----|---------|--------|
| macOS | Keychain | Supported |
| Linux (Desktop) | GNOME Keyring / KWallet | Supported |
| Windows | Credential Manager | Planned |
| Linux (Headless) | Encrypted file fallback | Planned |

## License

[MIT](LICENSE)
