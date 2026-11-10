package contentblocks_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	cb "github.com/gmomakeshop/makeshop-extensions-sdk-go/contentblocks"
)

func ptr[T any](v T) *T { return &v }

func allBlocks() cb.Blocks {
	return cb.Blocks{
		cb.Heading{Level: 1, Text: "ご注文ありがとうございます"},
		cb.Text{Spans: []cb.Span{{Text: "通常", Strong: ptr(false)}, {Text: "強調", Strong: ptr(true), Emphasis: ptr(true)}}},
		cb.Fields{Entries: []cb.FieldEntry{{Label: "合計", Value: "1,200円"}}},
		cb.Copyable{Value: "受取時にお伝えください", Label: ptr("ご案内")},
		cb.Link{Text: "詳細", URL: "https://example.com/details", Prominence: ptr(cb.ProminencePrimary)},
		cb.Image{URL: "https://example.com/image.png", Alt: "商品画像"},
		cb.Divider{},
		cb.Table{Columns: []string{"商品", "数量"}, Rows: [][]string{{"りんご", "2"}}},
		cb.List{Style: cb.ListStyleBullet, Items: [][]cb.Span{{{Text: "ご案内"}}}},
		cb.Card{Title: "おすすめ", Body: &[]cb.Span{{Text: "本文"}}, Image: &cb.Image{URL: "https://example.com/card.png", Alt: "カード画像"}, Link: &cb.Link{Text: "詳細", URL: "https://example.com/card", Prominence: ptr(cb.ProminenceSecondary)}},
		cb.Badge{Variant: cb.VariantAuto, Text: "おすすめ"},
		cb.QR{Value: "受取のご案内", Caption: ptr("レジでご提示ください")},
		cb.Alert{Variant: cb.VariantSuccess, Blocks: cb.Blocks{cb.Text{Spans: []cb.Span{{Text: "完了しました"}}}}},
		cb.Section{Heading: ptr("関連商品"), Blocks: cb.Blocks{cb.Grid{ColumnsHint: ptr(uint8(2)), Items: cb.Blocks{cb.Card{Title: "関連商品"}, cb.Image{URL: "https://example.com/related.png", Alt: "関連画像"}}}}},
		cb.Grid{Items: cb.Blocks{cb.Card{Title: "おすすめ"}}},
		cb.Accordion{Summary: "ご案内", Blocks: cb.Blocks{cb.Badge{Variant: cb.VariantInfo, Text: "確認事項"}}},
	}
}

func TestRoundTrip(t *testing.T) {
	blocks := allBlocks()
	encoded, err := cb.Marshal(blocks...)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := cb.Parse(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed, blocks) {
		t.Fatalf("round trip differs:\n%#v\n%#v", parsed, blocks)
	}
	var values []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(encoded), &values); err != nil {
		t.Fatal(err)
	}
	var card map[string]json.RawMessage
	for _, v := range values {
		if string(v["type"]) == `"card"` {
			card = v
		}
	}
	for _, field := range []string{"image", "link"} {
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(card[field], &payload); err != nil {
			t.Fatal(err)
		}
		if _, ok := payload["type"]; ok {
			t.Fatalf("card.%s contains a block discriminator", field)
		}
	}
}

func TestNilAndOptionalValues(t *testing.T) {
	for _, blocks := range []cb.Blocks{nil, {}} {
		got, err := cb.Marshal(blocks...)
		if err != nil || got != "[]" {
			t.Fatalf("empty blocks: %q %v", got, err)
		}
	}
	var image *cb.Image
	for _, blocks := range []cb.Blocks{{nil}, {image}, {cb.Section{Blocks: cb.Blocks{nil}}}} {
		if _, err := cb.Marshal(blocks...); err == nil {
			t.Fatal("accepted nil block")
		}
	}
	for _, data := range []string{`[{"type":"text","spans":[{"text":"a"}]}]`, `[{"type":"text","spans":[{"text":"a","strong":false}]}]`} {
		parsed, err := cb.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := cb.Marshal(parsed...)
		if err != nil {
			t.Fatal(err)
		}
		var want, got any
		_ = json.Unmarshal([]byte(data), &want)
		_ = json.Unmarshal([]byte(encoded), &got)
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("optional field changed: %s", encoded)
		}
	}
	for _, block := range []cb.Block{cb.Text{}, cb.Grid{}, cb.Section{}, cb.Table{Rows: [][]string{nil}}, cb.Card{Body: ptr([]cb.Span(nil))}} {
		encoded, err := cb.Marshal(block)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := cb.Parse(encoded); err != nil {
			t.Fatalf("nil slices did not encode as arrays: %s: %v", encoded, err)
		}
	}
}

