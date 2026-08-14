# osbuilder CLI 新人教程：从零逐步复刻复杂设计

> 本教程以 `cmd/osbuilder/osbuilder.go`（入口）→ `internal/osbuilder/cmd/cmd.go`（根命令）的**真实设计为蓝本**，
> 循序渐进地带你从「一个只打印 help 的根命令」逐步叠加出包含配置管理、依赖注入、命令分组、横切钩子、
> 插件机制的完整 CLI。
>
> 每一步都可独立编译运行。教程中的可运行参考实现见 `examples/osbuilder-simplified/`，
> 关键设计决策与原文一一对应。读完本教程，你应能理解 `osbuilder` 命令的每一个设计细节，
> 以及「为什么这样设计」。

---

## 设计全景（先看终点）

```
osbuilder 命令整体架构
┌─────────────────────────────────────────────────────────────┐
│ main()                                                        │
│   └─ cmd.NewDefaultOSCtlCommand()  → 构建整棵命令树            │
│        ├─ OSCtlOptions (PluginHandler / IOStreams / Args)      │
│        ├─ NewDefaultOSCtlCommandWithArgs()  → 插件探测         │
│        └─ NewOSCtlCommand()  → 根命令 + 全部子命令            │
│             ├─ PersistentFlags: --config / profiling / warnings│
│             ├─ AddFlags(opts)  → 全局配置 flag                 │
│             ├─ BindPFlags + OnInitialize  → viper 加载配置     │
│             ├─ PersistentPreRunE / PostRunE  → 横切钩子        │
│             ├─ NewFactory(opts)  → 依赖注入                    │
│             ├─ CommandGroups  → 命令分组展示                  │
│             └─ AddCommand(color/create/version/options/...)   │
└─────────────────────────────────────────────────────────────┘
```

> 上图是**调用层次（自顶向下）**：`main` 在最上，逐层向下构建出整棵命令树。
> 而下文的教程步骤是**构建难度（由简入繁）**：先讲最基础的空根命令，再逐步往上叠加配置、注入、钩子、分组等能力。
> 两者方向相反，但「先懂构件、再懂组装」是最适合新人的讲解顺序。

---

## 第 0 步：最小可运行的根命令

**目标**：理解 cobra 命令树的最小骨架——一个入口 + 一个根命令。

```go
// main.go
package main

import "github.com/spf13/cobra"

func main() {
	cmd := &cobra.Command{
		Use:   "osbuilder",
		Short: "osbuilder is a command-line tool for the onex stack scaffold",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help() // 无子命令时打印帮助
		},
	}
	_ = cmd.Execute()
}
```

运行 `go run .` 会打印帮助。这就是原版 `osbuilder.go` 的雏形——真实入口只有 6 行：

```go
// cmd/osbuilder/osbuilder.go（真实代码）
func main() {
	command := cmd.NewDefaultOSCtlCommand()
	if err := cli.RunNoErrOutput(command); err != nil {
		util.CheckErr(err)
	}
}
```

> **设计要点**：入口极薄，所有复杂度都藏在 `cmd.NewDefaultOSCtlCommand()` 里。
> 这样做是为了让「构建命令树」与「运行命令」解耦，便于测试和注入参数（见第 6 步插件探测）。

---

## 第 1 步：Options 模式（Complete / Validate / Run）

**目标**：引入 CLI 子命令的标准三阶段生命周期，这是整个设计的核心范式。

每个命令都持有自己的 `Options` 结构体，Run 时按 `Complete → Validate → Run` 顺序执行：

