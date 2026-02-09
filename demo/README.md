# Demo

Animated SVG generated via [svg-term-cli](https://github.com/marionebl/svg-term-cli).

## Regenerate

```bash
go run demo/generate_cast.go > demo/demo.cast
svg-term --in demo/demo.cast --out demo/demo.svg --window
```
