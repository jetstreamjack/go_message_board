# Message Board

HTTP-сервер — доска сообщений на Go 1.25. Все шесть этапов реализованы, `go test ./tests/ -v` зелёный.

## Запуск

```powershell
go run ./cmd/server              # порт из PORT (по умолчанию 8080)
curl localhost:8080/health       # → ok
```

## API

| Метод | Путь | Тело | Ответ |
|---|---|---|---|
| GET | `/health` | — | 200 `ok` |
| POST | `/echo` | любой текст | 200 тело без изменений |
| POST | `/echo` | `{"message":"..."}` + JSON content-type | 200 `{"message":"..."}` |
| POST | `/messages` | `{"message":"..."}` | 201 `{"id":1,"message":"...","created_at":"..."}` |
| GET | `/messages` | — | 200 `[...]` новые сверху |
| DELETE | `/messages/{id}` | — | 204 / 404 |

Bad input (сломанный JSON, пустое сообщение) → 400.

## Структура

```
cmd/server/main.go      entry point, routing, frontend static files
internal/handler/       HTTP handlers (slog logging)
internal/store/         in-memory store
internal/helper/        body reading, JSON decoding
internal/response/      JSON response helpers
tests/                  unit tests for handler & store + integration tests
frontend/               embedded web panel
```

## Тесты

```bash
go test ./tests/ -v                          # e2e (поднимают сервер сами)
go test ./tests/handler/ -v                  # unit, handlers
go test ./tests/store/ -v                    # unit, store
go test ./tests/ -v -count=1                 # сбросить кэш go test
```