```go
type VersionOptions struct {
	Short  bool
	Output string
	genericiooptions.IOStreams // 标准出入流，便于测试时替换
}

func (o *VersionOptions) Complete() error        { return nil }
func (o *VersionOptions) Validate() error {
	if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
		return errors.New(`--output must be 'yaml' or 'json'`)
	}
	return nil
}
func (o *VersionOptions) Run() error {
	fmt.Fprintf(o.Out, "Client Version: %s\n", "v0.0.1")
	return nil
}

func NewCmdVersion(streams genericiooptions.IOStreams) *cobra.Command {
	o := &VersionOptions{IOStreams: streams}
	cmd := &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			// 三阶段：补全 → 校验 → 执行
			cmdutil.CheckErr(o.Complete())
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().BoolVar(&o.Short, "short", o.Short, "print just the version number")
	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'")
	return cmd
}
```

> **设计要点**（对应原版 `cmd/version/version.go`）：
> - `Complete`：把外部输入（flag、环境变量、文件）回填到结构体，做默认值补全。
> - `Validate`：校验必填/互斥/格式，返回聚合错误。
> - `Run`：纯执行，不再读取 flag。
> - `genericiooptions.IOStreams` 注入 `In/Out/ErrOut`，使命令可测试（测试时替换为 `bytes.Buffer`）。
> - `cmdutil.CheckErr` 统一处理错误退出码。

---

## 第 2 步：全局配置结构体与 AddFlags

**目标**：把分散的 flag 收敛到一个全局 `Options`，通过 `AddFlags` 挂到根命令的 `PersistentFlags`。

```go
// util/options/options.go（对应原版）
type Options struct {
	Config           string        `json:"config"`
	UserOptions      *UserOptions  `json:"user"           mapstructure:"user"`
	UserCenterOptions *ServerOptions `json:"usercenter"   mapstructure:"usercenter"`
	GatewayOptions   *ServerOptions `json:"gateway"       mapstructure:"gateway"`
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Config, "config", o.Config, "Path to the config file to use for CLI.")
	o.UserOptions.AddFlags(fs)                    // 注册 user.username / user.token ...
	o.UserCenterOptions.AddFlags(fs, "usercenter") // 注册 usercenter.address ...
	o.GatewayOptions.AddFlags(fs, "gateway")       // 注册 gateway.address ...
}
```

`UserOptions` 用前缀 `user.` 注册认证字段，`ServerOptions` 用可变前缀注册多组服务地址：

```go
// util/options/server_options.go（对应原版）
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Addr, join(prefixes...)+"address", o.Addr,
		"The address and port of the OneX API server")
	// ...insecure-skip-tls-verify / certificate-authority / timeout / max-retries / retry-interval
}
```

> **设计要点**：
> - 全局配置用 `PersistentFlags`（持久 flag），所有子命令自动继承——这就是 `osbuilder options` 命令能列出「所有命令都继承的 flag」的原因。
> - `mapstructure` tag 让 viper 能把 YAML 文件反序列化回结构体。
> - `UserOptions.Validate()` 强制「username/password 必须同时设置」「secret-id/secret-key 必须同时设置」。

---

## 第 3 步：配置加载（viper + BindPFlags + OnInitialize）

**目标**：让配置支持三种来源——命令行 flag、环境变量、配置文件，且按优先级合并。

```go
opts := clioptions.NewOptions()
flags := cmds.PersistentFlags()
opts.AddFlags(flags)

// 1) 把根命令的持久 flag 镜像进 viper
_ = viper.BindPFlags(cmds.PersistentFlags())

// 2) 注册「执行前钩子」：命令执行前读配置文件 + 环境变量
cobra.OnInitialize(core.OnInitialize(
	ptr.To(viper.GetString("config")), // ← 注意这里的陷阱，见下方警告
	"OSCTL",          // 环境变量前缀：OSCTL_USER_USERNAME
	searchDirs(),     // 搜索目录：~/.onexstack 和 .
	defaultConfigName, // "osbuilder.yaml"
))
```

配置优先级（低 → 高）：

```
默认值 < osbuilder.yaml 的 user.* < 环境变量 OSCTL_USER_* < --config 指定文件 < 命令行 flag
```

