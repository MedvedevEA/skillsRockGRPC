# skillsRockGRPC
## Skills Rock auth service
Базовый gRPC сервис авторизации и аунтификации
## Запуск проекта
1. Клонируйте репозиторий
2. Создайте секретный ключ для JWT токена и поместите его в папку secret/. Публичный ключ вам не понадобится
```
go run cmd/createSecret/main.go -config config/local.yml
```
3. Выполните миграцию БД 
```
go run cmd/migrator/main.go -source-path migration/ -database-url "postgres://postgres:1234@localhost:5433/temp?sslmode=disable"
```
4. Создайте файл конфигурации config/local.yml по образцу

5. Запустите проект
```
go run cmd/todo/main.go -config config/local.yml
```