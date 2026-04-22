# docsfy

基于 Go 的 Markdown 文档静态站点服务，支持本地预览、静态导出、全文搜索与热更新。

## 使用

### 预览服务

```bash
go run ./cmd/main.go serve --docs ./docs --port 8080
```

### 构建静态站点

```bash
go run ./cmd/main.go build --docs ./docs --out ./dist
```

## Markdown 布局规范

`docsfy` 会按目录结构自动生成左侧菜单，建议按下面方式组织文档：

```text
docs/
  README.md              # 根目录首页（主菜单）
  guide/
    README.md            # guide 目录主菜单（概览）
    getting-started.md   # guide 子菜单
    advanced.md          # guide 子菜单
  risk/
    README.md            # risk 目录主菜单（概览）
    ARCHITECTURE.md      # risk 子菜单
    RULES.md             # risk 子菜单
```

### 菜单生成规则

- 目录下的 `README.md`（或 `index.md`）作为该目录的主菜单（概览）。
- 同目录下其它 `.md` 文件作为该目录子菜单。
- 文档中的 `##` / `###` 标题会作为下级锚点子菜单。

### 文档写法建议

- 每篇文档以 `# 一级标题` 开头，作为页面标题与搜索标题。
- 使用 `##`、`###` 划分章节，便于自动生成锚点导航。
- 代码块语法：

```markdown
```go
fmt.Println("hello")
```
```

### Kroki 图表写法

支持 Kroki 图形（如 `mermaid`、`plantuml`、`graphviz` 等）：

```markdown
```mermaid
flowchart LR
A --> B
```
```

运行时会自动渲染为图形，点击图可在当前页查看大图。

默认情况下，`mermaid` 使用前端本地脚本渲染（`/assets/vendor/mermaid.min.js`），不走 Kroki 服务端。

## 内网/离线环境使用

为便于在内网环境使用，前端核心样式脚本已改为本地静态资源（`/assets/vendor/tailwindcss.cdn.js`），不依赖外网 CDN。

### Kroki 配置

Kroki 图表默认使用 `https://kroki.io`。若内网无法访问公网，请使用以下方式之一：

- 指定内网 Kroki 服务：

```bash
go run ./cmd/main.go serve --docs ./docs --port 8080 --kroki-url http://your-kroki.internal
```

- 或禁用 Kroki 渲染（保留源码代码块显示）：

```bash
go run ./cmd/main.go serve --docs ./docs --port 8080 --kroki-disable
```

同样适用于 `build` 命令。