var malformed = []string{
	`null`, `{}`, `[null]`, `[{}]`, `[{"type":null}]`, `[{"type":"future"}]`,
	`[{"type":"heading","text":"a"}]`, `[{"type":"heading","text":"a","level":null}]`,
	`[{"type":"heading","text":"a","level":256}]`, `[{"type":"heading","text":"a","level":1.2}]`,
	`[{"type":"divider","extra":true}]`, `[{"type":"divider","type":"text"}]`,
	`[{"type":"text","spans":[{"text":"a","text":"b"}]}]`,
	`[{"type":"text","spans":[{"text":"a","Text":"b"}]}]`,
	`[{"type":"text","spans":null}]`, `[{"type":"text","spans":[null]}]`,
	`[{"type":"table","columns":[null],"rows":[["a"]]}]`,
	`[{"type":"table","columns":["a"],"rows":[null]}]`,
	`[{"type":"table","columns":["a"],"rows":[[null]]}]`,
	`[{"type":"badge","variant":"future","text":"a"}]`,
	`[{"type":"card","title":"a","image":{"type":"image","url":"https://example.com/a","alt":"a"}}]`,
	`[] true`,
}

func TestMalformedJSON(t *testing.T) {
	for _, data := range malformed {
		t.Run(data, func(t *testing.T) {
			if _, err := cb.Parse(data); err == nil {
				t.Fatal("accepted malformed JSON")
			}
		})
	}
}

// These values are structurally valid Go payloads. The runtime rejects their content.
var invalidContent = cb.Blocks{
	cb.Grid{Items: cb.Blocks{cb.Text{Spans: []cb.Span{{Text: "invalid grid item"}}}}},
	cb.Section{Blocks: cb.Blocks{cb.Section{Blocks: cb.Blocks{cb.Divider{}}}}},
	cb.Alert{Variant: cb.VariantWarning, Blocks: cb.Blocks{cb.Grid{Items: cb.Blocks{cb.Card{Title: "a"}}}}},
	cb.Table{Columns: []string{"a", "b"}, Rows: [][]string{{"a"}}},
	cb.Link{Text: "a", URL: "http://example.com"},
	cb.Badge{Variant: cb.VariantInfo, Text: ""},
}

func TestInteroperability(t *testing.T) {
	encoded, err := cb.Marshal(allBlocks()...)
	if err != nil {
		t.Fatal(err)
	}
	cases := struct {
		Valid   []string `json:"valid"`
		Invalid []string `json:"invalid"`
	}{Valid: []string{encoded}, Invalid: append([]string{}, malformed...)}
	for _, file := range []string{"order-completed-main.blocks.json"} {
		data, err := os.ReadFile(filepath.Join("testdata", file))
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := cb.Parse(string(data))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := cb.Marshal(parsed...)
		if err != nil {
			t.Fatal(err)
		}
		cases.Valid = append(cases.Valid, encoded)
	}
	for _, block := range invalidContent {
		encoded, err := cb.Marshal(block)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := cb.Parse(encoded); err != nil {
			t.Fatalf("unexpected content validation: %v", err)
		}
		cases.Invalid = append(cases.Invalid, encoded)
	}
	if output := os.Getenv("CONTENT_BLOCKS_INTEROP_OUT"); output != "" {
		data, err := json.Marshal(cases)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(output, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
