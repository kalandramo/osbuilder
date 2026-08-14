# 简化版向原版实现对齐的计划

> 目标：在 `examples/osbuilder-simplified` 上，逐步把当前「教学简化版」对齐到
> `internal/osbuilder/cmd` 的真实实现，使简化版既保留可读性，又覆盖真实代码的关键结构。
>
> 本文是**计划文档**，每项标注 `状态`：✅ 已对齐 / 🔲 待对齐 / ⚠️ 已对齐但有差异。
> 配套说明见 [`cli-config-management.md`](./cli-config-management.md)。

---

## 0. 当前简版已有的结构

| 模块 | 文件 | 对齐状态 |
|------|------|---------|
| 入口 | `main.go` | ✅ |
| 根命令 + 分组 + 配置 + AddCommand | `internal/cmd/cmd.go` | ⚠️ 已对齐核心，缺 profiling/插件/headers |
| 全局配置结构体 | `internal/util/options/options.go` | ⚠️ 字段被裁剪（见 §2） |
| 帮助子命令 `osbuilder options` | `internal/cmd/options/options.go` | ✅ |
| Factory 注入 | `internal/cmd/util/factory.go` | ✅ |
| OnInitialize（配置文件+环境变量） | `internal/cmd/util/config.go` | ✅ 已修复 `--config` 延迟求值（原版有 bug） |
| `version` 子命令 | `internal/cmd/version/version.go` | ⚠️ 仅打印 client version，缺 `--output/--short` 与版本变量 |
| `color` 子命令 | `internal/cmd/color/color.go` | ✅ 已高保真对齐 |
| `create project` 子命令 | `internal/cmd/create/project/project.go` | ⚠️ 真实版是 `create cmd`（脚手架生成器），语义不同 |
| `es` 领域配置范例 | `internal/cmd/es/{es,get,options}` | ✅ 简版新增，原版无（教学用） |
| README | `README.md` | ✅ |

---

## 1. 对齐路线图（按优先级分阶段）

### 阶段 A：配置结构体补全（低风险，纯数据）
**目标**：让简版 `Options` 与真实版字段一致，便于新人看到完整配置面。

- [ ] 🔲 `UserOptions` 补全字段：`token`、`password`、`secret-id`、`secret-key`、`client-certificate`、`client-key`
  （当前简版只保留了 `username`）。
- [ ] 🔲 `ServerOptions`（`usercenter`/`gateway`）补全字段：`insecure-skip-tls-verify`、`certificate-authority`、
  `address`（注意真实版是 `address` 而非 `addr`）、`timeout`、`max-retries`、`retry-interval`。
- [ ] 🔲 校验规则对齐：`UserOptions.Validate` 的「username/password 必须同时设置」「secret-id/secret-key 必须同时设置」。
- [ ] 🔲 `Options.Validate` 聚合各子结构错误（当前简版未实现 Validate）。

**产出**：配置 flag 面与原版一致，文档 §3 的字段表可补全。

---

### 阶段 B：根命令钩子与横切能力（中风险）
**目标**：对齐 `PersistentPreRunE` / `PersistentPostRunE` 中的横切逻辑。

- [ ] 🔲 **Profiling 标志**：`--profile` / `--profile-output`（新增 `internal/cmd/profiling.go`，
  对应原版 `cmd/profiling.go`），接入 `PersistentPreRunE: initProfiling` / `PersistentPostRunE: flushProfiling`。
- [ ] 🔲 **warnings-as-errors**：`--warnings-as-errors` flag + `PersistentPostRunE` 中按 warning 计数返回错误。
- [ ] 🔲 **Command Headers RoundTripper**（`addCmdHeaderHooks`）：受 `OSCTL_COMMAND_HEADERS` 环境变量开关，
  把命令路径作为 `X-Headers` 注入 REST 调用。教学版无真实 REST，可保留 hook 骨架但不接 RoundTripper。
- [ ] 🔲 **Warning Handler**：`rest.NewWarningWriter` + `rest.SetDefaultWarningHandler`（需引入 `client-go/rest`）。
- [ ] ⚠️ `PersistentPreRunE` 中 `opts.Complete()` 已对齐；补 `warningHandler` 设置。

---

### 阶段 C：子命令面扩充（中风险，按需）
**目标**：补齐原版中有代表性的子命令，让命令树贴近真实。

