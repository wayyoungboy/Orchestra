# Orchestra 静态官网实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Orchestra 创建一个深色科技风格的静态官网，部署到 GitHub Pages

**Architecture:** 纯 HTML + CSS + JavaScript 静态站点，使用 CSS Grid/Flexbox 布局，CSS 变量实现主题，部署到 `docs/` 目录供 GitHub Pages 使用

**Tech Stack:** HTML5, CSS3 (Variables, Grid, Flexbox, Animations), Vanilla JavaScript, GitHub Pages

---

## File Structure

```
docs/
├── index.html          # 主页面
├── styles.css          # 样式文件
├── script.js           # 交互脚本
├── favicon.ico         # 网站图标 (可选)
└── assets/
    └── images/         # 图片资源
```

---

### Task 1: 创建 HTML 主页面结构

**Files:**
- Create: `docs/index.html`

- [ ] **Step 1: 创建 HTML 基础结构和 SEO 元数据**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="Orchestra - 多智能体协作平台，支持 Claude Code、Gemini CLI 等多个 AI Agent 并行运行">
    <meta name="keywords" content="Orchestra, AI, Multi-Agent, Claude Code, Gemini CLI, Collaboration">
    <meta name="author" content="Orchestra Team">
    <meta property="og:title" content="Orchestra - Multi-Agent Collaboration Platform">
    <meta property="og:description" content="多智能体协作平台，支持多个 AI Agent 并行运行">
    <meta property="og:type" content="website">
    <meta property="og:url" content="https://wayyoungboy.github.io/Orchestra/">
    <title>Orchestra - Multi-Agent Collaboration Platform</title>
    <link rel="stylesheet" href="styles.css">
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🎵</text></svg>">
</head>
<body>
```

- [ ] **Step 2: 添加导航栏组件**

```html
    <!-- Navigation -->
    <nav class="navbar">
        <div class="container">
            <a href="#" class="logo">
                <span class="logo-icon">O</span>
                <span class="logo-text">Orchestra</span>
            </a>
            <div class="nav-links">
                <a href="#features">Features</a>
                <a href="#quick-start">Quick Start</a>
                <a href="https://github.com/wayyoungboy/Orchestra" target="_blank" class="nav-github">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                    </svg>
                    GitHub
                </a>
            </div>
        </div>
    </nav>
```

- [ ] **Step 3: 添加 Hero 区域**

```html
    <!-- Hero Section -->
    <section class="hero">
        <div class="hero-bg">
            <div class="hero-gradient"></div>
            <div class="hero-grid"></div>
        </div>
        <div class="container">
            <div class="hero-content">
                <h1 class="hero-title">
                    <span class="title-highlight">Orchestra</span>
                    <br>Multi-Agent Collaboration
                </h1>
                <p class="hero-subtitle">
                    多智能体协作平台，支持 Claude Code、Gemini CLI、Aider 等多个 AI Agent 并行运行与协同工作
                </p>
                <div class="hero-actions">
                    <a href="https://github.com/wayyoungboy/Orchestra" target="_blank" class="btn btn-primary">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                        </svg>
                        View on GitHub
                    </a>
                    <a href="#quick-start" class="btn btn-secondary">Quick Start</a>
                </div>
            </div>
            <div class="hero-visual">
                <div class="terminal-preview">
                    <div class="terminal-header">
                        <span class="terminal-dot red"></span>
                        <span class="terminal-dot yellow"></span>
                        <span class="terminal-dot green"></span>
                        <span class="terminal-title">Orchestra Terminal</span>
                    </div>
                    <div class="terminal-body">
                        <div class="terminal-line">
                            <span class="prompt">$</span>
                            <span class="command">claude</span>
                        </div>
                        <div class="terminal-output">Claude Code ready...</div>
                        <div class="terminal-line">
                            <span class="prompt">$</span>
                            <span class="command">gemini</span>
                        </div>
                        <div class="terminal-output">Gemini CLI ready...</div>
                        <div class="terminal-line">
                            <span class="prompt cursor">$</span>
                            <span class="typing">_</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
