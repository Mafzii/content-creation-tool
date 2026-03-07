package models

type Identifiable interface {
	GetId() int
}

type Topic struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
}

func (t Topic) GetId() int { return t.Id }

type Source struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`         // text|url|file
	Raw         string `json:"raw"`          // original input
	Content     string `json:"content"`      // cleaned text for LLM
	Status      string `json:"status"`       // ready|pending|partial|error
	ExtractMode string `json:"extract_mode"` // standard|ai
	TopicId     int    `json:"topic_id"`     // optional topic for AI extraction context
}

func (s Source) GetId() int { return s.Id }

type Style struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Prompt  string `json:"prompt"`
	Tone    string `json:"tone"`
	Example string `json:"example"`
}

func (s Style) GetId() int { return s.Id }

type Draft struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	TopicId   int    `json:"topic_id"`
	StyleId   int    `json:"style_id"`
	Status    string `json:"status"`
	Notes     string `json:"notes"`
	SourceIds []int  `json:"source_ids"`
}

func (d Draft) GetId() int { return d.Id }

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
