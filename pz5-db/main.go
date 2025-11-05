package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

)

func main() {
	

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// fallback — прямой DSN в коде (только для учебного стенда!)
		dsn = "postgres://postgres:YOUR_PASSWORD@localhost:5432/todo?sslmode=disable"
	}

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("openDB error: %v", err)
	}
	defer db.Close()

	repo := NewRepo(db)

	// 1) Вставим пару задач
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	titles := []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"}
	for _, title := range titles {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			log.Fatalf("CreateTask error: %v", err)
		}
		log.Printf("Inserted task id=%d (%s)", id, title)
	}

	// 2) Прочитаем список задач
	ctxList, cancelList := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelList()

	tasks, err := repo.ListTasks(ctxList)
	if err != nil {
		log.Fatalf("ListTasks error: %v", err)
	}

	// 3) Напечатаем
	fmt.Println("=== Tasks ===")
	for _, t := range tasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	tasksDone, _ := repo.ListDone(ctx, true)
fmt.Println("=== Done tasks ===", tasksDone)

	tasksUndone, _ := repo.ListDone(ctx, false)
fmt.Println("=== Undone tasks ===", tasksUndone)

	task, err := repo.FindByID(ctx, 1)
if err != nil {
    log.Println("FindByID error:", err)
} else {
    fmt.Printf("Task #%d: %s (done=%v)\n", task.ID, task.Title, task.Done)
}

titlesBatch := []string{"Задача 1", "Задача 2", "Задача 3"}
if err := repo.CreateMany(ctx, titlesBatch); err != nil {
    log.Fatalf("CreateMany error: %v", err)
}

}
