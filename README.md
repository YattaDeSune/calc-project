## About 👀
Веб-сервис для вычисления арифметических выражений. Поддерживает операции сложения, вычитания, умножения и деления, а также операции приоретизации. В случае некорректного ввода выдает конкретную ошибку.

**Структура проекта:**
```
сmd
  - main.go               // запуск приложения
internal
  - application
    -- application.go      // логика веб-сервера
    -- application_test.go // тесты для хэндлера
pkg
  - calculation
    -- calculation.go      // логика вычислений
    -- calculation_test.go // тесты для вычислений
    -- errors.go           // ошибки ввода
```

## How it works 🎯

Сервер по умолчанию запускается на порту `:8080`. У сервиса 1 endpoint с url-ом /api/v1/calculate. Пользователь отправляет на этот url POST-запрос с телом:
```json
{
    "expression": "введенное выражение"
}
```
В ответ пользователь получает HTTP-ответ с телом:
```json
{
    "result": "результат выражения"
}
```
и кодом <img src="https://img.shields.io/badge/status-200-brightgreen" alt="Status: 200">, если выражение вычислено успешно.

Eсли входные данные не соответствуют требованиям приложения (например введен невалидный символ или присутствует лишняя скобка) - пользователь получает HTTP-ответ с телом:
```json
{
    "error": "Ошибка, зависящая от типа некорректно введенных данных"
}
```
и кодом <img src="https://img.shields.io/badge/status-422-red" alt="Status: 422">.

Ещё один вариант HTTP-ответа:
```json
{
    "error": "internal server error"
}
```
и код <img src="https://img.shields.io/badge/status-500-red" alt="Status: 500"> в случае какой-либо иной ошибки.

В проекте реализовано простое **логгирование** с помощью стандартного пакета `log`, охватывающее запуск сервера, а также отлавливающее ошибки со статус кодом <img src="https://img.shields.io/badge/status-500-red" alt="Status: 500">

## Quick start ⚡
После клонирования проекта достаточно ввести в терминал команду:
```
go run cmd/main.go
```
После чего сервер запустится на порту `:8080`.

Далее, посылая запросы на адрес `localhost:8080/api/v1/calculate` (с помощью `curl`, `Postman` или любым другим способом) вы будете получать определенный ответ в формате JSON.

## Examples
**1. Успешный ответ**

Отправим JSON с помощью `curl`:
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```

Получим ответ:
```json
  {"result":6}
```
При проверке через `Postman` можно убедиться, что статус код - <img src="https://img.shields.io/badge/status-200-brightgreen" alt="Status: 200">

**2. Неверный ввод**

Отправим JSON с неверным выражением (остутствует открывающая скобка) с помощью `curl`:
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2)"
}'
```

Получим ответ:
```json
  {"error":"no opening parenthesis"}
```
При проверке через `Postman` можно убедиться, что статус код - <img src="https://img.shields.io/badge/status-422-red" alt="Status: 422">

**3. Иная ошибка**

Отправим пустые данные с помощью `curl`:
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data ''
```

Получим ответ:
```json
  {"error":"internal server error"}
```
При проверке через `Postman` можно убедиться, что статус код - <img src="https://img.shields.io/badge/status-500-red" alt="Status: 500">

## Contacts 💬
<div id="contacts">
  <a href="https://t.me/YattaDesuNe">
    <img src="https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white" alt="Telegram Badge"/>
  </a>
  <a href="mailto:belyaevlv742@gmail.com">
    <img src="https://img.shields.io/badge/Gmail-D14836?style=for-the-badge&logo=gmail&logoColor=white" alt="Email Badge"/>
  </a>
</div>
<img src="https://komarev.com/ghpvc/?username=YattaDeSune&style=flat-square&color=blue" alt=""/>
