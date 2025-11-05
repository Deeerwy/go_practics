# Практическое занятие №5  Бурылин Дмитрий ПИМО-01-25
**Тема:** Подключение к PostgreSQL через `database/sql`. Выполнение простых запросов (INSERT, SELECT)  

## Цели работы
- Установить и настроить PostgreSQL локально.  
- Подключиться к БД из Go с помощью `database/sql` и драйвера PostgreSQL.  
- Выполнить параметризованные запросы `INSERT` и `SELECT`.  
- Научиться работать с `context`, пулом соединений и обработкой ошибок.  


## Окружение
- **Go:** 1.21  
- **PostgreSQL:** 14.19  


## Подготовка базы данных
```
sql
CREATE DATABASE todo;
\c todo

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tasks (title) VALUES ('Первая задача из psql');
SELECT * FROM tasks;
```

## Структура Проекта
```
pz5-db/
  ├── db.go          # подключение к БД и пул соединений
  ├── repository.go  # функции CreateTask, ListTasks, ListDone, FindByID, CreateMany
  ├── main.go        # запуск приложения
  ├── go.mod
```

## Скриншоты выполненных заданий:

После выполнения команды go run .:
![PIC1](image.png)

SELECT id, title, done, created_at FROM tasks ORDER BY id;
![PIC2](image-1.png)

Проверочное задание 1:
![PIC3](image-2.png)

Проверочное задание 2:
![PIC4](image-3.png)

Проверочное задание 3:
![PIC5](image-4.png)

##  Фрагменты кода:
```
// 1 Вставим пару задач
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

	// 2 Прочитаем список задач
	ctxList, cancelList := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelList()

	tasks, err := repo.ListTasks(ctxList)
	if err != nil {
		log.Fatalf("ListTasks error: %v", err)
	}

	// 3 Напечатаем
	fmt.Println("=== Tasks ===")
	for _, t := range tasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}
```

##  Краткие ответы

**Что такое пул соединений `*sql.DB` и зачем его настраивать?**  
`*sql.DB` — это менеджер пула соединений, а не одно соединение. Он управляет открытыми подключениями к базе данных, переиспользует их и ограничивает нагрузку. Настройка пула позволяет балансировать производительность и ресурсы приложения.

**Почему используем плейсхолдеры `$1`, `$2`?**  
Плейсхолдеры защищают от SQL‑инъекций и позволяют безопасно подставлять параметры в запросы. Это стандартный механизм параметризации в PostgreSQL.

**Чем отличаются `Query`, `QueryRow` и `Exec`?**  
- `Exec` — выполняет запросы без возврата строк (например, `INSERT`, `UPDATE`, `DELETE`).  
- `Query` — выполняет запрос, возвращающий несколько строк (`SELECT ...`).  
- `QueryRow` — выполняет запрос, возвращающий ровно одну строку.  

---

##  Обоснование транзакций и настроек пула

**Транзакции**  
Используются для атомарности: либо все операции выполняются, либо ни одна. Это особенно важно при массовых вставках (`CreateMany`) или связанных изменениях данных. Транзакции гарантируют целостность и согласованность базы.

**Настройки пула соединений**  
- `SetMaxOpenConns` — ограничивает общее число одновременных соединений, чтобы не перегрузить сервер.  
- `SetMaxIdleConns` — задаёт количество соединений, которые могут оставаться «в простое» для быстрого повторного использования.  
- `SetConnMaxLifetime` — ограничивает время жизни соединения, предотвращая утечки и зависания.  

Эти параметры позволяют оптимально использовать ресурсы и обеспечивают стабильную работу приложения при росте нагрузки.
