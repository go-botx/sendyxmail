package apiv0

type Message struct {
	To      string      `json:"to"`
	Body    string      `json:"body"`
	Buttons []ButtonRow `json:"buttons"`
}

type ButtonRow []Button

type Button struct {
	Label           string `json:"label"`
	Link            string `json:"link"`
	TextColor       string `json:"text_color,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextAlign       string `json:"text_align,omitempty"`
	AlertText       string `json:"alert_text,omitempty"`
	HorizontalSize  int    `json:"h_size,omitempty"`
}
