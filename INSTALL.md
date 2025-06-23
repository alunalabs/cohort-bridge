# Installation Guide - Cohort Tokenize CLI

## 🚀 Quick Install

### Option 1: Install from Source (Recommended)

```bash
# Install directly from GitHub
go install github.com/auroradata-ai/cohort-bridge/cmd/cohort-tokenize@latest

# Or clone and install locally
git clone https://github.com/auroradata-ai/cohort-bridge.git
cd cohort-bridge
go install ./cmd/cohort-tokenize
```

### Option 2: Download Binary

1. Go to the [Releases page](https://github.com/auroradata-ai/cohort-bridge/releases)
2. Download the binary for your platform
3. Move it to a directory in your PATH

### Option 3: Build from Source

```bash
git clone https://github.com/auroradata-ai/cohort-bridge.git
cd cohort-bridge
go build -o cohort-tokenize cmd/cohort-tokenize/main.go

# Move to PATH (optional)
sudo mv cohort-tokenize /usr/local/bin/  # Linux/macOS
# or
move cohort-tokenize.exe C:\Windows\System32\  # Windows
```

## 📋 Prerequisites

- Go 1.21 or later
- PostgreSQL client (optional, for database mode)

## 🔧 Usage

After installation, you can use the tool from anywhere:

```bash
# Show help
cohort-tokenize -help

# Tokenize CSV file
cohort-tokenize -input data.csv -output tokens.json

# Tokenize from database
cohort-tokenize -database -main-config postgres.yaml -output tokens.json

# Interactive mode
cohort-tokenize -interactive

# Show version
cohort-tokenize -version
```

## 🏗️ Development Installation

For development:

```bash
git clone https://github.com/auroradata-ai/cohort-bridge.git
cd cohort-bridge

# Install dependencies
go mod download

# Run directly
go run cmd/cohort-tokenize/main.go -help

# Build for development
go build -o cohort-tokenize cmd/cohort-tokenize/main.go
```

## 🐳 Docker Installation

```bash
# Build Docker image
docker build -t cohort-tokenize .

# Run in container
docker run --rm -v $(pwd):/data cohort-tokenize -input /data/input.csv -output /data/tokens.json
```

## ✅ Verification

Test your installation:

```bash
cohort-tokenize -version
```

You should see version information printed to the console.

## 🔧 Configuration

Place your configuration files in one of these locations:
- Current directory
- `~/.config/cohort-bridge/`
- `/etc/cohort-bridge/`

Example configurations are provided in the repository.

## 🆘 Troubleshooting

### Command not found
- Ensure `$GOPATH/bin` is in your PATH
- Try `go env GOPATH` to find your Go path
- Add `export PATH=$PATH:$(go env GOPATH)/bin` to your shell profile

### Permission denied
- On Unix systems, make the binary executable: `chmod +x cohort-tokenize`
- On Windows, ensure the .exe extension is present

### Database connection issues
- Verify your network connectivity
- Check firewall settings for PostgreSQL port (5432)
- Ensure SSL settings match your database configuration

## 📚 Next Steps

1. Read the [README.md](README.md) for usage examples
2. Check the [examples/](examples/) directory for sample configurations
3. Review [SECURITY.md](SECURITY.md) for privacy considerations 