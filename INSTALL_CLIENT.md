# Client Install Guide (Phase 1)

## Persistent config (one-time)

```bash
gorenel config init
# or
gorenel config set api_key gk_xxx
gorenel config set port 3000
gorenel config set server wss://gorenel.site/tunnel/connect
```

Then connect without long command:

```bash
gorenel connect
# optional override
gorenel connect --port 4000
```

## Windows installer (Inno Setup)

1. Install Inno Setup 6
2. Ensure client binary exists at `bin/gorenel-client-windows-amd64.exe`
3. Open `packaging/windows/gorenel.iss`
4. Build installer

Installer output: `packaging/windows/gorenel-setup.exe`

## macOS Homebrew

Formula template: `packaging/homebrew/gorenel.rb`

1. Replace SHA256 placeholders
2. Publish formula to your tap repository
3. Users install with:

```bash
brew install <your-tap>/gorenel
```

## Validation

`ash
gorenel config validate
` 

## Homebrew formula SHA automation

`ash
bash scripts/update_homebrew_formula.sh 1.0.0
` 

