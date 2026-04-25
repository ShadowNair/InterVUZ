Для запуска с загрузкой данных в БД и обновления каждые 72 часа:
```
ACADEMIC_WEEK1_START_DATE=2026-02-09 SYNC_ON_STARTUP=true SYNC_INTERVAL=72h docker compose up --build
```

Если БД полная и для ускорения запуска(если расписание не нужно)
```
ACADEMIC_WEEK1_START_DATE=2026-02-09 docker compose up --build
```

`ACADEMIC_WEEK1_START_DATE` — понедельник первой учебной недели. По нему backend
определяет числитель (`ch`) и знаменатель (`zn`) для расписания аудиторий.
