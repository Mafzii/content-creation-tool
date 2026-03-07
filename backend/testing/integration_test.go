//go:build integration

package integration

import (
	"backend/internal/handlers"
	"backend/internal/models"
	"backend/internal/store"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "open sqlite: %v\n", err)
		os.Exit(1)
	}

	if _, err := testDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		fmt.Fprintf(os.Stderr, "pragma: %v\n", err)
		os.Exit(1)
	}

	ddl := `
CREATE TABLE topics (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	keywords    TEXT NOT NULL DEFAULT ''
);
CREATE TABLE sources (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	name         TEXT NOT NULL,
	url          TEXT NOT NULL DEFAULT '',
	type         TEXT NOT NULL DEFAULT 'text',
	raw          TEXT NOT NULL DEFAULT '',
	content      TEXT NOT NULL DEFAULT '',
	status       TEXT NOT NULL DEFAULT 'ready',
	extract_mode TEXT NOT NULL DEFAULT 'standard',
	topic_id     INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE styles (
	id      INTEGER PRIMARY KEY AUTOINCREMENT,
	name    TEXT NOT NULL,
	prompt  TEXT NOT NULL,
	tone    TEXT NOT NULL DEFAULT '',
	example TEXT NOT NULL DEFAULT ''
);
CREATE TABLE drafts (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	title    TEXT NOT NULL,
	content  TEXT NOT NULL DEFAULT '',
	topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE RESTRICT,
	style_id INTEGER NOT NULL REFERENCES styles(id) ON DELETE RESTRICT,
	status   TEXT NOT NULL DEFAULT 'draft',
	notes    TEXT NOT NULL DEFAULT ''
);
CREATE TABLE draft_sources (
	draft_id  INTEGER NOT NULL REFERENCES drafts(id) ON DELETE CASCADE,
	source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
	PRIMARY KEY (draft_id, source_id)
);`
	if _, err := testDB.Exec(ddl); err != nil {
		fmt.Fprintf(os.Stderr, "migration: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}

func setupMux() *http.ServeMux {
	topics := handlers.NewCrudHandler[models.Topic](store.NewTopicSQLiteStore(testDB))
	sources := handlers.NewCrudHandler[models.Source](store.NewSourceSQLiteStore(testDB))
	styles := handlers.NewCrudHandler[models.Style](store.NewStyleSQLiteStore(testDB))
	drafts := handlers.NewCrudHandler[models.Draft](store.NewDraftSQLiteStore(testDB))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /topics", topics.GetAll)
	mux.HandleFunc("GET /topics/{id}", topics.Get)
	mux.HandleFunc("POST /topics", topics.Create)
	mux.HandleFunc("DELETE /topics/{id}", topics.Delete)

	mux.HandleFunc("GET /sources", sources.GetAll)
	mux.HandleFunc("GET /sources/{id}", sources.Get)
	mux.HandleFunc("POST /sources", sources.Create)
	mux.HandleFunc("DELETE /sources/{id}", sources.Delete)

	mux.HandleFunc("GET /styles", styles.GetAll)
	mux.HandleFunc("GET /styles/{id}", styles.Get)
	mux.HandleFunc("POST /styles", styles.Create)
	mux.HandleFunc("DELETE /styles/{id}", styles.Delete)

	mux.HandleFunc("GET /drafts", drafts.GetAll)
	mux.HandleFunc("GET /drafts/{id}", drafts.Get)
	mux.HandleFunc("POST /drafts", drafts.Create)
	mux.HandleFunc("DELETE /drafts/{id}", drafts.Delete)

	return mux
}

func TestTopicsCRUD(t *testing.T) {
	srv := httptest.NewServer(setupMux())
	defer srv.Close()

	// POST /topics
	body, _ := json.Marshal(models.Topic{Name: "computers"})
	resp, err := http.Post(srv.URL+"/topics", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /topics: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /topics: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created models.Topic
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	if created.Id == 0 {
		t.Fatal("POST /topics: returned id 0")
	}
	if created.Name != "computers" {
		t.Errorf("POST /topics: name = %q, want %q", created.Name, "computers")
	}

	// GET /topics
	resp, err = http.Get(srv.URL + "/topics")
	if err != nil {
		t.Fatalf("GET /topics: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /topics: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var all []models.Topic
	json.NewDecoder(resp.Body).Decode(&all)
	resp.Body.Close()

	if len(all) < 1 {
		t.Fatalf("GET /topics: got %d items, want >= 1", len(all))
	}

	// GET /topics/{id}
	resp, err = http.Get(fmt.Sprintf("%s/topics/%d", srv.URL, created.Id))
	if err != nil {
		t.Fatalf("GET /topics/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /topics/%d: got %d, want %d", created.Id, resp.StatusCode, http.StatusOK)
	}
	var got models.Topic
	json.NewDecoder(resp.Body).Decode(&got)
	resp.Body.Close()

	if got.Id != created.Id || got.Name != "computers" {
		t.Errorf("GET /topics/%d: got %+v", created.Id, got)
	}

	// DELETE /topics/{id}
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/topics/%d", srv.URL, created.Id), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /topics/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE /topics/%d: got %d, want %d", created.Id, resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// GET after DELETE should 404
	resp, err = http.Get(fmt.Sprintf("%s/topics/%d", srv.URL, created.Id))
	if err != nil {
		t.Fatalf("GET /topics/%d after delete: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET /topics/%d after delete: got %d, want %d", created.Id, resp.StatusCode, http.StatusNotFound)
	}
	resp.Body.Close()
}

func TestSourcesCRUD(t *testing.T) {
	srv := httptest.NewServer(setupMux())
	defer srv.Close()

	body, _ := json.Marshal(models.Source{Name: "wiki", Type: "text", Raw: "some content"})
	resp, err := http.Post(srv.URL+"/sources", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /sources: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /sources: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created models.Source
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	if created.Id == 0 || created.Name != "wiki" {
		t.Errorf("POST /sources: got %+v", created)
	}

	// GET /sources/{id}
	resp, err = http.Get(fmt.Sprintf("%s/sources/%d", srv.URL, created.Id))
	if err != nil {
		t.Fatalf("GET /sources/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /sources/%d: got %d", created.Id, resp.StatusCode)
	}
	resp.Body.Close()

	// DELETE
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/sources/%d", srv.URL, created.Id), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /sources/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE /sources/%d: got %d", created.Id, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestStylesCRUD(t *testing.T) {
	srv := httptest.NewServer(setupMux())
	defer srv.Close()

	body, _ := json.Marshal(models.Style{Name: "formal", Prompt: "be formal"})
	resp, err := http.Post(srv.URL+"/styles", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /styles: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /styles: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created models.Style
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	if created.Id == 0 || created.Name != "formal" || created.Prompt != "be formal" {
		t.Errorf("POST /styles: got %+v", created)
	}

	// GET /styles/{id}
	resp, err = http.Get(fmt.Sprintf("%s/styles/%d", srv.URL, created.Id))
	if err != nil {
		t.Fatalf("GET /styles/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /styles/%d: got %d", created.Id, resp.StatusCode)
	}
	resp.Body.Close()

	// DELETE
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/styles/%d", srv.URL, created.Id), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /styles/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE /styles/%d: got %d", created.Id, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestDraftsCRUD(t *testing.T) {
	srv := httptest.NewServer(setupMux())
	defer srv.Close()

	// Create prerequisite topic and style first
	topicBody, _ := json.Marshal(models.Topic{Name: "tech"})
	resp, _ := http.Post(srv.URL+"/topics", "application/json", bytes.NewReader(topicBody))
	var topic models.Topic
	json.NewDecoder(resp.Body).Decode(&topic)
	resp.Body.Close()

	styleBody, _ := json.Marshal(models.Style{Name: "casual", Prompt: "be casual"})
	resp, _ = http.Post(srv.URL+"/styles", "application/json", bytes.NewReader(styleBody))
	var style models.Style
	json.NewDecoder(resp.Body).Decode(&style)
	resp.Body.Close()

	body, _ := json.Marshal(models.Draft{
		Title:   "draft1",
		Content: "hello world",
		TopicId: topic.Id,
		StyleId: style.Id,
		Status:  "pending",
	})
	resp, err := http.Post(srv.URL+"/drafts", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /drafts: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /drafts: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created models.Draft
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	if created.Id == 0 || created.Title != "draft1" || created.Status != "pending" {
		t.Errorf("POST /drafts: got %+v", created)
	}

	// GET /drafts/{id}
	resp, err = http.Get(fmt.Sprintf("%s/drafts/%d", srv.URL, created.Id))
	if err != nil {
		t.Fatalf("GET /drafts/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /drafts/%d: got %d", created.Id, resp.StatusCode)
	}
	resp.Body.Close()

	// DELETE draft first, then style and topic
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/drafts/%d", srv.URL, created.Id), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /drafts/%d: %v", created.Id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE /drafts/%d: got %d", created.Id, resp.StatusCode)
	}
	resp.Body.Close()
}
