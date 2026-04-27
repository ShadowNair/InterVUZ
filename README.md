# InterVUZ Backend

## Run with Docker

1. Create local env file:

```bash
cp .env.example .env
```

On Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

2. Fill required values in `.env`:
- `ASSISTANT_API_KEY`
- `ACADEMIC_WEEK1_START_DATE` (format `YYYY-MM-DD`, Monday of week 1)

3. Start:

```bash
docker compose up --build
```

Optional startup sync every 72h:

```bash
SYNC_ON_STARTUP=true SYNC_INTERVAL=72h docker compose up --build
```
