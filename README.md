<h1 align="center">makeshop Extensions SDK for Go</h1>

<p align="center">
  <strong>makeshopの処理を拡張するWebAssemblyコンポーネントをGoで開発できます。</strong><br />
  型付きのentrypoint契約とhost APIを提供します。
</p>

<p align="center">
  SDKs: <a href="https://github.com/gmomakeshop/makeshop-extensions-sdk-go">Go</a> · <a href="https://github.com/gmomakeshop/makeshop-extensions-sdk-rust">Rust</a>
</p>

---

## 目次

- [このSDKについて](#このsdkについて)
- [導入](#導入)
- [最小実装](#最小実装)
- [DeploymentConfig](#deploymentconfig)
- [SDKの構成](#sdkの構成)

---

## このSDKについて

makeshop Extensionsは、makeshopの処理に独自のロジックを追加するための仕組みです。
本SDKを使った拡張機能の利用には、対応プランの契約が必要です。

Go SDKは次のentrypointを公開しています。

| Entrypoint | 種別 | Handler | 用途 |
|---|---|---|---|
| `finalize_checkout_validation` | Hook | `Run` | 注文確定前の検証 |
| `order_completed` | Event | `Deliver` | 注文完了後の処理 |
| `order_completed_content` | Hook | `Run` | 注文完了画面へのコンテンツ追加 |

一つのWebAssemblyコンポーネントには、一つのentrypointを実装します。
Go 1.27.0以降と[`componentize-go`](https://github.com/bytecodealliance/componentize-go) v0.3.4が必要です。

---

## 導入

> 利用するentrypointに対応したプロジェクト一式を使用してください。
> プロジェクトにはWIT契約、生成済みのexport wiring、`main` packageが含まれます。

拡張機能のGo moduleで、ビルドツールとSDKを追加します。

```console
go install github.com/bytecodealliance/componentize-go@v0.3.4
go get github.com/gmomakeshop/makeshop-extensions-sdk-go
```

Handlerを実装したら、プロジェクトに含まれるWIT worldを指定してビルドします。

```console
mkdir -p target
componentize-go -d ./wit -w finalize-checkout-validation-guest build -o target/extension.wasm
```

`order_completed`ではworldを`order-completed-guest`、`order_completed_content`では`order-completed-content-guest`へ変更します。

---

## 最小実装

### Hook

次の例は、合計金額が100,000円を超える注文を停止します。

```go
package export_makeshop_finalize_checkout_validation_handler

import (
	hook "github.com/gmomakeshop/makeshop-extensions-sdk-go/contract/finalize-checkout-validation/v1/bindings/makeshop_finalize_checkout_validation_handler"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func Run(input hook.Input) witTypes.Result[hook.Decision, hook.GuestError] {
	if input.Checkout.TotalPrice > 100_000 {
		return witTypes.Ok[hook.Decision, hook.GuestError](
			hook.MakeDecisionFail([]string{"注文金額が上限を超えています。"}),
		)
	}

	return witTypes.Ok[hook.Decision, hook.GuestError](hook.MakeDecisionPass())
}
```

`DecisionFail`のmessageは購入者に表示される可能性があります。
運用情報はmessageに含めず、runtime logへ出力してください。

### 注文完了コンテンツ

`contentblocks.Marshal`で生成したJSONを、返却するfragmentの`BlocksJson`へ設定します。

```go
blocks, err := contentblocks.Marshal(
    contentblocks.Heading{Level: 2, Text: "ご注文ありがとうございます"},
    contentblocks.Text{Spans: []contentblocks.Span{{Text: "発送後にメールでお知らせします。"}}},
)
if err != nil {
    return witTypes.Err[hook.Output, hook.GuestError](hook.GuestError{
        Code: "content_encoding_failed", Message: "Unable to encode content.",
    })
}
return witTypes.Ok[hook.Output, hook.GuestError](hook.Output{
    SchemaVersion: contentblocks.SchemaVersion,
    Fragments: []hook.ContentFragment{{Placement: "main", BlocksJson: blocks}},
    Data: []hook.ContentEntry{},
})
```

この例では`contentblocks`をSDKの`contentblocks` package、`hook`を`contract/order-completed-content/v1/bindings/makeshop_order_completed_content_handler`からimportします。
各structはpayloadを表し、`Marshal`または`Blocks`のJSONエンコードが`type`を付与します。
コンテナ階層、文字列長、URL、出力制限はruntimeが検証します。

### Event

次の例は、注文完了Eventを受け取り、処理済みとして返します。

```go
package export_makeshop_order_completed_handler

import (
	eventTypes "github.com/gmomakeshop/makeshop-extensions-sdk-go/contract/order-completed/v1/bindings/makeshop_event_types_types"
	event "github.com/gmomakeshop/makeshop-extensions-sdk-go/contract/order-completed/v1/bindings/makeshop_order_completed_handler"
	"github.com/gmomakeshop/makeshop-extensions-sdk-go/host"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func Deliver(input event.Input) witTypes.Result[event.DeliveryResult, event.GuestError] {
	host.Info("order completed: "+input.Event.EventId, "{}")
	return witTypes.Ok[event.DeliveryResult, event.GuestError](eventTypes.MakeDeliveryResultOk())
}
```

外部サービスを呼び出す場合は`host.Call`を使用します。
一時的な失敗は`Retry`、再試行しても解消しない失敗は`DeadLetter`として返します。

`order_completed`のhandlerでは、`host.SendToBuyer`で購入者へのメールを予約できます。
この時点ではメールは送信されず、`Deliver`が`Ok`を返した後に送信されます。
`Retry`や`DeadLetter`を返した場合、その run で予約したメールは破棄されます。
`BuyerMail.Body`は自由文のテキストまたはHTMLで、`Format`（`MailBodyFormatText`/`MailBodyFormatHTML`）で指定した形式のままメールに反映されます。
本文中の`[MAIL_HEAD]`と`[MAIL_TAIL]`だけは、送信時にショップの共通ヘッダーと共通フッターへ置き換わります（件名は置き換えません）。
配送が`Ok`で完了するのは1回だけなので、同じメールが二度送られることはありません。
`BuyerMail.OperationKey`はrun内でメールを識別する名前です。ASCII英数字と`.` `_` `:` `-`からなる128文字以内の値にし、run内では一意にしてください。
メールはメルマガ受信可否にかかわらず送られます。販促を含むメールは`Order.Buyer.MailMagazineAllowed`（会員購入でのみ値があります）を確認してから予約してください。

---

## DeploymentConfig

WebAssemblyコンポーネントと`DeploymentConfig`は一つのartifactとして登録します。
EntrypointはWIT契約から判定されるため、JSONには指定しません。
`metadata.name`は表示用の名前であり、HookやEventの指定ではありません。

次の例は、`order_webhook`というconnectorを定義します。

```json
{
  "metadata": {
    "name": "order-export"
  },
  "runtime": {
    "timeout_ms": 3000
  },
  "connectors": [
    {
      "name": "order_webhook",
      "method": "POST",
      "endpoint": "https://orders.example.com/events",
      "required": false,
      "timeout_ms": 2000,
      "request": {
        "headers": {
          "Content-Type": "application/json"
        }
      }
    }
  ]
}
```

Handlerから呼び出す`ConnectorRequest.Connector`には、`connectors[].name`と同じ値を指定します。
Connectorを使わない場合は`connectors: []`とします。
本番環境のendpointにはHTTPSを使用してください。

---

## SDKの構成

| Path | 内容 |
|---|---|
| `contract/<entrypoint>/v<major>/bindings/` | Entrypointの入出力型 |
| `host/` | Connector、runtime log、実行情報、UUID、購入者メール予約、テスト用API |
| `contentblocks/` | Content Blocksの型とJSONエンコード、デコード |

Contract versionはGo moduleのversionとは独立しており、package pathの一部になります。

---

## リポジトリについて

このリポジトリは配布用のスナップショットです。
Issueは受け付けていますが、Pull Requestは受け付けていません。

## ライセンス

[LICENSE](LICENSE)を参照してください。
