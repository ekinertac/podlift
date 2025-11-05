# podlift Development

## Quick Start

### Build

```bash
make build
# or
go build -o bin/podlift ./cmd/podlift
```

### Run

```bash
./bin/podlift --help
./bin/podlift version
```

### Test

```bash
make test
# or
go test ./...
```

## Project Structure

```
podlift/
├── cmd/
│   └── podlift/          # CLI entry point
│       ├── main.go
│       └── commands/     # Cobra commands
│           └── root.go
├── internal/             # Internal packages (not exported)
│   ├── config/           # Configuration parsing
│   ├── deploy/           # Deployment logic
│   ├── ssh/              # SSH client
│   ├── docker/           # Docker client
│   ├── git/              # Git integration
│   └── ui/               # TUI components (Charm)
├── tests/                # Test files
│   ├── integration/
│   └── e2e/
├── testdata/             # Test fixtures
├── docs/                 # Documentation
└── Makefile             # Build commands
```

## Phase 0 Tasks

### Week 1-2: Foundation

- [x] Go module initialized
- [x] Project structure created
- [x] Dependencies added (cobra, charm, yaml, docker, ssh)
- [x] Basic CLI skeleton (version, init commands)
- [ ] Configuration parser (internal/config)
- [ ] SSH client wrapper (internal/ssh)
- [ ] Git integration (internal/git)
- [ ] Basic tests

## Next Steps

1. **Implement config parser** (`internal/config/`)
   - Parse podlift.yml
   - Validate configuration
   - Handle defaults

2. **Implement SSH client** (`internal/ssh/`)
   - Connect to servers
   - Execute remote commands
   - Handle errors gracefully

3. **Implement Git integration** (`internal/git/`)
   - Check git state
   - Get commit hash
   - Validate clean working tree

4. **Add TUI components** (`internal/ui/`)
   - Progress bars
   - Spinners
   - Tables
   - Status displays

## Development Workflow

1. Make changes
2. Run `make test` to verify
3. Run `make build` to build
4. Test manually: `./bin/podlift <command>`
5. Commit

## Dependencies

- **cobra**: CLI framework
- **charmbracelet**: TUI components (bubbletea, lipgloss, bubbles)
- **yaml.v3**: YAML parsing
- **docker**: Docker client SDK
- **golang.org/x/crypto/ssh**: SSH client

## Building with Version Info

```bash
make build VERSION=v0.1.0
# or
go build -ldflags "-X 'github.com/ekinertac/podlift/cmd/podlift/commands.Version=v0.1.0'" -o bin/podlift ./cmd/podlift
```

