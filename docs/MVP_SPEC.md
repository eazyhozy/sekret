# sekret — MVP Spec

> A CLI tool to securely and conveniently manage local API keys in one place.

---

## 1. Project Overview

### The Problem

API keys on developers' local machines are **scattered in plaintext** across `.zshrc`, `.env`, config files, etc.
With the explosive growth of AI tools (Claude Code, Cursor, Aider, Gemini CLI, etc.), the number of keys to manage is rapidly increasing, yet most developers still rely on `export` statements in `.zshrc`.

### Core Values

| Principle | Description |
|-----------|-------------|
| **Security** | Store API keys encrypted in the OS keychain, not in plaintext files |
| **Convenience** | No change to existing workflow. Just replace one line in `.zshrc` |
| **Centralization** | Register, view, update, and delete all API keys from a single place |

### Target Users

- Developers who actively use AI tools
- Developers managing multiple API keys (OpenAI, Anthropic, Google, GitHub, etc.) locally
- Developers who care about security but don't want complex setups

---

## 2. User Journey

### Initial Setup (one-time, 2 minutes)

```bash
# Install
$ brew install sekret        # macOS
$ go install github.com/xxx/sekret@latest  # or Go install

# Register existing keys
$ sekret add OPENAI_API_KEY
  API Key: ••••••••••••••••
  Saved to OS keychain (OPENAI_API_KEY)

$ sekret add ANTHROPIC_API_KEY
  API Key: ••••••••••••••••
  Saved to OS keychain (ANTHROPIC_API_KEY)

# Built-in shorthands also work
$ sekret add openai       # → OPENAI_API_KEY
$ sekret add anthropic    # → ANTHROPIC_API_KEY

# Replace in .zshrc (remove existing export statements first)
$ echo 'eval "$(sekret env)"' >> ~/.zshrc
$ source ~/.zshrc
```

### Daily Use (no change)

```bash
# sekret env runs automatically when opening a terminal, loading env vars
# Exactly the same experience as before

$ claude-code          # just works
$ aider --model sonnet # just works
$ curl -H "Authorization: Bearer $OPENAI_API_KEY" ...  # just works
```

### Key Management (occasional)

```bash
$ sekret list                    # list registered keys
$ sekret add GEMINI_API_KEY      # add a new key
$ sekret set OPENAI_API_KEY      # update an existing key
$ sekret remove OLD_SERVICE_KEY  # delete a key
```

---

## 3. MVP Commands in Detail

### `sekret add <ENV_VAR>`

Register a new API key.

```
$ sekret add OPENAI_API_KEY
  API Key: ••••••••••••••••
  Saved to OS keychain (OPENAI_API_KEY)
```

**Behavior:**
1. Accept env var name as argument (e.g. `OPENAI_API_KEY`)
2. Prompt for key input interactively (masked input)
3. Store in OS keychain

**Built-in Shorthands:**

For common services, shorthand names are accepted and expanded with a confirmation prompt:

| Shorthand | Env Variable | Key Format Pattern |
|-----------|-------------|-------------------|
| `openai` | `OPENAI_API_KEY` | `sk-` or `sk-proj-` |
| `anthropic` | `ANTHROPIC_API_KEY` | `sk-ant-` |
| `gemini` | `GEMINI_API_KEY` | `AIza` |
| `github` | `GITHUB_TOKEN` | `ghp_` / `github_pat_` |
| `groq` | `GROQ_API_KEY` | `gsk_` |

- Any valid env var name can be registered directly: `sekret add MY_SERVICE_KEY`
- Simple format validation for known keys (prevents paste errors)

### `sekret list`

List all registered keys.

```
$ sekret list

  Env Variable         Key Preview    Added
  ─────────────────────────────────────────────
  OPENAI_API_KEY       sk-...Qx7f     3 months ago
  ANTHROPIC_API_KEY    sk-ant-...w2   1 month ago
  GEMINI_API_KEY       AIza...8kP     2 weeks ago
```

**Behavior:**
- Key values show only the prefix + last 4 characters (never expose the full key)

### `sekret set <ENV_VAR>`

Update an existing key.

```
$ sekret set OPENAI_API_KEY
  Current: sk-...Qx7f
  New API Key: ••••••••••••••••
  Updated
```

### `sekret remove <ENV_VAR>`

Delete a key.

```
$ sekret remove OLD_SERVICE_KEY
  Remove 'OLD_SERVICE_KEY'? [y/N]: y
  Removed
```

### `sekret env`

Output all registered keys as `export` statements. Used with `eval` in `.zshrc`.

```
$ sekret env
export OPENAI_API_KEY="sk-xxx..."
export ANTHROPIC_API_KEY="sk-ant-xxx..."
export GEMINI_API_KEY="AIza..."
```

**Behavior:**
- Read all sekret entries from the OS keychain and generate export statements
- Shell-safe escaping
- Performance is critical: runs every time a terminal starts, target **< 100ms**

### `sekret run -- <command>`

Run a command with keys injected only into that process (optional advanced feature).

```
$ sekret run -- claude-code
$ sekret run --only openai,anthropic -- python script.py
```

**Behavior:**
- Inject keys only into the child process without exposing them in the current shell
- Use `--only` flag to select specific keys
- More secure than `eval "$(sekret env)"` (process isolation)
- **Nice-to-have for MVP, not required**

---

## 4. Security Design

### Storage

