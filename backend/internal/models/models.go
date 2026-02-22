package models

type Identifiable interface {
	GetId() int
}

type Topic struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (t Topic) GetId() int { return t.Id }

type Source struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (s Source) GetId() int { return s.Id }

type Style struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

func (s Style) GetId() int { return s.Id }

type Draft struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	TopicId int    `json:"topic_id"`
	StyleId int    `json:"style_id"`
	Status  string `json:"status"`
}

func (d Draft) GetId() int { return d.Id }
