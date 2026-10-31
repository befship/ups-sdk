# ups-sdk

[English](./README.md) | 中文

非官方 Go SDK，基于 [UPS REST API](https://developer.ups.com/) 官方 OpenAPI 规范自动生成，并封装 OAuth、重试与稳定对外接口。

> **与 UPS 无隶属关系。** 使用前请自行注册 UPS Developer Portal，并申请 OAuth 凭证。详见 [DISCLAIMER.md](./DISCLAIMER.md)。

## 目录结构说明

| 路径 | 用途 |
|------|------|
| `ups/` | 手写封装层（OAuth 缓存/刷新、询价 / 建单 / 取消 / 轨迹 / 地址校验、错误标准化） |
| `internal/generated/` | `oapi-codegen` 根据 UPS OpenAPI 生成的代码（已提交，保证 `go get` 可用） |
| `scripts/` + `generate.sh` | 拉取官方 YAML → 修补 → 生成 |
| `.cache/specs/` | 下载的 UPS YAML（**已 gitignore，切勿提交**） |

官方 OpenAPI 文件**不会**进入 git。CI / `make generate` 会从 [UPS-API/api-documentation](https://github.com/UPS-API/api-documentation) 动态拉取。

当前已覆盖官方仓库中全部 REST API（共 33 个独立规范，有 `*-Ready.yaml` 时优先使用 Ready 版本），不仅限于 Rating / Shipping / Tracking。

## 安装

```bash
go get github.com/befship/ups-sdk@latest
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/befship/ups-sdk/ups"
)

func main() {
	client := ups.New(ups.Config{
		ClientID:     os.Getenv("UPS_CLIENT_ID"),
		ClientSecret: os.Getenv("UPS_CLIENT_SECRET"),
		Environment:  ups.EnvSandbox, // 或 ups.EnvProduction
	})

	ctx := context.Background()
	tok, err := client.AccessToken(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("token ok, len=", len(tok))
}
```

### 询价 Rate（支持多包裹）

```go
resp, err := client.Rate(ctx, ups.RateRequest{
	RateRequest: &ups.RATERequestWrapper{ /* ... */ }.RateRequest,
}, &ups.RateOptions{RequestOption: "Shop"})
```

请求体使用对外再导出的 UPS schema 类型（如 `ups.RATERequestWrapper`、包裹数组等）。完整示例见 `examples/`。

### 建单 / 取消 / 轨迹

```go
shipResp, err := client.CreateShipment(ctx, body, nil)
voidResp, err := client.VoidShipment(ctx, shipmentID, nil)
trackResp, err := client.Track(ctx, "1Z...", nil)
```

### 使用完整生成客户端

封装层未覆盖的 API，可通过 `ups` 包拿到类型化客户端（无需直接 import `internal/`）：

```go
rc, err := client.RatingClient()
sc, err := client.ShippingClient()
tc, err := client.TrackingClient()
pc, err := client.PickupClient()
lc, err := client.LandedCostClient()
// …其余见 ups/*Client()
```

`make generate` 之后，`scripts/manifest.yaml` 中列出的每个 UPS API 都会生成到 `internal/generated/<package>`。核心便捷方法覆盖：Rating、Shipping（建单 / 取消 / 面单恢复）、Tracking、Address Validation、OAuth。

## 从 UPS 规范重新生成

```bash
# 依赖：go、python3 + PyYAML、网络（或 SPEC_LOCAL_DIR）
make generate

# 从本地 UPS-API/api-documentation 克隆目录灌入
SPEC_LOCAL_DIR=/path/to/api-documentation make generate
```

流水线：

1. 从 UPS GitHub 下载 YAML（或从 `SPEC_LOCAL_DIR` 复制）
2. `scripts/preprocess.py` 修复已知 UPS OAS 问题
3. `oapi-codegen` 输出到 `internal/generated/`
4. 生成 `ups/servers.gen.go`（各 API 的 CIE / Production 基址）以及 `ups/apis.gen.go`（其余 API 的 `*Client()` 访问器）

## GitHub Actions

| Workflow | 触发方式 |
|----------|----------|
| `CI` | push / PR / 手动 |
| `Generate UPS SDK` | **仅手动**（`workflow_dispatch`）— 拉取最新规范、重新生成并开 PR |

**没有定时任务**，需要时在 Actions 页面手动跑即可。

## 许可证

- 手写代码、脚本、示例、测试：**MIT**（见 `LICENSE`）
- 生成的模型 / 客户端：源自 UPS OpenAPI，受 UPS API Access Agreement 约束（见 `DISCLAIMER.md`）

## 已支持的 API（manifest）

Address Validation、Commerce Guard、Customs Detail、Dangerous Goods、Delivery Defense / Intercept、Forwarding、Global Checkout、Landed Cost、Locator、OAuth（client credentials + auth code）、Paperless、Pickup（含 notifications / points）、Pre-Notification、Protected Delivery Token、Quantum View、**Rating**、SCS Inventory / Logistics / Transportation、**Shipping**、Smart Pickup、Time in Transit、Trade Direct、Export Assure、Track Alert（含 Enhanced）、**Tracking**、WorldEase 等 — 完整列表见 `scripts/manifest.yaml`。