```

- [ ] **Step 4: 添加功能特性区域**

```html
    <!-- Features Section -->
    <section id="features" class="features">
        <div class="container">
            <h2 class="section-title">Features</h2>
            <p class="section-subtitle">强大的多智能体协作能力</p>
            <div class="features-grid">
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
                            <line x1="8" y1="21" x2="16" y2="21"></line>
                            <line x1="12" y1="17" x2="12" y2="21"></line>
                        </svg>
                    </div>
                    <h3 class="feature-title">多终端并行</h3>
                    <p class="feature-desc">同时运行多个 AI Agent，独立终端会话，互不干扰</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                        </svg>
                    </div>
                    <h3 class="feature-title">实时协作</h3>
                    <p class="feature-desc">聊天界面支持 @提及，团队成员即时沟通协作</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                            <circle cx="9" cy="7" r="4"></circle>
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                        </svg>
                    </div>
                    <h3 class="feature-title">角色管理</h3>
                    <p class="feature-desc">Owner、Admin、Assistant、Member 四种角色，精细权限控制</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                        </svg>
                    </div>
                    <h3 class="feature-title">工作区管理</h3>
                    <p class="feature-desc">多工作区支持，快速切换，项目隔离管理</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="16 18 22 12 16 6"></polyline>
                            <polyline points="8 6 2 12 8 18"></polyline>
                        </svg>
                    </div>
                    <h3 class="feature-title">多 Agent 支持</h3>
                    <p class="feature-desc">Claude Code、Gemini CLI、Aider、Cursor Agent 等主流工具</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                            <line x1="3" y1="9" x2="21" y2="9"></line>
                            <line x1="9" y1="21" x2="9" y2="9"></line>
                        </svg>
                    </div>
                    <h3 class="feature-title">现代化 UI</h3>
                    <p class="feature-desc">Vue 3 + TypeScript + Tailwind CSS，流畅的交互体验</p>
                </div>
            </div>
        </div>
    </section>
```

- [ ] **Step 5: 添加快速开始区域**

```html
    <!-- Quick Start Section -->
    <section id="quick-start" class="quick-start">
        <div class="container">
            <h2 class="section-title">Quick Start</h2>
            <p class="section-subtitle">几分钟内开始使用 Orchestra</p>
            <div class="steps-container">
                <div class="step">
                    <div class="step-number">1</div>
                    <h3 class="step-title">Clone Repository</h3>
                    <div class="code-block">
                        <code>git clone https://github.com/wayyoungboy/Orchestra.git</code>
                    </div>
                </div>
                <div class="step">
                    <div class="step-number">2</div>
                    <h3 class="step-title">Start Backend</h3>
                    <div class="code-block">
                        <code>cd backend && make build && make run</code>
                    </div>
                </div>
                <div class="step">
                    <div class="step-number">3</div>
                    <h3 class="step-title">Start Frontend</h3>
                    <div class="code-block">
                        <code>cd frontend && npm install && npm run dev</code>
                    </div>
                </div>
            </div>
            <div class="requirements">
                <h4>Requirements</h4>
                <ul>
                    <li>Go 1.21+</li>
                    <li>Node.js 18+</li>
                    <li>SQLite3</li>
                </ul>
            </div>
        </div>
    </section>
```

- [ ] **Step 6: 添加页脚**

```html
    <!-- Footer -->
    <footer class="footer">
        <div class="container">
            <div class="footer-content">
                <div class="footer-brand">
                    <span class="logo-icon">O</span>
                    <span class="logo-text">Orchestra</span>
                </div>
                <p class="footer-desc">Multi-Agent Collaboration Platform</p>
                <div class="footer-links">
                    <a href="https://github.com/wayyoungboy/Orchestra" target="_blank">GitHub</a>
                    <a href="https://github.com/wayyoungboy/Orchestra/issues" target="_blank">Issues</a>
                    <a href="https://github.com/wayyoungboy/Orchestra/blob/main/LICENSE" target="_blank">License</a>
                </div>
                <p class="footer-copyright">&copy; 2024 Orchestra. MIT License.</p>
            </div>
        </div>
    </footer>
    <script src="script.js"></script>
