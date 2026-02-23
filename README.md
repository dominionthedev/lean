<p align="center">
  <img src="assets/logo.svg" alt="lean logo" width="400">
</p>

# lean

## ✨ Overview

lean is a lightweight environment profile manager for managing multiple `.env` configurations safely and efficiently.

## 🚀 Features

- Initialize lean configuration
- Create multiple environment profiles
- Switch between profiles seamlessly
- Track current active profile
- List all available profiles

## 📦 Installation

```bash
go install github.com/dominionthedev/lean@latest
```

## 🛠️ Usage

### Initialize lean

```bash
lean init
```

### Create a profile

```bash
lean create <profile_name>
```

### Apply a profile

```bash
lean apply <profile_name>
```

### List profiles

```bash
lean list
```

### Show current profile

```bash
lean current
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please read the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines.