> ⚠️ **原版真实 bug（重要）**：原版 `cmd.go:201` 写成
> `core.OnInitialize(ptr.To(viper.GetString(clioptions.FlagConfig)), ...)`
> 这里 `viper.GetString("config")` 在**构造根命令期**就被求值，而此时命令行 flag 还没解析，返回 `""`。
> `ptr.To("")` 得到**非 nil** 指针（指向空串），闭包里 `if configFile != nil` 为真 → `viper.SetConfigFile("")` →
> `ReadInConfig` 找不到空路径文件，错误被 `_` 忽略 → **`--config` 指定的文件被静默丢弃**。
>
> 简化版已修复：改为延迟求值 `func() *string`，
> 在 `OnInitialize` 闭包真正执行时（命令执行前、flag 已解析）才读取 `--config`。
> 对齐真实代码时，**不要回退到原版的写法**。

---

## 第 4 步：Factory 依赖注入

**目标**：用 Factory 解耦「配置持有」与「命令使用」，让子命令无需关心配置怎么加载。

```go
// cmd/util/factory.go（对应原版）
type Factory interface {
	GetOptions() *clioptions.Options
}

func NewFactory(opts *clioptions.Options) Factory {
	return &factoryImpl{opts: opts}
}
```

根命令构造时创建 Factory，注入每个子命令：

```go
f := cmdutil.NewFactory(opts)

groups := templates.CommandGroups{
	{Message: "Basic Commands (Beginner):", Commands: []*cobra.Command{
		color.NewCmdColor(f, ioStreams),
	}},
	{Message: "Project Commands:", Commands: []*cobra.Command{
		create.NewCmdCreate(f, ioStreams),
	}},
}
groups.Add(cmds)
```

> **设计要点**（kubectl 范式）：
> - `Factory` 是「能力接口」，当前只有 `GetOptions()`；真实 kubectl 的 Factory 还包含
>   `ToRESTMapper`/`ToRawKubeConfigLoader` 等。这里保持最小接口，方便后续扩展。
> - 子命令签名统一为 `NewXxxCmd(f cmdutil.Factory, ioStreams)`，测试时可传入 mock Factory。

---

## 第 5 步：PersistentPreRunE / PersistentPostRunE 横切钩子

**目标**：在「每条命令执行前后」插入全局逻辑（配置补全、性能分析、warning 处理）。

```go
cmds := &cobra.Command{
	Use: "osbuilder",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		rest.SetDefaultWarningHandler(warningHandler) // 设置服务端 warning 处理
		opts.Complete()                                 // 把 viper 配置反序列化进 opts
		return initProfiling()                          // 启动 CPU/heap profile
	},
	PersistentPostRunE: func(*cobra.Command, []string) error {
		if err := flushProfiling(); err != nil {       // 写出 profile 文件
			return err
		}
		if warningsAsErrors {                            // --warnings-as-errors
			if c := warningHandler.WarningCount(); c > 0 {
				return fmt.Errorf("%d warnings received", c)
			}
		}
		return nil
	},
}
```

`opts.Complete()` 内部就是 `viper.Unmarshal(&o)`：

```go
func (o *Options) Complete() {
	if err := viper.Unmarshal(&o); err != nil {
		panic(err)
	}
}
```

> **设计要点**：
> - `PersistentPreRunE` **先于任何子命令的 Run** 执行，是做「全局配置加载」的最佳位置。
> - profiling 通过 `--profile cpu --profile-output p.pprof` 开启，不改业务代码。
> - `addProfilingFlags(flags)` 在根命令注册 profiling flag（见 `cmd/profiling.go`）。

---

## 第 6 步：插件机制与入口参数注入

**目标**：理解原版 `osbuilder.go` 为什么把入口拆成 `NewDefaultOSCtlCommand` +
`NewDefaultOSCtlCommandWithArgs` + `NewOSCtlCommand` 三层。

