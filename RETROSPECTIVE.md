# sekret Retrospective — Had I First Asked "Isn't the Simplest Alternative Already Sufficient?"

> A record of building, validating, and deciding the direction of an API key management CLI.

---

## Project Overview

**sekret** — A CLI tool that stores API keys in the OS keychain and loads them as environment variables.

- GitHub: github.com/eazyhozy/sekret
- Language: Go
- Distribution: Homebrew tap (`brew install eazyhozy/sekret/sekret`)
- Duration: Early February 2026 ~ Mid-February 2026 (approx. 2 weeks)

### Core Idea

Instead of placing plaintext keys like `export OPENAI_API_KEY="sk-..."` in `.zshrc`,
store them in the OS keychain (macOS Keychain / GNOME Keyring) and load them with `eval "$(sekret env)"`.

```bash
sekret add openai        # save to keychain
sekret list              # list stored keys
eval "$(sekret env)"     # load as environment variables
```

---

## Timeline

### Idea to MVP (First week of February)

- Started from the inconvenience of managing local API keys
- Project name candidates: `keyring` (conflict) → `lockey` (conflict) → `keybox` (conflict) → `hush` (weak conflict) → **`sekret`** adopted
- Why sekret: no practical conflicts, 6 characters easy to type, SEO-dominable
- Wrote MVP spec → implemented with Claude Code → released v0.0.1 (13 commits)

### v0.2.0 Release (End of first week)

- `sekret scan`: detect plaintext keys in shell config files
- `sekret import`: interactively migrate plaintext keys to keychain
- goreleaser setup, Homebrew tap distribution
- Added demo GIF to README
- Adopted testify, organized test structure

### Validation and Analysis (Second week of February)

- Self-validation began while preparing to promote on Reddit, to colleagues, etc.
- Conducted market analysis, threat model analysis, collected peer feedback
- Reached conclusions on project direction

---

## Technical Takeaways

### Hands-on Go Experience

- First Go project as a Java developer
- Key differences internalized: single binary, `(result, error)` pattern, case-based access control, struct + receiver
- CLI command structure design with cobra

### Distribution Pipeline

- **goreleaser**: tag push → cross-compiled binaries + Homebrew formula auto-generated
- **Homebrew tap**: created `homebrew-sekret` repo, instant distribution without review
- Strategy: start with own tap → challenge homebrew-core after building recognition

### Understanding OS Keychain Security

- **macOS Keychain**: AES-256 encryption, per-app access control, lock integration
- **GNOME Keyring**: AES-128 encryption, no app isolation within the same session (CVE-2018-19358)
- GNOME project's stance: "app isolation within the same session is security theater"
- Both platforms achieve "no plaintext on disk," but at the runtime memory/session level, there is little practical difference

### CLI Design Patterns

- Built-in key registry (shorthand → environment variable name mapping)
- Non-invasive integration into existing shell workflows via `eval "$(cmd)"` pattern
- `scan` / `import` to lower the migration barrier for existing users

---

## Market Analysis

### Target Market Size

- Approximately 4 million OpenAI API developers (as of 2025 DevDay)
- Estimated 5-8 million total LLM API users
- However, "developers who directly use API keys as environment variables in their local terminal" are only a subset

### Shrinking Market Trend

- Major tools like Claude Code, Cursor, and GitHub Copilot are shifting to OAuth/subscription-based authentication
- Areas where API keys are still needed: open-source agents like OpenCode and Aider, direct script calls
- However, even OpenCode and similar tools have built-in `auth.json` management

### Competitive Landscape

| Tool | Characteristics | vs. sekret |
|------|----------------|------------|
| `~/secrets.sh` + `source` | No installation needed, universally known | Replaces 99% of functionality |
| 1Password + `op run` | Team sharing, cross-device, CLI support | Superset (paid $2.99/mo) |
| pass (GPG-based) | Open source, Git sync | Requires GPG setup but feature-rich |
| direnv + `.envrc` | Per-directory env vars | Suited for per-project management |
| HashiCorp Vault | Enterprise secret management | Entirely different category |
| Infisical / Doppler | Team secret management SaaS/open-source | Solves collaborative secret sharing |

---

## Validation Process

### Self-validation

1. **Personal usage**: I was only managing 1 key in plaintext. The narrative "I built this while managing 8 API keys" didn't hold up.
2. **Threat model analysis**: Scenarios where `.zshrc` files get leaked are extremely rare. In an era where macOS FileVault / Linux LUKS are standard, disk theft threats are low. Within the same user session, access rights are identical whether using keychain or plaintext files.
3. **`secrets.sh` comparison**: Putting all keys in `~/secrets.sh` and running `source` is functionally identical to sekret's core feature. No installation or learning required.

### Peer Feedback (2 people)

- Already using 1Password effectively at work (financial sector, isolated network environment)
- In collaborative settings, `.env` files are shared within the team
- Some don't use LLM APIs at all due to isolated network constraints
- Reaction to sekret: "Storing in OS keychain feels somewhat safer, but when you dig in, there might not be a big difference"
- Key question: "Is there a significant advantage over managing with `.env` or `.zshrc`?"

### Positioning Exploration

Angles attempted and their limitations:

1. **"Security" focus**: Examining the threat model reveals marginal real security gains
2. **"Convenience / unification" focus**: `.zshrc` export is already convenient enough; sekret adds another management surface
3. **"Collaborative secret sharing" direction**: Requires server + auth + permissions, fundamentally different architecture. Infisical/Doppler already exist

---

## Conclusion

### Key Findings

Key management falls into two domains:

- **Local/personal**: Adequately solved by `~/secrets.sh`, `.env`, `.zshrc` export
- **Collaborative/deployment/shared**: Requires key sharing, rotation, access control → domain of Vault, Infisical, 1Password, etc.

sekret targeted the former, but lacked sufficient motivation to switch from `secrets.sh`.
Expanding to the latter would require a fundamental architecture change, and powerful existing tools already occupy that space.

### Project Value Assessment

The problem sekret set out to solve — "plaintext API keys in dotfiles" — was **already a sufficiently solved problem** for most developers. I couldn't find a compelling answer to "why should I use this?" rather than "is this more secure?"

### What I Still Gained

- Hands-on Go experience (first Go project as a Java developer)
- goreleaser + Homebrew tap distribution pipeline
- CLI tool design patterns (cobra, eval pattern, built-in registry)
- Deep understanding of OS keychain security architecture
- Open-source project launch preparation experience (Show HN, Reddit strategy)
- **Completing the "idea → build → validate → decide" cycle in just 2 weeks**

---

## Lessons Learned

### 1. Detect "solution looking for a problem" early

sekret was technically well-built, but the intensity of the problem it aimed to solve was weak.
The fact that I was only managing 1 plaintext key was the first warning sign.

### 2. Do the `secrets.sh` test before building

If I had first asked "isn't the simplest alternative already sufficient?", I could have set direction before building. Make it a habit to check whether a single line of bash can solve the same problem before creating a new tool.

### 3. Seek peer feedback when you're uncertain, not when you're confident

Asking "do you need this?" instead of "isn't this great?" yields far more honest answers. Seeking feedback while in a state of doubt actually produces better feedback.

### 4. The validation process itself is valuable

In 2 weeks, I completed one full cycle: idea → MVP → distribution → market analysis → peer feedback → direction decision. Many side projects spend months without this validation. Fast validation is fast learning.

### 5. Knowing when to fold is a skill

Shelving a project isn't failure — it's judgment. Rather than forcing reasons to continue, accepting what the data tells you is the best thing you can do for the next project.
