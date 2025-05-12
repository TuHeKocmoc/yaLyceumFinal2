# EXAMPLE.md

# Примеры использования "CalcAPI"

---

## :green_heart: 0. Регистрация и логин


**0.1 Регистрация**:
```bash
curl --location 'http://localhost:8080/api/v1/register' \
  --header 'Content-Type: application/json' \
  --data '{
    "login": "alice",
    "password": "secret123"
  }'
```
**Ожидаемый ответ (код 200 OK)**:
```json
{"status":"ok"}
```
**0.2 Логин**:
```bash
curl --location 'http://localhost:8080/api/v1/login' \
  --header 'Content-Type: application/json' \
  --data '{
    "login": "alice",
    "password": "secret123"
  }'
```
**Ожидаемый ответ (код 200 OK)**:
```json
{
  "token": "<jwt_token_string>"
}
```

## :heavy_check_mark: 1. Успешное добавление выражения

**Запрос**:  
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
  --header 'Content-Type: application/json' \
  --header "Authorization: Bearer $TOKEN" \
  --data '{
    "expression": "2+2*2"
  }'
```
**Ожидаемый ответ (код 201 Created)**:
```json
{
  "id": "e4a95b12-8c2f-49d5-8914-8eac522c8512"
}
```
где `id` — уникальный идентификатор созданного выражения.

## :mag: 2. Проверка статуса выражения

После добавления выражения, можно периодически смотреть, посчитан ли результат.

**Получение списка выражений**:
```bash
curl -X GET 'http://localhost:8080/api/v1/expressions' \
  --header "Authorization: Bearer $TOKEN"
```

**Пример ответа (код 200 OK)**:
```
json
{
  "expressions": [
    {
      "id": "e4a95b12-8c2f-49d5-8914-8eac522c8512",
      "status": "IN_PROGRESS",
      "result": null
    }
  ]
}
```
Пока агент не вычислил значение, `status` может быть `IN_PROGRESS`, а `result` — `null`.
**Получение одного выражения по ID**:
```bash
curl -X GET "http://localhost:8080/api/v1/expressions/e4a95b12-8c2f-49d5-8914-8eac522c8512" \
  --header "Authorization: Bearer $TOKEN"
```
Если агент уже выполнил задачу, ответ будет:
```json
{
  "expression": {
    "id": "e4a95b12-8c2f-49d5-8914-8eac522c8512",
    "status": "DONE",
    "result": 6
  }
}
```

## :warning: 3. Ошибка 422 (Unprocessable Entity)

Если в выражении содержатся некорректные символы (например, буква a), оркестратор вернёт `422 Unprocessable Entity`.

**Запрос**:
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $TOKEN" \
  --data '{
    "expression": "2+a"
  }'
}'
```

**Ожидаемый ответ (код 422)**:
```json
{
  "error": "Expression is not valid"
}
```

## :x: 4. Ошибка 500 (Internal Server Error)

Если при вычислении произойдёт критическая ошибка (например, деление на ноль), то, после того как агент попытается вычислить задачу, оркестратор может установить выражению статус ошибки и вернуть при последующих запросах `result = null` и `status = "ERROR"`.

Например, если `calc.Calc("10/0")` выбрасывает ошибку `«division by zero»`, агент передаст обратно ошибку (если такая логика реализована). В итоге оркестратор может ответить:

```bash
curl --location "http://localhost:8080/api/v1/expressions/<id>" \
  --header "Authorization: Bearer $TOKEN"
```

```json
{
  "expression": {
    "id": "<id>",
    "status": "ERROR",
    "result": null
  }
}```