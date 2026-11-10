// Code generated from the Content Blocks JSON Schema; DO NOT EDIT.
package contentblocks

import (
	"encoding/json"
	"fmt"
)

// SchemaVersion is the major version used by blocks-json.
const SchemaVersion uint32 = 1

// Accordion represents collapsible content.
type Accordion struct {
	// Blocks represents leaf blocks shown when the content is expanded.
	Blocks Blocks `json:"blocks"`
	// Summary represents summary shown while the content is collapsed.
	Summary string `json:"summary"`
}

func (Accordion) blockType() string { return "accordion" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Accordion) MarshalJSON() ([]byte, error) {
	type plain Accordion
	v.Blocks = nonNil(v.Blocks)
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Accordion) UnmarshalJSON(data []byte) error {
	type plain Accordion
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"blocks": {true, 1}, "summary": {true, 0}}); err != nil {
		return err
	}
	*v = Accordion(value)
	return nil
}

// Alert represents a container for notices and guidance.
type Alert struct {
	// Blocks represents leaf blocks displayed inside the container.
	Blocks Blocks `json:"blocks"`
	// Variant represents presentation intent.
	Variant Variant `json:"variant"`
}

func (Alert) blockType() string { return "alert" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Alert) MarshalJSON() ([]byte, error) {
	type plain Alert
	v.Blocks = nonNil(v.Blocks)
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Alert) UnmarshalJSON(data []byte) error {
	type plain Alert
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"blocks": {true, 1}, "variant": {true, 0}}); err != nil {
		return err
	}
	*v = Alert(value)
	return nil
}

// Badge represents a short label representing a rank, restriction, status, or similar property.
type Badge struct {
	// Text represents text displayed in the label.
	Text string `json:"text"`
	// Variant represents presentation intent of the label.
	Variant Variant `json:"variant"`
}

func (Badge) blockType() string { return "badge" }

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Badge) UnmarshalJSON(data []byte) error {
	type plain Badge
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"text": {true, 0}, "variant": {true, 0}}); err != nil {
		return err
	}
	*v = Badge(value)
	return nil
}

// Card represents a card containing a title, body, image, and link.
type Card struct {
	// Body represents card body.
	Body *[]Span `json:"body,omitempty"`
	// Image represents image displayed on the card.
	Image *Image `json:"image,omitempty"`
	// Link represents link opened from the card.
	Link *Link `json:"link,omitempty"`
	// Title represents card title.
	Title string `json:"title"`
}

func (Card) blockType() string { return "card" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Card) MarshalJSON() ([]byte, error) {
	type plain Card
	if v.Body != nil {
		copy0 := *v.Body
		copy0 = nonNil(copy0)
		v.Body = &copy0
	}
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Card) UnmarshalJSON(data []byte) error {
	type plain Card
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"body": {false, 1}, "image": {false, 0}, "link": {false, 0}, "title": {true, 0}}); err != nil {
		return err
	}
	*v = Card(value)
	return nil
}

// Copyable represents a value that provides a copy action.
type Copyable struct {
	// Label represents label describing the value.
	Label *string `json:"label,omitempty"`
	// Value represents value to copy.
	Value string `json:"value"`
}

func (Copyable) blockType() string { return "copyable" }

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Copyable) UnmarshalJSON(data []byte) error {
	type plain Copyable
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"label": {false, 0}, "value": {true, 0}}); err != nil {
		return err
	}
	*v = Copyable(value)
	return nil
}

// Divider represents a divider.
type Divider struct {
}

func (Divider) blockType() string { return "divider" }

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Divider) UnmarshalJSON(data []byte) error {
	type plain Divider
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{}); err != nil {
		return err
	}
	*v = Divider(value)
	return nil
}

// FieldEntry represents a label-value pair displayed in a fields block.
type FieldEntry struct {
	// Label represents label describing the value.
	Label string `json:"label"`
	// Value represents value to display.
	Value string `json:"value"`
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *FieldEntry) UnmarshalJSON(data []byte) error {
	type plain FieldEntry
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"label": {true, 0}, "value": {true, 0}}); err != nil {
		return err
	}
	*v = FieldEntry(value)
	return nil
}

// Fields represents a block that displays labels alongside their values.
type Fields struct {
	// Entries represents labels and values to display.
	Entries []FieldEntry `json:"entries"`
}

func (Fields) blockType() string { return "fields" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Fields) MarshalJSON() ([]byte, error) {
	type plain Fields
	v.Entries = nonNil(v.Entries)
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Fields) UnmarshalJSON(data []byte) error {
	type plain Fields
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"entries": {true, 1}}); err != nil {
		return err
	}
	*v = Fields(value)
	return nil
}