</body>
</html>
```

---

### Task 2: 创建 CSS 样式文件

**Files:**
- Create: `docs/styles.css`

- [ ] **Step 1: 定义 CSS 变量和重置样式**

```css
/* CSS Variables */
:root {
    --color-bg: #0f0f14;
    --color-bg-secondary: #16161d;
    --color-bg-tertiary: #1e1e28;
    --color-surface: #1a1a24;
    --color-border: rgba(255, 255, 255, 0.08);
    --color-primary: #6366f1;
    --color-primary-light: #818cf8;
    --color-primary-dark: #4f46e5;
    --color-text: #ffffff;
    --color-text-secondary: rgba(255, 255, 255, 0.7);
    --color-text-muted: rgba(255, 255, 255, 0.4);
    --color-success: #10b981;
    --color-warning: #f59e0b;
    --color-error: #ef4444;
    --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    --font-mono: 'SF Mono', 'Fira Code', 'Consolas', monospace;
    --transition-fast: 150ms ease;
    --transition-base: 250ms ease;
    --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
    --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.4);
    --shadow-lg: 0 10px 25px rgba(0, 0, 0, 0.5);
    --shadow-glow: 0 0 40px rgba(99, 102, 241, 0.3);
}

/* Reset */
*, *::before, *::after {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

html {
    scroll-behavior: smooth;
}

body {
    font-family: var(--font-sans);
    background-color: var(--color-bg);
    color: var(--color-text);
    line-height: 1.6;
    overflow-x: hidden;
}

a {
    color: inherit;
    text-decoration: none;
}

ul {
    list-style: none;
}

/* Container */
.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
}
```

- [ ] **Step 2: 添加导航栏样式**

```css
/* Navbar */
.navbar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
    padding: 16px 0;
    background: rgba(15, 15, 20, 0.8);
    backdrop-filter: blur(12px);
    border-bottom: 1px solid var(--color-border);
}

.navbar .container {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.logo {
    display: flex;
    align-items: center;
    gap: 10px;
    font-weight: 700;
    font-size: 20px;
}

.logo-icon {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, var(--color-primary), var(--color-primary-light));
    border-radius: 10px;
    font-weight: 800;
    font-size: 18px;
}

.nav-links {
    display: flex;
    align-items: center;
    gap: 32px;
}

.nav-links a {
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-secondary);
    transition: color var(--transition-fast);
}

.nav-links a:hover {
    color: var(--color-text);
}

.nav-github {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: var(--color-surface);
    border-radius: 8px;
    border: 1px solid var(--color-border);
}

.nav-github:hover {
    background: var(--color-bg-tertiary);
    border-color: var(--color-primary);
}
```

- [ ] **Step 3: 添加 Hero 区域样式**

```css
/* Hero Section */
.hero {
    position: relative;
    min-height: 100vh;
    display: flex;
    align-items: center;
    padding: 120px 0 80px;
    overflow: hidden;
}

.hero-bg {
    position: absolute;
    inset: 0;
    z-index: -1;
}

.hero-gradient {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse at 50% 0%, rgba(99, 102, 241, 0.15) 0%, transparent 60%),
                radial-gradient(ellipse at 100% 50%, rgba(139, 92, 246, 0.1) 0%, transparent 50%),
                radial-gradient(ellipse at 0% 80%, rgba(59, 130, 246, 0.1) 0%, transparent 50%);
}

.hero-grid {
    position: absolute;
    inset: 0;
    background-image:
        linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px),
        linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
    background-size: 60px 60px;
    mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
}

.hero .container {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 60px;
    align-items: center;
}

.hero-content {
    max-width: 560px;
}

.hero-title {
    font-size: 56px;
    font-weight: 800;
    line-height: 1.1;
    letter-spacing: -0.02em;
    margin-bottom: 24px;
}

.title-highlight {
    background: linear-gradient(135deg, var(--color-primary-light), #a78bfa, #60a5fa);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
}

.hero-subtitle {
    font-size: 18px;
    color: var(--color-text-secondary);
    margin-bottom: 32px;
    line-height: 1.7;
}

.hero-actions {
    display: flex;
    gap: 16px;
}

/* Buttons */
.btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 14px 28px;
    font-size: 15px;
    font-weight: 600;
    border-radius: 12px;
    cursor: pointer;
    transition: all var(--transition-base);
    border: none;
}

.btn-primary {
    background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
    color: white;
    box-shadow: var(--shadow-glow);
}

.btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 0 60px rgba(99, 102, 241, 0.4);
}

.btn-secondary {
    background: var(--color-surface);
    color: var(--color-text);
    border: 1px solid var(--color-border);
}

.btn-secondary:hover {
    background: var(--color-bg-tertiary);
    border-color: var(--color-primary);
}
```

- [ ] **Step 4: 添加终端预览样式**

```css
/* Terminal Preview */
.hero-visual {
    display: flex;
    justify-content: center;
}

