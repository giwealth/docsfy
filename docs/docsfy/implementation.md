# Markdown 静态站点服务实现文档（Golang）

## 1. 文档目的

本文档基于 `requirements.md`，给出可直接落地的技术实现方案，用于指导开发一个类似 docsify 的 Markdown 文档静态站点服务，支持：

- Markdown 渲染为 HTML
- 本地 HTTP 访问（`serve`）
- 静态站点导出（`build`）
- 全文搜索（索引 JSON + 前端检索）
- 增量构建与文件监听热更新

## 2. 总体架构

系统分为 6 个核心层次：

1. **配置层（config）**：读取默认配置、文件配置、命令行参数。
2. **内容层（content）**：扫描文档目录、构建文档树和路由信息。
3. **渲染层（render）**：Markdown -> HTML，并输出标题锚点、文本摘要。
4. **站点层（site）**：组装模板数据、导航结构、搜索索引。
5. **构建层（build）**：静态导出到 `dist/`，复制资源，输出搜索索引。
6. **服务层（server）**：HTTP 路由、静态资源托管、监听变更、热更新通知。

## 3. 目录与模块设计

建议目录结构如下：

```text
cmd/main.go
internal/config
internal/content
internal/render
internal/site
internal/build
internal/server
internal/search
web/templates
web/assets
docs
dist
```

### 3.1 `cmd/main.go`

- CLI 入口
- 子命令：
  - `serve --docs ./docs --port 8080`
  - `build --docs ./docs --out ./dist`

### 3.2 `internal/config`

- 结构体 `Config`：
  - `DocsDir string`
  - `OutDir string`
  - `Port int`
  - `SiteTitle string`
  - `AssetsDir string`
  - `TemplatesDir string`
- 配置优先级：CLI 参数 > `docsfy.yaml` > 默认值

### 3.3 `internal/content`

- 扫描 `DocsDir` 下所有 `.md`
- 产出 `Document` 列表与目录树：
  - `SourcePath`
  - `RoutePath`（如 `/guide/getting-started/`）
  - `Title`
  - `RawMarkdown`
  - `LastModified`
- 首页规则：`README.md` 或 `index.md` 映射 `/`

### 3.4 `internal/render`

- 采用 `goldmark` 渲染 Markdown
- 开启常用扩展（GFM、表格、自动链接、删除线等）
- 输出：
  - `HTML string`
  - `PlainText string`（用于搜索索引）
  - `TOC []Heading`（可用于后续目录增强）

### 3.5 `internal/site`

- 组装页面数据：
  - 当前页内容
  - 导航树
  - 站点元信息
- 通过 `html/template` 渲染 `web/templates/*.tmpl`

### 3.6 `internal/search`

- 生成 `search-index.json`
- 索引项 `SearchItem`：
  - `title`
  - `route`
  - `content`（纯文本摘要或截断内容）
- 构建时全量生成；`serve` 下支持局部更新（按变更文档更新索引项）

### 3.7 `internal/build`

- 全量构建：
  1. 清理或创建 `OutDir`
  2. 为每个路由写入 `index.html`
  3. 复制 `web/assets` 到 `dist/assets`
  4. 写入 `dist/search-index.json`

### 3.8 `internal/server`

- 提供 HTTP 服务：
  - 页面路由（优先命中已渲染内容）
  - 资源路由（`/assets/*`）
  - 404 页面
- `serve` 下启动 watcher：
  - 监听 `docs/`、`web/templates/`、`web/assets/`
  - 增量更新内存缓存 + 通知前端刷新

## 4. 关键数据结构（建议）

```go
type Document struct {
    SourcePath   string
    RoutePath    string
    Title        string
    RawMarkdown  string
    HTML         string
    PlainText    string
    LastModified time.Time
}

type NavNode struct {
    Name     string
    Route    string
    Children []*NavNode
    IsDoc    bool
}

type SearchItem struct {
    Title   string `json:"title"`
    Route   string `json:"route"`
    Content string `json:"content"`
}
```

## 5. 路由与文件映射规则

- `docs/a/b.md` -> `/a/b/` -> `dist/a/b/index.html`
- `docs/README.md` 或 `docs/index.md` -> `/` -> `dist/index.html`
- 目录级 `README.md` 可映射为目录根路径（可选扩展）：
  - `docs/guide/README.md` -> `/guide/`

## 6. 增量构建与热更新方案

## 6.1 监听策略

- 使用文件监听库（如 `fsnotify`）监听以下目录：
  - `docs/`
  - `web/templates/`
  - `web/assets/`

## 6.2 变更处理策略

- **Markdown 变更**：
  - 重新解析该文件
  - 更新文档缓存与导航（若路径变化）
  - 仅重渲染受影响路由
  - 更新对应搜索索引项
- **模板变更**：
  - 重新加载模板
  - 重渲染全部页面（模板影响全局）
- **静态资源变更**：
  - 直接重新提供新资源内容
  - 向浏览器发送刷新通知

## 6.3 前端热更新机制（简化版）

- 页面注入一个轻量脚本，连接 `/__livereload`（SSE）。
- 服务端在构建更新后向 SSE 客户端推送 `reload` 事件。
- 前端收到事件后执行 `location.reload()`。

## 7. 全文搜索实现方案

## 7.1 构建阶段

- 从每个文档提取：
  - 第一标题作为 `title`（若无则使用文件名）
  - 路由 `route`
  - 正文纯文本 `content`（可截断到固定长度）
- 写入 `dist/search-index.json`

## 7.2 前端检索阶段

- 页面加载后按需请求 `/search-index.json`。
- 前端执行简单匹配（`includes`/分词匹配，先做 MVP）。
- 展示匹配列表（标题 + 摘要 + 路由链接）。

## 8. 命令行为定义

### 8.1 `build`

- 输入：`--docs`、`--out`、`--config`
- 输出：`dist/` 完整静态站点
- 退出码：成功 0，失败非 0

### 8.2 `serve`

- 输入：`--docs`、`--port`、`--config`
- 行为：
  - 启动 HTTP 服务
  - 初始构建内存页面缓存
  - 启动监听和热更新

## 9. 错误处理与日志

- 统一日志接口（后续可替换 zap/logrus，MVP 使用标准库 `log`）
- 常见错误处理：
  - Markdown 解析失败：记录错误并跳过该文档
  - 模板加载失败：`serve` 启动失败并返回错误
  - 路由冲突：启动时输出冲突清单并终止

## 10. 测试计划

最小测试集合：

- `content`：
  - 路由映射正确性
  - 首页映射优先级
- `render`：
  - Markdown 到 HTML 关键语法输出
  - 标题提取与纯文本提取
- `search`：
  - 索引 JSON 结构正确
  - 单文档更新可正确替换索引项
- `build`：
  - 输出目录结构与文件存在性

## 11. 分阶段实施计划

### 阶段 1（MVP 可访问）

- 完成扫描、渲染、模板、`serve`、`build`
- 支持基础导航和静态资源

### 阶段 2（增强能力）

- 完成搜索索引与前端搜索 UI
- 完成 watcher + 增量构建 + SSE 热更新

### 阶段 3（工程化）

- 完善测试覆盖、错误提示和文档
- 增加性能优化（缓存、并发渲染、去抖动）

## 12. 交付清单

- 可执行命令：`docsfy serve`、`docsfy build`
- 默认模板与样式
- `search-index.json` 生成能力
- 增量构建与热更新能力
- 使用说明文档（README）