// Grid represents a container that displays cards or images in a grid.
type Grid struct {
	// ColumnsHint represents preferred column count. The rendered column count may vary by display environment.
	ColumnsHint *uint8 `json:"columns-hint,omitempty"`
	// Items represents cards or images to display.
	Items Blocks `json:"items"`
}

func (Grid) blockType() string { return "grid" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Grid) MarshalJSON() ([]byte, error) {
	type plain Grid
	v.Items = nonNil(v.Items)
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Grid) UnmarshalJSON(data []byte) error {
	type plain Grid
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"columns-hint": {false, 0}, "items": {true, 1}}); err != nil {
		return err
	}
	*v = Grid(value)
	return nil
}

// Heading represents a heading.
type Heading struct {
	// Level represents heading level.
	Level uint8 `json:"level"`
	// Text represents heading text.
	Text string `json:"text"`
}

func (Heading) blockType() string { return "heading" }

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Heading) UnmarshalJSON(data []byte) error {
	type plain Heading
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"level": {true, 0}, "text": {true, 0}}); err != nil {
		return err
	}
	*v = Heading(value)
	return nil
}

// Image represents an image.
type Image struct {
	// Alt represents alternative text for the image.
	Alt string `json:"alt"`
	// URL represents hTTPS URL of the image.
	URL string `json:"url"`
}

func (Image) blockType() string { return "image" }

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Image) UnmarshalJSON(data []byte) error {
	type plain Image
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"alt": {true, 0}, "url": {true, 0}}); err != nil {
		return err
	}
	*v = Image(value)
	return nil
}

// Link represents a link to another page.
type Link struct {
	// Prominence represents visual prominence.
	Prominence *Prominence `json:"prominence,omitempty"`
	// Text represents link text.
	Text string `json:"text"`
	// URL represents hTTPS destination URL.
	URL string `json:"url"`
}

func (Link) blockType() string { return "link" }

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Link) UnmarshalJSON(data []byte) error {
	type plain Link
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"prominence": {false, 0}, "text": {true, 0}, "url": {true, 0}}); err != nil {
		return err
	}
	*v = Link(value)
	return nil
}

// List represents a bulleted or numbered list.
type List struct {
	// Items represents list items.
	Items [][]Span `json:"items"`
	// Style represents marker style.
	Style ListStyle `json:"style"`
}

func (List) blockType() string { return "list" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v List) MarshalJSON() ([]byte, error) {
	type plain List
	v.Items = nonNil(v.Items)
	v.Items = append(make([][]Span, 0, len(v.Items)), v.Items...)
	for i0 := range v.Items {
		v.Items[i0] = nonNil(v.Items[i0])
	}
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *List) UnmarshalJSON(data []byte) error {
	type plain List
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"items": {true, 2}, "style": {true, 0}}); err != nil {
		return err
	}
	*v = List(value)
	return nil
}

// ListStyle represents marker style for a list block.
type ListStyle string

const (
	// ListStyleBullet represents a bulleted list.
	ListStyleBullet ListStyle = "bullet"
	// ListStyleOrdered represents a numbered list.
	ListStyleOrdered ListStyle = "ordered"
)

// UnmarshalJSON decodes a supported ListStyle value.
func (v *ListStyle) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value {
	case "bullet", "ordered":
		*v = ListStyle(value)
		return nil
	default:
		return fmt.Errorf("unknown ListStyle: %q", value)
	}
}

// Prominence represents visual prominence of a link.
type Prominence string

const (
	// ProminencePrimary represents a primary action.
	ProminencePrimary Prominence = "primary"
	// ProminenceSecondary represents a secondary action.
	ProminenceSecondary Prominence = "secondary"
)

// UnmarshalJSON decodes a supported Prominence value.
func (v *Prominence) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value {
	case "primary", "secondary":
		*v = Prominence(value)
		return nil
	default:
		return fmt.Errorf("unknown Prominence: %q", value)
	}
}

// QR represents a value rendered as a QR code.
type QR struct {
	// Caption represents caption displayed with the QR code.
	Caption *string `json:"caption,omitempty"`
	// Value represents value encoded in the QR code.
	Value string `json:"value"`
}

func (QR) blockType() string { return "qr" }

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *QR) UnmarshalJSON(data []byte) error {
	type plain QR
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"caption": {false, 0}, "value": {true, 0}}); err != nil {
		return err
	}
	*v = QR(value)
	return nil
}

// Section represents a group of content with an optional heading.
type Section struct {
	// Blocks represents blocks displayed in the section. A section cannot contain another section.
	Blocks Blocks `json:"blocks"`
	// Heading represents optional heading.
	Heading *string `json:"heading,omitempty"`
}