.terminal-preview {
    width: 100%;
    max-width: 480px;
    background: var(--color-bg-secondary);
    border-radius: 16px;
    border: 1px solid var(--color-border);
    box-shadow: var(--shadow-lg);
    overflow: hidden;
}

.terminal-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: var(--color-bg-tertiary);
    border-bottom: 1px solid var(--color-border);
}

.terminal-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
}

.terminal-dot.red { background: #ff5f56; }
.terminal-dot.yellow { background: #ffbd2e; }
.terminal-dot.green { background: #27c93f; }

.terminal-title {
    margin-left: auto;
    font-size: 13px;
    color: var(--color-text-muted);
}

.terminal-body {
    padding: 20px;
    font-family: var(--font-mono);
    font-size: 14px;
    line-height: 1.8;
}

.terminal-line {
    display: flex;
    gap: 8px;
}

.prompt {
    color: var(--color-primary-light);
}

.command {
    color: var(--color-success);
}

.terminal-output {
    color: var(--color-text-muted);
    padding-left: 24px;
}

.cursor .typing {
    animation: blink 1s step-end infinite;
}

@keyframes blink {
    50% { opacity: 0; }
}
```

- [ ] **Step 5: 添加功能特性区域样式**

```css
/* Features Section */
.features {
    padding: 100px 0;
    background: var(--color-bg-secondary);
}

.section-title {
    font-size: 40px;
    font-weight: 700;
    text-align: center;
    margin-bottom: 12px;
}

.section-subtitle {
    font-size: 18px;
    color: var(--color-text-secondary);
    text-align: center;
    margin-bottom: 60px;
}

.features-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 24px;
}

.feature-card {
    padding: 32px;
    background: var(--color-bg);
    border-radius: 16px;
    border: 1px solid var(--color-border);
    transition: all var(--transition-base);
}

.feature-card:hover {
    border-color: var(--color-primary);
    transform: translateY(-4px);
    box-shadow: var(--shadow-md);
}

.feature-icon {
    width: 56px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(139, 92, 246, 0.2));
    border-radius: 14px;
    margin-bottom: 20px;
    color: var(--color-primary-light);
}

.feature-title {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 12px;
}

.feature-desc {
    font-size: 14px;
    color: var(--color-text-secondary);
    line-height: 1.7;
}
```

- [ ] **Step 6: 添加快速开始和页脚样式**

```css
/* Quick Start Section */
.quick-start {
    padding: 100px 0;
}

.steps-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
    max-width: 700px;
    margin: 0 auto 40px;
}

.step {
    display: flex;
    align-items: center;
    gap: 24px;
    padding: 24px;
    background: var(--color-bg-secondary);
    border-radius: 16px;
    border: 1px solid var(--color-border);
}

.step-number {
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-primary);
    border-radius: 12px;
    font-size: 20px;
    font-weight: 700;
    flex-shrink: 0;
}

.step-title {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 8px;
}

.code-block {
    background: var(--color-bg);
    padding: 12px 16px;
    border-radius: 8px;
    font-family: var(--font-mono);
    font-size: 14px;
    color: var(--color-primary-light);
    overflow-x: auto;
}

.requirements {
    text-align: center;
}

.requirements h4 {
    font-size: 14px;
    color: var(--color-text-muted);
    margin-bottom: 12px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
}

.requirements ul {
    display: flex;
    justify-content: center;
    gap: 32px;
}

.requirements li {
    font-size: 14px;
    color: var(--color-text-secondary);
}

/* Footer */
.footer {
    padding: 60px 0;
    background: var(--color-bg-secondary);
    border-top: 1px solid var(--color-border);
}

.footer-content {
    text-align: center;
}

.footer-brand {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    font-weight: 700;
    font-size: 20px;
    margin-bottom: 12px;
}

.footer-desc {
    color: var(--color-text-secondary);
    margin-bottom: 24px;
}

.footer-links {
    display: flex;
    justify-content: center;
    gap: 32px;
    margin-bottom: 24px;
}

.footer-links a {
    font-size: 14px;
    color: var(--color-text-secondary);
    transition: color var(--transition-fast);
}

.footer-links a:hover {
    color: var(--color-primary-light);
}

.footer-copyright {
    font-size: 13px;
    color: var(--color-text-muted);
}

