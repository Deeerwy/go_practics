package integration

import (
    "database/sql"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"
    "fmt"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"

    "example.com/pz16-integration/internal/db"
    "example.com/pz16-integration/internal/httpapi"
    "example.com/pz16-integration/internal/repo"
    "example.com/pz16-integration/internal/service"
)

func newServer(t *testing.T, dsn string) *httptest.Server {
    t.Helper()
    dbx, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatal(err)
    }
    db.MustApplyMigrations(dbx)

    r := gin.Default()
    svc := service.Service{Notes: repo.NoteRepo{DB: dbx}}
    httpapi.Router{Svc: &svc}.Register(r)

    return httptest.NewServer(r)
}

func TestCreateAndGetNote(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    // 1) Create
    resp, err := http.Post(srv.URL+"/notes", "application/json",
        strings.NewReader(`{"title":"Hello","content":"World"}`))
    if err != nil {
        t.Fatal(err)
    }
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("status %d != 201", resp.StatusCode)
    }
    body, _ := io.ReadAll(resp.Body)
    _ = resp.Body.Close()
    var created map[string]any
    _ = json.Unmarshal(body, &created)
    id := int64(created["id"].(float64))

    // 2) Get
    resp2, err := http.Get(fmt.Sprintf("%s/notes/%d", srv.URL, id))
    if err != nil {
        t.Fatal(err)
    }
    if resp2.StatusCode != http.StatusOK {
        t.Fatalf("status %d != 200", resp2.StatusCode)
    }
    _ = resp2.Body.Close()
}

func TestGetNonExistentNote(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    // GET несуществующей заметки
    resp, err := http.Get(fmt.Sprintf("%s/notes/%d", srv.URL, 999999))
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusNotFound {
        t.Fatalf("expected status 404, got %d", resp.StatusCode)
    }
}

func TestUpdateNote(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    // 1) Create
    resp, err := http.Post(srv.URL+"/notes", "application/json",
        strings.NewReader(`{"title":"Original","content":"Content"}`))
    if err != nil {
        t.Fatal(err)
    }
    body, _ := io.ReadAll(resp.Body)
    _ = resp.Body.Close()
    var created map[string]any
    _ = json.Unmarshal(body, &created)
    id := int64(created["id"].(float64))

    // 2) Update using PUT
    client := &http.Client{}
    req, err := http.NewRequest("PUT", fmt.Sprintf("%s/notes/%d", srv.URL, id),
        strings.NewReader(`{"title":"Updated","content":"New content"}`))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp2, err := client.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp2.Body.Close()
    
    if resp2.StatusCode != http.StatusOK {
        t.Fatalf("expected status 200 for update, got %d", resp2.StatusCode)
    }

    // 3) Verify update
    resp3, err := http.Get(fmt.Sprintf("%s/notes/%d", srv.URL, id))
    if err != nil {
        t.Fatal(err)
    }
    defer resp3.Body.Close()
    
    body3, _ := io.ReadAll(resp3.Body)
    var note map[string]any
    _ = json.Unmarshal(body3, &note)
    
    if note["title"] != "Updated" {
        t.Errorf("expected title 'Updated', got %v", note["title"])
    }
    if note["content"] != "New content" {
        t.Errorf("expected content 'New content', got %v", note["content"])
    }
}

func TestDeleteNote(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    // 1) Create
    resp, err := http.Post(srv.URL+"/notes", "application/json",
        strings.NewReader(`{"title":"To Delete","content":"Content"}`))
    if err != nil {
        t.Fatal(err)
    }
    body, _ := io.ReadAll(resp.Body)
    _ = resp.Body.Close()
    var created map[string]any
    _ = json.Unmarshal(body, &created)
    id := int64(created["id"].(float64))

    // 2) Delete
    client := &http.Client{}
    req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notes/%d", srv.URL, id), nil)
    if err != nil {
        t.Fatal(err)
    }
    
    resp2, err := client.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp2.Body.Close()
    
    if resp2.StatusCode != http.StatusNoContent && resp2.StatusCode != http.StatusOK {
        t.Fatalf("expected status 204 or 200 for delete, got %d", resp2.StatusCode)
    }

    // 3) Verify deletion
    resp3, err := http.Get(fmt.Sprintf("%s/notes/%d", srv.URL, id))
    if err != nil {
        t.Fatal(err)
    }
    defer resp3.Body.Close()
    
    if resp3.StatusCode != http.StatusNotFound {
        t.Fatalf("expected status 404 after deletion, got %d", resp3.StatusCode)
    }
}

func TestListNotesWithPagination(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    // Create multiple notes
    for i := 1; i <= 5; i++ {
        resp, err := http.Post(srv.URL+"/notes", "application/json",
            strings.NewReader(fmt.Sprintf(`{"title":"Note %d","content":"Content %d"}`, i, i)))
        if err != nil {
            t.Fatal(err)
        }
        _ = resp.Body.Close()
    }

    // Test with pagination parameters
    resp, err := http.Get(srv.URL + "/notes?limit=3&offset=1")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected status 200 for list, got %d", resp.StatusCode)
    }
    
    body, _ := io.ReadAll(resp.Body)
    var notes []map[string]any
    _ = json.Unmarshal(body, &notes)
    
    if len(notes) > 3 {
        t.Errorf("expected at most 3 notes with limit=3, got %d", len(notes))
    }
}

func TestListAllNotes(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    // Create multiple notes
    for i := 1; i <= 3; i++ {
        resp, err := http.Post(srv.URL+"/notes", "application/json",
            strings.NewReader(fmt.Sprintf(`{"title":"Test Note %d","content":"Test Content %d"}`, i, i)))
        if err != nil {
            t.Fatal(err)
        }
        _ = resp.Body.Close()
    }

    // Get all notes
    resp, err := http.Get(srv.URL + "/notes")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected status 200, got %d", resp.StatusCode)
    }
    
    body, _ := io.ReadAll(resp.Body)
    var notes []map[string]any
    _ = json.Unmarshal(body, &notes)
    
    if len(notes) < 3 {
        t.Errorf("expected at least 3 notes, got %d", len(notes))
    }
    
    // Verify each note has required fields
    for _, note := range notes {
        if note["id"] == nil {
            t.Error("note missing id field")
        }
        if note["title"] == nil {
            t.Error("note missing title field")
        }
        if note["content"] == nil {
            t.Error("note missing content field")
        }
    }
}

func TestUpdateNonExistentNote(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    client := &http.Client{}
    req, err := http.NewRequest("PUT", fmt.Sprintf("%s/notes/%d", srv.URL, 999999),
        strings.NewReader(`{"title":"Updated","content":"New content"}`))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusNotFound {
        t.Fatalf("expected status 404 for non-existent update, got %d", resp.StatusCode)
    }
}

func TestDeleteNonExistentNote(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    client := &http.Client{}
    req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notes/%d", srv.URL, 999999), nil)
    if err != nil {
        t.Fatal(err)
    }
    
    resp, err := client.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusNotFound {
        t.Fatalf("expected status 404 for non-existent delete, got %d", resp.StatusCode)
    }
}

func TestCreateNoteInvalidJSON(t *testing.T) {
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        t.Skip("DB_DSN not set (use `make up` and `make test`)")
    }
    srv := newServer(t, dsn)
    defer srv.Close()

    resp, err := http.Post(srv.URL+"/notes", "application/json",
        strings.NewReader(`{"title":"Invalid JSON`))
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusBadRequest {
        t.Fatalf("expected status 400 for invalid JSON, got %d", resp.StatusCode)
    }
}
