# Demo SVG Generation

Animated SVG demo for the project README, generated via [svg-term-cli](https://github.com/marionebl/svg-term-cli).

## Prerequisites

```bash
brew install asciinema
npm install -g svg-term-cli
```

## Pipeline

```
list_{1,2,3}.txt  ->  generate_cast.go  ->  demo.cast  ->  svg-term  ->  demo.svg
```

## Regenerate

```bash
# 1. Generate asciicast (.cast)
go run demo/generate_cast.go > demo/demo.cast

# 2. Generate SVG
svg-term --in demo/demo.cast --out demo/demo.svg --window
```

## Files

| File | Description |
|------|-------------|
| `list_{1,2,3}.txt` | Captured `sekret list` output per round (input for generate_cast.go) |
| `generate_cast.go` | Generates asciicast v2 (.cast) with typing animation |
| `demo.cast` | Generated asciicast file (input for svg-term) |
| `demo.svg` | Final animated SVG used in README |
