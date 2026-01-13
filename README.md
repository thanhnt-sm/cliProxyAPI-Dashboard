# 🚀 CLIProxy Manager Dashboard

<div align="center">

![CLIProxy](https://img.shields.io/badge/CLIProxy-v6.0-0075FF?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)

[![GitHub Sponsor](https://img.shields.io/badge/Sponsor-❤️-ea4aaa?style=for-the-badge&logo=github-sponsors)](https://github.com/sponsors/0xAstroAlpha)
[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-☕-FFDD00?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://buymeacoffee.com/astroalpha)

[![GitHub Stars](https://img.shields.io/github/stars/0xAstroAlpha/cliProxyAPI-Dashboard?style=flat-square&logo=github)](https://github.com/0xAstroAlpha/cliProxyAPI-Dashboard/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/0xAstroAlpha/cliProxyAPI-Dashboard?style=flat-square&logo=github)](https://github.com/0xAstroAlpha/cliProxyAPI-Dashboard/network/members)
[![GitHub Issues](https://img.shields.io/github/issues/0xAstroAlpha/cliProxyAPI-Dashboard?style=flat-square&logo=github)](https://github.com/0xAstroAlpha/cliProxyAPI-Dashboard/issues)
[![Contributors](https://img.shields.io/github/contributors/0xAstroAlpha/cliProxyAPI-Dashboard?style=flat-square&logo=github)](https://github.com/0xAstroAlpha/cliProxyAPI-Dashboard/graphs/contributors)

**A modern, beautiful dashboard for managing your CLIProxy instances**

[Dashboard Features](#-dashboard-features) • [CLIProxy Features](#-cliproxy-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Support](#-support--donations)

</div>


![CLIProxy Dashboard Preview](docs/Homepage.jpg)


---

## 📢 Recent Updates (v6.0.1)

- **Fixed Missing Custom UI**: The custom management dashboard is now correctly baked into the Docker image, ensuring it appears for all users without manual volume mounting.
- **Fixed Usage Persistence**: Resolved an issue where `usage.db` was initializing in an ephemeral directory. Local history now persists correctly across container restarts.
- **Restored Legacy Data**: Automatically migrated legacy local data (from native runs) to the Docker volume.

---

## 📖 Documentation

| Document | Description |
|----------|-------------|
| **[Dashboard Guide](docs/DASHBOARD.md)** | Hướng dẫn sử dụng Dashboard (Tiếng Việt) |
| **[SDK Usage](docs/sdk-usage.md)** | How to use the Go SDK |
| **[SDK Advanced](docs/sdk-advanced.md)** | Advanced SDK features |
| **[SDK Access](docs/sdk-access.md)** | Access control documentation |
| **[SDK Watcher](docs/sdk-watcher.md)** | File watcher documentation |

---

## ✨ Dashboard Features

The CLIProxy Manager Dashboard provides a **premium Vision UI** experience for monitoring and managing your proxy server.

### 🎯 Overview Panel
- **Real-time Server Status** - Live connection monitoring with animated indicators
- **Usage Statistics** - Total requests, tokens, success/failure rates at a glance
- **Saved Cost Display** - Track how much you've saved with dynamic emoji indicators (🪙💸💵💰💎🚀)
- **Sparkline Charts** - Mini trend visualizations for quick insights

### 🏆 Model Leaderboard
- **Top 10 Models Ranking** - See your most-used models with medal icons (🥇🥈🥉)
- **Request & Token Badges** - Beautiful stat badges for easy comparison
- **Real-time Updates** - Data refreshes automatically every 5 seconds

### 📊 Activity Monitor
- **Usage Trends Chart** - Gradient area chart with smooth animations
- **Activity History Table** - Zebra-striped rows with status pills
- **Advanced Filtering** - Filter by model, status, and time range
- **Request Details Modal** - View full request/response data

### 🔐 Account Health Grid
- **Multi-Provider Support** - Gemini, Claude, OpenAI, Qwen, iFlow, Vertex
- **OAuth Authentication** - One-click login for supported providers
- **Status Badges** - Active, refreshing, error states with visual indicators
- **Hover Effects** - Cards scale and glow on interaction

### 💬 AI Playground
- **Multi-Model Chat** - Test any model directly in the dashboard
- **System Prompts** - Customize assistant behavior
- **Parameter Controls** - Temperature, Top P, Max Tokens sliders
- **Thinking Process** - View reasoning (for supported models)
- **Image Attachments** - Upload images for vision models

### 🎨 UI/UX Polish
- **Welcome Message** - Dynamic greeting based on time of day (☀️🌤️🌙)
- **Footer Stats Bar** - Uptime counter, last sync time, version info
- **Quick Actions FAB** - Floating button for common actions
- **Glassmorphism Design** - Modern frosted glass effects
- **Responsive Layout** - Works on desktop and mobile

---

## 🔧 CLIProxy Features

This dashboard is built for [**CLIProxyAPI**](https://github.com/router-for-me/CLIProxyAPI) - a powerful proxy server that provides **OpenAI/Gemini/Claude/Codex compatible API interfaces** for CLI tools and coding assistants.

> 📚 **Original Project:** [github.com/router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)
> 
> 📖 **Documentation:** [help.router-for.me](https://help.router-for.me/)

### Core Features

- OpenAI/Gemini/Claude compatible API endpoints for CLI models
- OpenAI Codex support (GPT models) via OAuth login
- Claude Code support via OAuth login
- Qwen Code support via OAuth login
- iFlow support via OAuth login
- Amp CLI and IDE extensions support with provider routing
- Streaming and non-streaming responses
- Function calling/tools support
- Multimodal input support (text and images)
- Multiple accounts with round-robin load balancing
- Simple CLI authentication flows
- Generative Language API Key support
- OpenAI-compatible upstream providers via config (e.g., OpenRouter)
- Reusable Go SDK for embedding the proxy

### Supported Providers

| Provider | Features |
|----------|----------|
| **Google Gemini** | AI Studio & Gemini CLI multi-account |
| **Anthropic Claude** | Claude Code OAuth + load balancing |
| **OpenAI Codex** | GPT models via OAuth |
| **Alibaba Qwen** | Qwen Code support |
| **iFlow** | iFlow integration |
| **Vertex AI** | Service account authentication |

### Amp CLI Support

CLIProxyAPI includes integrated support for [Amp CLI](https://ampcode.com) and Amp IDE extensions:

- Provider route aliases for Amp's API patterns
- Management proxy for OAuth authentication
- Smart model fallback with automatic routing
- Model mapping to route unavailable models to alternatives

**→ [Complete Amp CLI Integration Guide](https://help.router-for.me/agent-client/amp-cli.html)**

---

## 🚀 Quick Start

### Prerequisites
- **Docker** (recommended) or **Go 1.21+**
- A terminal/command line

---

### Step 1: Clone the Repository

```bash
git clone https://github.com/0xAstroAlpha/cliProxyAPI-Dashboard.git
cd cliProxyAPI-Dashboard
```

---

### Step 2: Create Configuration File

```bash
cp config.example.yaml config.yaml
```

---

### Step 3: Review Key Configuration (Optional)

The default `config.yaml` is ready to use with these settings:

| Setting | Default Value | Description |
|---------|---------------|-------------|
| `remote-management.secret-key` | `setup-secret-key` | Password to access the dashboard |
| `remote-management.allow-remote` | `true` | Allow access from any IP |
| `api-keys[0]` | `sk-antigravity-client-key` | API key for making AI requests |
| `logging-to-file` | `true` | Enable logs tab in dashboard |
| `usage-statistics-enabled` | `true` | Enable activity tracking |

> [!TIP]
> For production, change `secret-key` to a strong password!

---

### Step 4: Build and Run with Docker

```bash
docker-compose up -d --build
```

Wait for the build to complete (first time may take 2-3 minutes).

---

### Step 5: Access the Dashboard

Open your browser and go to:

**🌐 [http://localhost:8317/management.html](http://localhost:8317/management.html)**

When prompted, enter the secret key: `setup-secret-key`

---

### Step 6: Test the API

Use curl to verify the API is working:

```bash
curl -X POST http://localhost:8317/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-antigravity-client-key" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

> [!NOTE]
> The API acts as a proxy. You need to configure AI provider credentials (Gemini, Claude, etc.) in the dashboard's **Config** tab or via `config.yaml` for actual AI responses.

---

### Alternative: Run with Go

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

---

## 🔧 Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| **Dashboard shows 404** | `secret-key` is empty | Set `secret-key` in `config.yaml` |
| **Popup keeps asking for key** | `allow-remote: false` | Set `allow-remote: true` in `config.yaml` |
| **Logs tab shows 400 error** | `logging-to-file: false` | Set `logging-to-file: true` |
| **Activity tab is empty** | `usage-statistics-enabled: false` | Set `usage-statistics-enabled: true` |
| **Playground returns 401** | Wrong API key | Use `sk-antigravity-client-key` or add your key to `api-keys` |
| **Dashboard looks different** | Auto-update overwrote files | Ensure `MANAGEMENT_AUTO_UPDATE: "false"` in `docker-compose.yml` |
| **Changes not applying** | Old Docker image | Run `docker-compose up -d --build` |

---

## 📡 API Endpoints

Once running, the proxy provides these OpenAI-compatible endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/chat/completions` | POST | Chat completions (GPT/Claude/Gemini) |
| `/v1/models` | GET | List available models |
| `/v1/completions` | POST | Text completions |
| `/management.html` | GET | Dashboard UI |

**Base URL:** `http://localhost:8317`

**Authentication:** `Authorization: Bearer <your-api-key>`

---

## 💖 Support & Donations

If you find this project useful, consider supporting the development!

### ☕ Buy Me a Coffee

| Method | Address/Link |
|--------|--------------|
| 🇻🇳 **Vietnam (QR)** | Vietcombank QR available in Dashboard |
| 💳 **PayPal** | `wikigamingmovies@gmail.com` |
| 💚 **USDT (TRC20)** | `TNGsaurWeFhaPPs1yxJ3AY15j1tDecX7ya` |
| 💛 **USDT (BEP20)** | `0x463695638788279F234386a77E0afA2Ee87b57F5` |
| 💜 **Solana (SOL)** | `HkgpzujF8uTBuYEYGSFMnmGzBYmEFyajzTiZacRtXzTr` |

---

## 👨‍💻 Credits

| Role | Credit |
|------|--------|
| **Dashboard UI/UX** | [Brian Le](https://www.facebook.com/lehuyducanh/) |
| **CLIProxyAPI** | [router-for-me](https://github.com/router-for-me/CLIProxyAPI) |

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**⭐ Star the original project: [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)**

Made with ❤️ by the CLIProxy community

</div>