- [ ] 🔲 `create cmd`：把简版 `create project` 改为真实语义的「脚手架生成 cobra 子命令」（对齐 `cmd/cmd/cmd.go`），
  引入 `internal/osbuilder/file` + `internal/osbuilder/helper` 的轻量替代（模板渲染、驼峰转换）。
- [ ] 🔲 `completion` 子命令（cobra 自带补全）。
- [ ] 🔲 `alpha` 分组 + 过滤逻辑（`NewCmdAlpha` + 空子命令隐藏）。
- [ ] 🚫 `plugin` 插件机制（原版 `genericcmd.HandlePluginCommand`）：**建议不纳入简版**，
  依赖 kubectl 插件加载，与配置管理教学无关，保持简版「无插件」设计。
- [ ] 🚫 `emoji`/`semver`/`addlicense`/`cleanupzombies`/`sysload`/`upgrade`：业务命令，**按需**挑选 1–2 个作为范例，
  不必全量对齐。

---

### 阶段 D：版本与构建变量（低风险）
- [ ] 🔲 `version` 子命令支持 `--output yaml|json` 与 `--short`（对齐 `cmd/version/version.go`）。
- [ ] 🔲 引入 build-time 版本变量（ldflags 注入 `version.Info`）：简版可在 `main.go` 用 `var version = "dev"` 占位，
  并在 `version` 命令读取；保持对 `onexstack/pkg/version` 的可选依赖。

---

### 阶段 E：文档与示例收尾
- [ ] 🔲 本文每完成一项，更新 `状态` 列与 `cli-config-management.md` 的差异表。
- [ ] 🔲 README 目录结构随子命令扩充同步更新。

---

## 2. 配置结构体差异明细（简版 → 原版）

### `UserOptions`（全局 `user.*`）
| 字段 | 简版 | 原版 | 说明 |
|------|------|------|------|
| `username` | ✅ | ✅ | 基础认证用户名 |
| `token` | ❌ | ✅ | Bearer token |
| `password` | ❌ | ✅ | 基础认证密码 |
| `secret-id` | ❌ | ✅ | JWT SecretID |
| `secret-key` | ❌ | ✅ | JWT SecretKey |
| `client-certificate` | ❌ | ✅ | TLS 客户端证书 |
| `client-key` | ❌ | ✅ | TLS 客户端密钥 |

### `ServerOptions`（全局 `usercenter.*` / `gateway.*`）
| 字段 | 简版 | 原版 | 说明 |
|------|------|------|------|
| `addr` | ✅（命名 `addr`） | ⚠️ `address` | **flag 名不同**，对齐需改为 `address` |
| `insecure-skip-tls-verify` | ❌ | ✅ | 跳过 TLS 校验 |
| `certificate-authority` | ❌ | ✅ | CA 证书 |
| `timeout` | ❌ | ✅ | 请求超时 |
| `max-retries` | ❌ | ✅ | 最大重试 |
| `retry-interval` | ❌ | ✅ | 重试间隔 |

> 注：`es` 领域配置 `ESOptions`（简版新增）**不属于**全局 `Options`，保持独立，不在此表内。

---

## 3. 关键差异与设计取舍

1. **`--config` 加载**：原版 `cmd.go:201` 存在构造期求值 bug（见 `cli-config-management.md` §已记录），
   简版已用延迟求值 `func() *string` 修复。**对齐时不要回退到原版的写法**。
2. **插件机制**：原版依赖 kubectl 插件加载（`NewDefaultPluginHandler` 等）。简版刻意省略，保持轻量。
3. **`create` 语义**：简版用 `create project` 演示「创建项目」业务命令；原版实际是 `create cmd`
   （脚手架生成器）。两者都演示 Options 模式，阶段 C 可改为对齐原版语义。
4. **横切能力**（profiling / warnings / headers）：原版在 `PersistentPreRunE/PostRunE` 内实现，
   简版目前仅有 `opts.Complete()`。阶段 B 补齐骨架即可，不必接真实 REST 客户端。

---

## 4. 验收标准

- [ ] `go build ./...` 与 `go vet ./...` 通过（注意 `go build ./...` 会因 `tpl`/`template` 模板目录报错，
      应改用 `go build ./cmd/...` 或限定目录）。
- [ ] 配置优先级（默认值 < 文件 < 环境变量 < flag）在全局与领域配置两种场景均验证通过。
- [ ] `osbuilder options` 输出的继承 flag 与当前 `Options` 字段一致。
- [ ] 每项对齐在本文与 `cli-config-management.md` 中标注状态。
