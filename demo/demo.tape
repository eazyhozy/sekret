# sekret demo — register a key and load as env var

Output demo/demo.gif

Set Shell "zsh"
Set Theme { "name": "sekret", "black": "#45475a", "red": "#f38ba8", "green": "#a6e3a1", "yellow": "#f9e2af", "blue": "#89b4fa", "magenta": "#cba6f7", "cyan": "#94e2d5", "white": "#cdd6f4", "brightBlack": "#585b70", "brightRed": "#f38ba8", "brightGreen": "#a6e3a1", "brightYellow": "#f9e2af", "brightBlue": "#89b4fa", "brightMagenta": "#cba6f7", "brightCyan": "#94e2d5", "brightWhite": "#cdd6f4", "background": "#0e0e16", "foreground": "#cdd6f4", "selection": "#313244", "cursor": "#f5e0dc" }
Set FontFamily "JetBrains Mono"
Set FontSize 22
Set LineHeight 1.4
Set Width 1600
Set Height 800
Set Padding 40
Set TypingSpeed 40ms
Set Framerate 50

# Step 1: Register Anthropic key (shorthand)
Type "sekret add anthropic"
Enter
Sleep 1s

# Confirm shorthand
Type "y"
Enter
Sleep 500ms

# Enter API key (masked input)
Type "sk-ant-sekret-v0.2.0"
Enter
Sleep 1.5s

# Step 2: List registered keys
Type "sekret list"
Enter
Sleep 3s

# Step 3: Load as env vars
Type 'eval "$(sekret env)"'
Enter
Sleep 1s

# Step 4: Verify
Type "echo $ANTHROPIC_API_KEY"
Enter
Sleep 3s
