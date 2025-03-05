# EXAMPLE.md

# Примеры использования "CalcAPI"

---

## :heavy_check_mark: 1. Успешное добавление выражения

**Запрос**:  
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
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
curl -X GET 'http://localhost:8080/api/v1/expressions'
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
curl -X GET 'http://localhost:8080/api/v1/expressions/e4a95b12-8c2f-49d5-8914-8eac522c8512'
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
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+a"
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
curl --location 'http://localhost:8080/api/v1/expressions/<id>'
```

```json
{
  "expression": {
    "id": "<id>",
    "status": "ERROR",
    "result": null
  }
}```