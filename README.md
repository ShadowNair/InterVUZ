Для запуска с загрузкой данных в БД и обновления каждые 72 часа:
```
SYNC_ON_STARTUP=true SYNC_INTERVAL=72h docker compose up --build
```

Если БД полная и для ускорения запуска(если расписание не нужно)
```
docker compose up --build
```