# YAKIT LMS project — MySQL edition

Проект веб-приложения для онлайн-обучения: публичный сайт колледжа и LMS-система.

## Состав

```text
frontend/
  college-site.html   публичный сайт колледжа
  lms.html            LMS
backend/
  *.go                backend на Go/Gin/GORM
  mysql_init.sql      схема MySQL и демо-данные
scripts/
  mysql_init.sql      отдельный SQL-скрипт для создания базы
```


```bash
mysql -u root -p < scripts/mysql_init.sql
```

Скрипт создаёт базу:

```text
yakit_lms
```

и добавляет демо-аккаунты:

```text
student@yakit.ru  / 123456
student2@yakit.ru / 123456
teacher@yakit.ru  / 123456
teacher2@yakit.ru / 123456
admin@yakit.ru    / 123456
```

## Настройка backend

Файл `backend/.env.example` содержит пример DSN:

```text
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/yakit_lms?charset=utf8mb4&parseTime=True&loc=Local
```

Можно либо создать `.env`, либо задать переменную окружения вручную.

Windows PowerShell:

```powershell
$env:MYSQL_DSN="root:password@tcp(127.0.0.1:3306)/yakit_lms?charset=utf8mb4&parseTime=True&loc=Local"
```

CMD:

```cmd
set MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/yakit_lms?charset=utf8mb4&parseTime=True&loc=Local
```

Linux/macOS:

```bash
export MYSQL_DSN='root:password@tcp(127.0.0.1:3306)/yakit_lms?charset=utf8mb4&parseTime=True&loc=Local'
```

## Запуск

```bash
cd backend
go mod tidy
go run .
```

Открыть:

```text
http://localhost:8080/
http://localhost:8080/lms
```

## Frontend

Файл `frontend/college-site.html` взят из финальной версии сайта колледжа. Изменения минимальные:

- исправлены цвета input и placeholder в светлой теме;
- кнопка LMS в футере заменена на кнопку «Наверх».