/* Responsive */
@media (max-width: 968px) {
    .hero .container {
        grid-template-columns: 1fr;
        text-align: center;
    }

    .hero-content {
        max-width: 100%;
    }

    .hero-actions {
        justify-content: center;
    }

    .hero-visual {
        order: -1;
    }

    .features-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 640px) {
    .hero-title {
        font-size: 36px;
    }

    .section-title {
        font-size: 28px;
    }

    .features-grid {
        grid-template-columns: 1fr;
    }

    .nav-links a:not(.nav-github) {
        display: none;
    }

    .step {
        flex-direction: column;
        text-align: center;
    }

    .requirements ul {
        flex-direction: column;
        gap: 8px;
    }
}
```

---

### Task 3: 创建 JavaScript 交互脚本

**Files:**
- Create: `docs/script.js`

- [ ] **Step 1: 创建基础交互脚本**

```javascript
// Orchestra Landing Page Scripts

document.addEventListener('DOMContentLoaded', () => {
    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Navbar background on scroll
    const navbar = document.querySelector('.navbar');
    window.addEventListener('scroll', () => {
        if (window.scrollY > 50) {
            navbar.style.background = 'rgba(15, 15, 20, 0.95)';
        } else {
            navbar.style.background = 'rgba(15, 15, 20, 0.8)';
        }
    });

    // Animate elements on scroll
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    document.querySelectorAll('.feature-card, .step').forEach(el => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(20px)';
        el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(el);
    });

    // Terminal typing animation
    const typingElement = document.querySelector('.typing');
    if (typingElement) {
        const commands = ['claude --help', 'gemini chat', 'aider .', 'git status'];
        let commandIndex = 0;
        let charIndex = 0;
        let isDeleting = false;

        function typeCommand() {
            const currentCommand = commands[commandIndex];

            if (isDeleting) {
                typingElement.textContent = currentCommand.substring(0, charIndex - 1);
                charIndex--;
            } else {
                typingElement.textContent = currentCommand.substring(0, charIndex + 1);
                charIndex++;
            }

            let typeSpeed = isDeleting ? 50 : 100;

            if (!isDeleting && charIndex === currentCommand.length) {
                typeSpeed = 2000;
                isDeleting = true;
            } else if (isDeleting && charIndex === 0) {
                isDeleting = false;
                commandIndex = (commandIndex + 1) % commands.length;
                typeSpeed = 500;
            }

            setTimeout(typeCommand, typeSpeed);
        }

        setTimeout(typeCommand, 1000);
    }
});
```

---

### Task 4: 配置 GitHub Pages 部署

**Files:**
- Create: `.github/workflows/deploy.yml`
- Modify: `docs/CNAME` (optional)

- [ ] **Step 1: 创建 GitHub Actions 部署工作流**

```yaml
name: Deploy to GitHub Pages

on:
  push:
    branches: [ main ]
    paths:
      - 'docs/**'
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: "pages"
  cancel-in-progress: true

jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Pages
        uses: actions/configure-pages@v4

      - name: Upload artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: 'docs'

      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 2: 提交并推送到 GitHub**

```bash
git add docs/ .github/workflows/deploy.yml
git commit -m "feat: add static landing page for GitHub Pages"
git push origin main
```

- [ ] **Step 3: 启用 GitHub Pages**

在 GitHub 仓库设置中：
1. 进入 Settings > Pages
2. Source 选择 "GitHub Actions"
3. 等待部署完成

访问地址: `https://wayyoungboy.github.io/Orchestra/`

---

## Self-Review Checklist

**1. Spec Coverage:**
- ✅ Hero 区域 - 包含标题、描述、CTA 按钮
- ✅ 功能特性 - 6个核心功能卡片
- ✅ 快速开始 - 3步安装指南
- ✅ GitHub 链接 - 导航栏和 Hero 区域
- ✅ 响应式设计 - 移动端适配
- ✅ 深色科技风格 - CSS 变量实现

**2. Placeholder Scan:**
- ✅ 所有代码完整，无 TBD 或 TODO
- ✅ 无占位符图片
- ✅ 无模糊描述

**3. Type Consistency:**
- ✅ CSS 变量命名一致
- ✅ 类名命名规范统一

---

Plan complete and saved to `docs/superpowers/plans/2026-03-30-landing-page.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**