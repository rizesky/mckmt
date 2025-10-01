# MCKMT Tools

This directory contains utility tools for development and maintenance of the MCKMT project.

## Available Tools

### hash-password.go

Generates password hashes using the same Argon2id algorithm used by the  authentication system.

**Usage:**
```bash
go run tools/hash-password.go <password>
```

**Example:**
```bash
go run tools/hash-password.go admin123
```

**Output:**
```
Password: admin123
Hash: $argon2id$v=19$m=65536,t=4,p=2$eHtOigMBSF8OqtSnq43o/A$SRP0Y1+V+a6npQdERQirv5mJoAIgOEeZ+BMJRq9aEKc
✓ Hash verification successful
```

This tool is useful for:
- Generating password hashes for database seeds
- Testing password verification
- Creating test users with known passwords

## Adding New Tools

When adding new tools to this directory:

1. Use the same naming convention: `kebab-case.go`
2. Include proper error handling and usage instructions
3. Update this README with documentation
4. Use the project's existing packages when possible (e.g., `internal/auth` for password operations)
5. Make tools self-contained and easy to run with `go run tools/tool-name.go`