```go
func NewDefaultOSCtlCommand() *cobra.Command {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	return NewDefaultOSCtlCommandWithArgs(OSCtlOptions{
		PluginHandler: genericcmd.NewDefaultPluginHandler([]string{"osbuilder"}),
		Arguments:     os.Args,
		IOStreams:     ioStreams,
	})
}

func NewDefaultOSCtlCommandWithArgs(o OSCtlOptions) *cobra.Command {
	cmd := NewOSCtlCommand(o)

	// 若用户输入的 "osbuilder xxx" 不是内置命令，则探测同名插件可执行文件
	if o.PluginHandler != nil && len(o.Arguments) > 1 {
		cmdPathPieces := o.Arguments[1:]
		if foundCmd, foundArgs, err := cmd.Find(cmdPathPieces); err != nil {
			// 没找到内置命令 → 尝试作为插件运行（kubectl 插件机制）
			genericcmd.HandlePluginCommand(o.PluginHandler, cmdPathPieces, 1)
		} else if err == nil {
			// 内置命令存在但子命令可能不存在 → 同样允许插件接管
			// ...（略，见原版 cmd.go:105-128）
		}
	}
	return cmd
}
```

> **设计要点**：
> - 三层拆分是为了**可测试**：测试可传入自定义 `Arguments` 和 `IOStreams`，无需真读 `os.Args`/`os.Stdout`。
> - 插件机制让 `osbuilder` 像 kubectl 一样支持外部二进制（`osbuilder-xxx`）作为子命令。
> - 简化版**刻意不实现**插件机制（与配置管理教学无关），保持轻量。

---

## 第 7 步：命令分组、alpha 过滤与 options 帮助命令

**目标**：让 `osbuilder --help` 输出按业务分组、隐藏空分组、并支持 `osbuilder options`。

```go
groups := templates.CommandGroups{
	{Message: "Basic Commands (Beginner):", Commands: []*cobra.Command{...}},
	{Message: "Project Commands:",         Commands: []*cobra.Command{...}},
}
groups.Add(cmds)

filters := []string{"options"}
alpha := NewCmdAlpha(f, ioStreams)
if !alpha.HasSubCommands() { // 没有 alpha 子命令则隐藏
	filters = append(filters, alpha.Name())
}
templates.ActsAsRootCommand(cmds, filters, groups...)

cmds.AddCommand(alpha)
cmds.AddCommand(plugin.NewCmdPlugin(ioStreams))
cmds.AddCommand(version.NewCmdVersion(f, ioStreams))
cmds.AddCommand(options.NewCmdOptions(ioStreams.Out)) // 打印所有继承 flag
```

- `templates.CommandGroups`：把子命令按 `Message` 分块展示。
- `filters`：从 help 中隐藏 `options`（避免它作为普通命令出现，但仍可用）。
- `NewCmdAlpha`：若构建里没有任何 alpha 子命令，则自动隐藏 `alpha`（见原版 `cmd/alpha.go`）。
- `options.NewCmdOptions`：打印所有「继承的持久 flag」，对应 `osbuilder options`。

> **设计要点**：`osbuilder options` 只打印**全局**继承 flag（`user.*` / `config`），
> 不会显示任何子命令的**领域配置** flag——这正是配置分层解耦的体现（见第 8 步）。

---

## 第 8 步：领域配置（子命令级配置，与全局解耦）

**目标**：理解「连接 ES 这类只服务于单一子命令的配置」不属于全局 `Options`。

原版 `Options` 混入所有全局配置会让根命令 flag 表无限膨胀。正确做法：把领域配置放进子命令自己的 options 包。

```go
// internal/cmd/es/options/options.go（简化版教学范例）
type ESOptions struct {
	Addr     string `mapstructure:"addr"`
	Index    string `mapstructure:"index"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func (o *ESOptions) AddFlags(fs *pflag.FlagSet) {
	// 挂在子命令自身（非 persist），前缀 es. 隔离
	fs.StringVar(&o.Addr, "es.addr", o.Addr, "Elasticsearch address")
	fs.StringVar(&o.Index, "es.index", o.Index, "Elasticsearch index")
}

