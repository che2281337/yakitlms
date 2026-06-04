# Backend YAKIT LMS — MySQL only

Backend работает только с MySQL. Автоматические миграции удалены.

Перед запуском нужно создать базу и таблицы:

```bash
mysql -u root -p < mysql_init.sql
```

Затем задать строку подключения:

```text
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/yakit_lms?charset=utf8mb4&parseTime=True&loc=Local
```

Запуск:

```bash
go mod tidy
go run .
```
