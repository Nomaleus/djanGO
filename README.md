[![Pipeline](https://github.com/Nomaleus/djanGO/actions/workflows/go.yml/badge.svg)](https://github.com/Nomaleus/djanGO/actions/workflows/go.yml) 
# 📊 **API калькулятора**

Этот проект — это API для вычисления математических выражений. Поддерживает базовые операции, такие как сложение, вычитание, умножение, деление и использование скобок.

---

## 📥 Установка
1. Установите [Go](https://go.dev/doc/install):

2. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/nomaleus/djanGO.git
   cd djanGO
   ```

3. Установите зависимости:
   ```bash
   go mod tidy
   ```

4. Запустите сервер:
   ```bash
   go run main.go
   ```

   Сервер запустится по адресу: [http://localhost:80](http://localhost:80).

---

## 🚀 Использование

### **POST /api/v1/calculate**

Отправьте POST-запрос с выражением в формате JSON для вычисления.

#### Пример запроса:
```json
{
  "expression": "3 + (2 * 5) / (7 - 4)"
}
```

#### Пример успешного ответа:
```json
{
  "result": "6.333333"
}
```

#### Пример ошибки:
```json
{
  "error": "Не дели на ноль!"
}
```

---

### Команды для использования:

#### Пример использования с `curl`:
```bash
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```

---

## 🛠 Тестирование

В проекте есть 30 тестов в файле `main_test.go`. Чтобы запустить тесты, выполните:

```bash
go test
```

Если все прошло успешно, вы увидите вывод вроде:
```
ok  	djanGO	0.671s
```

---