func (o *ESOptions) Complete(cmd *cobra.Command) {
	// 底：flag 解析值（含默认值）
	o.Addr = cmd.Flags().Lookup("es.addr").Value.String()
	// 高优先级覆盖：viper 中文件/环境变量显式设置的值
	if viper.IsSet("es.addr") {
		o.Addr = viper.GetString("es.addr")
	}
}
```

优先级与全局完全一致：

```
flag 默认值 < osbuilder.yaml 的 es.* < 环境变量 OSCTL_ES_* < --es.addr 显式值
```

> **设计要点**：
> - 领域配置的 key 用前缀隔离（`es.addr` ↔ 环境变量 `OSCTL_ES_ADDR`）。
> - `es get` 自己持有 `ESOptions`，不通过 `Factory.GetOptions()` 暴露——`Factory` 只承载全局配置。
> - `osbuilder options` 不显示 `es.*`，证明两者解耦成功。

---

## 第 9 步：横切能力收尾（command headers / warnings / 全局 normalize）

**目标**：补齐根命令在 flag 规范化与 REST 调用链路上的两个细节。

### 9.1 Flag 规范化（下划线 → 连字符）

```go
// 捕获带 "_" 的 flag 名，提示应使用 "-"
cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)
flags := cmds.PersistentFlags()
// ...
flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc) // 把其他包的 glog 等 _ 转 -
```

### 9.2 Command Headers RoundTripper（KEP 859）

```go
func addCmdHeaderHooks(cmds *cobra.Command, _ *clioptions.Options) {
	if value, exists := os.LookupEnv("OSCTL_COMMAND_HEADERS"); exists {
		if value == "false" || value == "0" {
			return // 显式关闭则不加 command headers
		}
	}
	crt := &genericclioptions.CommandHeaderRoundTripper{}
	existing := cmds.PersistentPreRunE
	cmds.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		crt.ParseCommandHeaders(cmd, args) // 解析命令路径，注入 X-Headers
		return existing(cmd, args)
	}
}
```

> **设计要点**：受 `OSCTL_COMMAND_HEADERS` 环境变量开关，把当前命令路径作为 `X-Headers` 注入所有 REST 调用，
> 用于服务端遥测。简化版保留 hook 骨架（无真实 REST 客户端），教学足够。

---

## 收尾：设计原则回顾

| 原则 | 体现 |
|------|------|
| **入口极薄** | `osbuilder.go` 仅 6 行，复杂度在 `NewDefaultOSCtlCommand` |
| **三层构造可测试** | Default → WithArgs → Command，注入 `Args`/`IOStreams` |
| **Options 三阶段** | `Complete → Validate → Run`，职责清晰 |
| **配置分层** | 全局 `Options`（持久 flag）+ 领域配置（子命令自身 flag）解耦 |
| **viper 优先级** | 默认 < 文件 < 环境变量 < flag，统一合并 |
| **Factory 注入** | 子命令不关心配置怎么加载，只调 `GetOptions()` |
| **横切钩子** | `PersistentPreRunE/PostRunE` 统管 profiling / warnings / config |
| **命令分组** | `CommandGroups` + `filters` + alpha 自动隐藏 |
| **插件扩展** | kubectl 风格插件探测（原版有，简化版省略） |

---

## 配套资源

- 真实实现：`cmd/osbuilder/osbuilder.go`、`internal/osbuilder/cmd/cmd.go`
- 配置管理专项说明：[`cli-config-management.md`](./cli-config-management.md)
- 简化版对齐计划：[`cli-alignment-plan.md`](./cli-alignment-plan.md)
- 可运行参考实现：`examples/osbuilder-simplified/`

> **已知待办（原版 bug）**：`cmd.go:201` 的 `--config` 构造期求值问题（见第 3 步警告），
> 修复方案见 `cli-config-management.md` §11 / 工作记忆，后续在原版修复。
