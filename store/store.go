package store

type Task struct {
    ID          string
    Title       string
    Description *string
    Done        bool
}

// Глобальное in-memory хранилище с предзаполненными данными
var Tasks = []*Task{
    {ID: "t_001", Title: "Первая задача", Description: strPtr("Учебный пример"), Done: false},
    {ID: "t_002", Title: "Вторая задача", Description: strPtr("GraphQL API"), Done: true},
}

func strPtr(s string) *string {
    return &s
}