func (Section) blockType() string { return "section" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Section) MarshalJSON() ([]byte, error) {
	type plain Section
	v.Blocks = nonNil(v.Blocks)
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Section) UnmarshalJSON(data []byte) error {
	type plain Section
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"blocks": {true, 1}, "heading": {false, 0}}); err != nil {
		return err
	}
	*v = Section(value)
	return nil
}

// Span represents a span within a text block.
type Span struct {
	// Emphasis represents whether to apply weak emphasis.
	Emphasis *bool `json:"emphasis,omitempty"`
	// Strong represents whether to apply strong emphasis.
	Strong *bool `json:"strong,omitempty"`
	// Text represents text to display.
	Text string `json:"text"`
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Span) UnmarshalJSON(data []byte) error {
	type plain Span
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"emphasis": {false, 0}, "strong": {false, 0}, "text": {true, 0}}); err != nil {
		return err
	}
	*v = Span(value)
	return nil
}

// Table represents a table containing column headings and rows.
type Table struct {
	// Columns represents column headings.
	Columns []string `json:"columns"`
	// Rows represents table rows. Each row must contain the same number of cells as there are columns.
	Rows [][]string `json:"rows"`
}

func (Table) blockType() string { return "table" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Table) MarshalJSON() ([]byte, error) {
	type plain Table
	v.Columns = nonNil(v.Columns)
	v.Rows = nonNil(v.Rows)
	v.Rows = append(make([][]string, 0, len(v.Rows)), v.Rows...)
	for i0 := range v.Rows {
		v.Rows[i0] = nonNil(v.Rows[i0])
	}
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Table) UnmarshalJSON(data []byte) error {
	type plain Table
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"columns": {true, 1}, "rows": {true, 2}}); err != nil {
		return err
	}
	*v = Table(value)
	return nil
}

// Text represents a paragraph.
type Text struct {
	// Spans represents spans to display.
	Spans []Span `json:"spans"`
}

func (Text) blockType() string { return "text" }

// MarshalJSON encodes the payload, representing nil slices as empty arrays.
func (v Text) MarshalJSON() ([]byte, error) {
	type plain Text
	v.Spans = nonNil(v.Spans)
	return json.Marshal(plain(v))
}

// UnmarshalJSON decodes the payload and rejects missing or unknown fields.
func (v *Text) UnmarshalJSON(data []byte) error {
	type plain Text
	var value plain
	if err := decodeObject(data, &value, map[string]fieldShape{"spans": {true, 1}}); err != nil {
		return err
	}
	*v = Text(value)
	return nil
}

// Variant represents presentation intent for alerts and badges.
type Variant string

const (
	// VariantAuto represents default style of the display environment.
	VariantAuto Variant = "auto"
	// VariantInfo represents informational content.
	VariantInfo Variant = "info"
	// VariantSuccess represents successful outcome.
	VariantSuccess Variant = "success"
	// VariantWarning represents warning or caution.
	VariantWarning Variant = "warning"
)

// UnmarshalJSON decodes a supported Variant value.
func (v *Variant) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value {
	case "auto", "info", "success", "warning":
		*v = Variant(value)
		return nil
	default:
		return fmt.Errorf("unknown Variant: %q", value)
	}
}
func decodeBlock(kind string, data []byte) (Block, error) {
	switch kind {
	case "accordion":
		var value Accordion
		err := json.Unmarshal(data, &value)
		return value, err
	case "alert":
		var value Alert
		err := json.Unmarshal(data, &value)
		return value, err
	case "badge":
		var value Badge
		err := json.Unmarshal(data, &value)
		return value, err
	case "card":
		var value Card
		err := json.Unmarshal(data, &value)
		return value, err
	case "copyable":
		var value Copyable
		err := json.Unmarshal(data, &value)
		return value, err
	case "divider":
		var value Divider
		err := json.Unmarshal(data, &value)
		return value, err
	case "fields":
		var value Fields
		err := json.Unmarshal(data, &value)
		return value, err
	case "grid":
		var value Grid
		err := json.Unmarshal(data, &value)
		return value, err
	case "heading":
		var value Heading
		err := json.Unmarshal(data, &value)
		return value, err
	case "image":
		var value Image
		err := json.Unmarshal(data, &value)
		return value, err
	case "link":
		var value Link
		err := json.Unmarshal(data, &value)
		return value, err
	case "list":
		var value List
		err := json.Unmarshal(data, &value)
		return value, err
	case "qr":
		var value QR
		err := json.Unmarshal(data, &value)
		return value, err
	case "section":
		var value Section
		err := json.Unmarshal(data, &value)
		return value, err
	case "table":
		var value Table
		err := json.Unmarshal(data, &value)
		return value, err
	case "text":
		var value Text
		err := json.Unmarshal(data, &value)
		return value, err
	default:
		return nil, fmt.Errorf("unknown block type: %q", kind)
	}
}