```
┌──────────────────────────────────────────┐
│        OS Keychain                       │
│  ┌────────────────────────────────────┐  │
│  │ service: "sekret"                  │  │
│  │ account: "OPENAI_API_KEY"          │  │
│  │ password: "sk-xxx..."              │  │
│  └────────────────────────────────────┘  │
│  ┌────────────────────────────────────┐  │
│  │ service: "sekret"                  │  │
│  │ account: "ANTHROPIC_API_KEY"       │  │
│  │ password: "sk-ant-xxx..."          │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

- **Key values**: Stored in OS keychain (OS-level encryption)
- **Metadata**: (registered env var list, timestamps) stored in `~/.config/sekret/config.json`
  - Metadata never contains key values
- Key values are never written to plaintext files

### Security Principles

| Principle | Implementation |
|-----------|---------------|
| Keys are always stored encrypted | OS keychain only, no custom file storage |
| Minimize plaintext exposure | Masked in `list`, `env` output is for eval use only |
| Format validation | Known keys are validated by prefix pattern to prevent mistakes |
| Safe input | Terminal echo disabled during key input |
| Shell history protection | Keys are never accepted as CLI arguments (always interactive input) |

### Security Limitations (documented transparently)

- When using `eval "$(sekret env)"`, keys exist as environment variables in that shell session. Other processes from the same user can access them via `/proc/<pid>/environ`, etc.
- This is the same level of exposure as the traditional `.zshrc` export approach. For stronger isolation, use `sekret run --`.
- sekret's primary goal is to **eliminate the risk of storing keys in plaintext files**. Runtime environment variable exposure is a separate concern.

---

## 5. Tech Stack

### Language: Go

**Rationale:**
- Single binary distribution (no dependencies, simple installation)
- Easy cross-compilation (macOS, Linux, Windows)
- Cross-platform OS keychain support via `go-keyring`
- Mature CLI ecosystem (cobra, bubbletea, etc.)
- Fast build times — ideal for side projects

### Key Dependencies

| Library | Purpose |
|---------|---------|
| `github.com/zalando/go-keyring` | OS keychain access (macOS/Linux/Windows) |
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/term` | Secure terminal input (echo disabled) |

### Metadata Storage

```
~/.config/sekret/config.json
```

```json
{
  "version": 1,
  "keys": [
    {
      "name": "",
      "env_var": "OPENAI_API_KEY",
      "added_at": "2025-02-01T09:00:00Z"
    },
    {
      "name": "",
      "env_var": "ANTHROPIC_API_KEY",
      "added_at": "2025-02-05T14:30:00Z"
    }
  ]
}
```

- Key values are never stored here
- The `name` field exists for backward compatibility with legacy entries; new entries have it empty

---

## 6. Platform Support

| OS | Keychain Backend | Priority |
|----|-----------------|----------|
| macOS | Keychain | MVP required |
| Linux (Desktop) | GNOME Keyring / KWallet (Secret Service API) | MVP required |
| Linux (Headless) | Encrypted file fallback (post-MVP) | v0.2 |
| Windows | Credential Manager | v0.2 |

---

## 7. Performance Requirements

| Operation | Target |
|-----------|--------|
| `sekret env` (10 keys) | < 100ms |
| `sekret add` | < 500ms (interactive, so more lenient) |
| `sekret list` | < 200ms |

`sekret env` runs every time a terminal starts, so its performance directly impacts user experience.
Go's fast startup time + local OS keychain access should make this easily achievable.

---

## 8. MVP Scope

### MVP (v0.1) — Included

- `sekret add <ENV_VAR>` — Interactive key registration
- `sekret list` — List registered keys (masked)
- `sekret set <ENV_VAR>` — Update key
- `sekret remove <ENV_VAR>` — Delete key
- `sekret env` — Export environment variables
- Built-in shorthands (openai, anthropic, gemini, github, groq)
- macOS + Linux Desktop support
- Homebrew distribution

### v0.2 — Future

- [ ] `sekret run -- <command>` — Process-isolated execution
- [ ] `sekret scan` — Detect plaintext keys in local files (`.zshrc`, `.env`, config, etc.)
- [ ] `sekret doctor` — Diagnose required keys per AI tool
- [ ] `sekret import` — Auto-migrate from existing `export` statements in `.zshrc`
- [ ] Windows support
- [ ] Linux headless fallback
- [ ] MCP server mode (direct AI tool integration)

### v0.3 — Future

- [ ] Per-project key profiles (`sekret env --profile work`)
- [ ] Key expiration alerts / rotation reminders
- [ ] `sekret audit` — Key access history (which processes accessed keys)
- [ ] Team sharing (encrypted key export/import)

---

## 9. Competitive Advantage Summary

| | .zshrc export | 1Password CLI | pass + GPG | dotenvx | **sekret** |
|---|---|---|---|---|---|
| Setup difficulty | None | Medium | High | Low | **Low** |
| Security | No (plaintext) | Strong | Strong | Encrypted | **OS keychain** |
| Cost | Free | Paid | Free | Free | **Free** |
| Workflow change | None | Significant | Significant | Moderate | **None** |
| API key focused | No | No | No | No | **Yes** |
| AI tool awareness | No | No | No | No | **Yes** |
| Scope | - | General secrets | General passwords | Project .env | **Local API keys** |

---

## 10. Success Metrics

| Metric | Target (3 months post-launch) |
|--------|-------------------------------|
| GitHub Stars | 500+ |
| Homebrew installs | 1,000+ |
| HN/Reddit reception | 50+ upvotes on Show HN |
| Issue/PR participation | 5+ external contributors |

---

## 11. Development Timeline (2-3 hours/week)

| Week | Work | Deliverable |
|------|------|-------------|
| 1-2 | Project setup + implement `add`, `list`, `env` | Basic working CLI |
| 3-4 | `set`, `remove` + built-in registry + format validation | Full CRUD complete |
| 5-6 | Tests + macOS/Linux CI + README + Homebrew formula | Release-ready |
| 7 | Launch on Show HN / Reddit / GeekNews | Public launch |

**Estimated total development time: 7 weeks (14-21 hours)**
