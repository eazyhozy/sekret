# sekret demo — register keys and load as env vars

Output demo/demo.gif

Set Shell "zsh"
Set Theme "Catppuccin Mocha"
Set FontSize 22
Set Width 1200
Set Height 600
Set Padding 30
Set TypingSpeed 40ms
Set Framerate 50

# Step 1: Register OpenAI key (shorthand)
Type "sekret add openai"
Enter
Sleep 1s

# Confirm shorthand
Type "y"
Enter
Sleep 500ms

# Enter API key (masked input)
Type "sk-proj-sekret-v0.2.0"
Enter
Sleep 1.5s

# Step 2: Register Anthropic key (shorthand)
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

# Step 3: Register custom key (direct env var)
Type "sekret add MY_SERVICE_KEY"
Enter
Sleep 500ms

# Enter API key (masked input)
Type "msk-sekret-v0.2.0"
Enter
Sleep 1.5s

# Step 4: List registered keys
Type "sekret list"
Enter
Sleep 3s

# Step 5: Load as env vars
Type 'eval "$(sekret env)"'
Enter
Sleep 1s

# Step 6: Verify
Type "echo $OPENAI_API_KEY"
Enter
Sleep 3s
