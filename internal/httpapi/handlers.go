package httpapi

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    
    "example.com/pz16-integration/internal/models"
    "example.com/pz16-integration/internal/service"
)

type Router struct{ Svc *service.Service }

func (rt Router) Register(r *gin.Engine) {
    r.POST("/notes", rt.createNote)
    r.GET("/notes/:id", rt.getNote)
    r.PUT("/notes/:id", rt.updateNote)
    r.DELETE("/notes/:id", rt.deleteNote)
    r.GET("/notes", rt.listNotes)
}

func (rt Router) createNote(c *gin.Context) {
    var in struct{ Title, Content string }
    if err := c.BindJSON(&in); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
        return
    }
    
    n := models.Note{Title: in.Title, Content: in.Content}
    if err := rt.Svc.Create(c, &n); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, n)
}

func (rt Router) getNote(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    n, err := rt.Svc.Get(c, id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    
    c.JSON(http.StatusOK, n)
}

func (rt Router) updateNote(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    var in struct{ Title, Content string }
    if err := c.BindJSON(&in); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
        return
    }
    
    if err := rt.Svc.Update(c, id, in.Title, in.Content); err != nil {
        // Предполагаем, что сервис возвращает "not found" для несуществующей записи
        if err.Error() == "not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "updated successfully"})
}

func (rt Router) deleteNote(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    if err := rt.Svc.Delete(c, id); err != nil {
        if err.Error() == "not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.Status(http.StatusNoContent)
}

func (rt Router) listNotes(c *gin.Context) {
    // Получаем параметры пагинации из query string
    limitStr := c.DefaultQuery("limit", "10")
    offsetStr := c.DefaultQuery("offset", "0")
    
    limit, err := strconv.ParseInt(limitStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
        return
    }
    
    offset, err := strconv.ParseInt(offsetStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
        return
    }
    
    // Ограничиваем максимальный лимит для защиты от перегрузки
    if limit > 100 {
        limit = 100
    }
    
    notes, err := rt.Svc.List(c, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, notes)
}