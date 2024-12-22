# EXAMPLE.md

# Примеры использования CalcAPI

Ниже приведены примеры запросов к нашему серверу. Перед этим убедитесь, что сервер запущен (см. инструкцию в [README.md](README.md)):

## :heavy_check_mark: Успешное вычисление
```bash
curl --location 'https://hseastro.space/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```
**Ожидаемый результат** (`200 OK`):
```json
{
  "result": 6
}
```

## :warning: Ошибка 422 (Unprocessable Entity)
Запрос с некорректным выражением (например, содержится буква `a`):

```bash
curl --location 'https://hseastro.space/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+a"
}'
```
**Ожидаемый результат** (`422 Unprocessable Entity`):
```json
{
  "error": "Expression is not valid"
}
```

## :x: Ошибка 500 (Internal Server Error)
Запрос, в котором произойдёт непредвиденная ошибка (например, деление на ноль):
```bash
curl --location 'https://hseastro.space/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "10/0"
}'
```
**Ожидаемый результат** (`500 Internal Server Error`):
```json
{
  "error": "Internal server error"
}
```

----

**Приятного использования!**