# Pull Request Service

Сервис для управления командами, пользователями и pull request'ами.

## Запуск приложения

### Предварительные требования
- Docker
- Docker Compose

### Установка и запуск

1. Создайте файл `.env` на основе примера:
```bash
cp .env.example .env
```
2. Запустите приложение
```bash
docker-compose up --build
```
3. Для остановки работы
```bash
docker-compose down -v
```