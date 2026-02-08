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
# Register your keys (env var name directly)
sekret add OPENAI_API_KEY
sekret add ANTHROPIC_API_KEY

# Or use built-in shorthands
sekret add openai       # → OPENAI_API_KEY
sekret add anthropic    # → ANTHROPIC_API_KEY

# Add to .zshrc (replace existing export statements)
echo 'eval "$(sekret env)"' >> ~/.zshrc
source ~/.zshrc

# Done. Everything works as before.
```

## Commands

| Command | Description |
|---------|-------------|
| `sekret add <ENV_VAR>` | Register a new API key (interactive input) |
| `sekret list` | List registered keys (values are masked) |
| `sekret set <ENV_VAR>` | Update an existing key |
| `sekret remove <ENV_VAR>` | Remove a key (with confirmation) |
| `sekret env` | Output all keys as `export` statements |

## Built-in Shorthands

For common services, you can use shorthand names instead of full env var names:

| Shorthand | Env Variable | Key Prefix |
|-----------|-------------|------------|
| `openai` | `OPENAI_API_KEY` | `sk-` / `sk-proj-` |
| `anthropic` | `ANTHROPIC_API_KEY` | `sk-ant-` |
| `gemini` | `GEMINI_API_KEY` | `AIza` |
| `github` | `GITHUB_TOKEN` | `ghp_` / `github_pat_` |
| `groq` | `GROQ_API_KEY` | `gsk_` |

For any other key, use the env var name directly:

```bash
sekret add MY_SERVICE_KEY
```

## How It Works

- **Key values** are stored in the OS keychain (OS-level encryption)
- **Metadata** (registered env var list) is stored in `~/.config/sekret/config.json`
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
