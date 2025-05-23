[![Pipeline](https://github.com/Nomaleus/djanGO/actions/workflows/go.yml/badge.svg)](https://github.com/Nomaleus/djanGO/actions/workflows/go.yml) 
# 📊 **API калькулятора**

Этот проект — это API для вычисления математических выражений. Поддерживает базовые операции, такие как сложение, вычитание, умножение, деление и использование скобок.

![GOЙДА](goyda.gif)

## Для связи ТГ - @nomaleus
#### До сих пор не оставлял его, но поскольку я посчитал, что установка проекта под Windows может вызвать у кого-нибудь абсолютно непредсказуемые затруднения(хоть я и расписал в README инструкцию как запустить библиотеки с линукса), по вопросам билда под эту платформу можете строчить гневные письма. Сам же я запустил на этой системе, и описал проделанные мной пункты, поэтому проект должен запуститься и у вас :)

---

## 🏗 Архитектура
Сей проект работает по следующей методике:

![Architecture](Scheme.png)

---

## 📥 Установка
## **(MAC/LINUX)**:
### 1. Установите [Go](https://go.dev/doc/install)

### 2. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/nomaleus/djanGO.git
   cd djanGO
   ```

### 3. Установите зависимости:
   ```bash
   go mod tidy
   ```

### 4. Запустите cервер
   ```bash
   cd cmd
   go run main.go
   ```

Сервер запустится по адресу: [http://localhost:80](http://localhost:80).
Если сервер не запустился, то **повторите** данную инструкцию сначала, или освободите порт (:
---
# Windows
## Дисклеймер перед запуском на WINDOWS, приятного прочтения! 
#### В принципе, можно не выполнять все описанные ниже действия, проект покрыт модульными и интеграционными тестами, имеет интеграцию с базой данных sqlite3, имеется grpc, веб-интерфейс и всё такое. О том, что он билдится и проходят все возможные и невозможные тесты можно догадаться, взглянув на бейдж пайплайна, или подробно на коммиты, но я бы рекомендовал посмотреть на веб хотя бы) 
Если у вас не Linux, и даже не OSX, то придется ~~поплакать~~, потому что одна из библиотек проекта использует инструменты GNU, которые изначально не могут быть в системе Windows. Методом проб, гнева и ошибок составил максимально простую инструкцию как можно запустить проект под Windows:

### 1. Установите [Go](https://go.dev/doc/install)
### 2. Устанавливаем Chocolatey
Для PowerShell:
```bash
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```
Для CMD:
```bash
@powershell -NoProfile -ExecutionPolicy Bypass -Command "iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
```
После установки проверьте, что оно работает командой choco --version.

### 3. Устанавливаем mingw
Откройте CMD или PowerShell от имени администратора, и введите команду:
```bash
choco install mingw
```
Затем следуйте инструкции. В конце введите "gcc --version", и убедитесь, что установка прошла успешно.

### 4. Клонируем репозиторий
Перейдите в любую папку и склонируйте репозиторий 
```bash
git clone https://github.com/nomaleus/djanGO.git
cd djanGO
```
### 5. Соберите зависимости
```bash
go mod tidy
```
### 6. Соберите и запустите сервер
```bash
cd cmd
go run main.go
```

   Сервер запустится по адресу: [http://localhost:80](http://localhost:80). 
   Если сервер не запустился, то **повторите** данную инструкцию сначала, или освободите порт (: 

---
# Переменные окружения(dotENV)
Проект поддерживает загрузку переменных окружения.

Для LINUX/OSX создайте файл .env и выставьте необходимые настройки по образу и подобию файла .env.example

Для Windows повторите тоже самое, только положите этот файл в ту же папку, откуда и запускаете проект(по инструкции в папку cmd, и тогда сервер найдет его).

**Если вы включили GRPC, не забудьте перед отправкой выражения запустить воркера из директории "cmd/grpc_worker"**:
```bash
go run main.go
```
Если не хотите лицезреть спам запросами от клиента, который не понимает, почему его задачу никто не берёт.

## 🤓 WebUI
### В проекте реализован веб-интерфейс для демонстрации возможностей API, поэтому использование нижеописанными способами необязательно.
Чтобы перейти в веб-интерфейс, необходимо всего лишь перейти по адресу, на котором стартовал сервер, по умолчанию это - localhost:80 :)

~~Не~~стандартный дисклеймер Павла: Веб-интерфейс может быть требователен к железу и предоставлен лишь для демонстрации возможностей API, но никак **не может** являться решением для **PROD**.

~~Не не~~стандартный дисклеймер: YOUR_ACCESS_TOKEN заменяйте на полученный любым способом токен авторизации. И да, это не гпт писал...) 

---

## 🚀 Использование
#### Для отправки каждого из предоставленных запросов, необходимо установить в заголовке Authorization действующий JWT Bearer, который можно получить через инструменты разработчика в веб-интерфейсе.

### **POST /api/v1/calculate**

Отправьте POST-запрос с выражением в формате JSON для вычисления.
#### Пример запроса:
```json
{
  "expression": "3 + (2 * 5) / (7 - 4)"
}
```

#### CURL:
```bash
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
--data '{
  "expression": "2+2*2"
}'
```

#### Пример успешного ответа:
```json
{
   "id": "ca79f6ef-9861-44ac-bd7f-0702e89559e7"
}
```

### **GET /api/v1/expressions**

Отправьте GET-запрос для получения списка выражений.
#### CURL:
```bash
curl --location 'localhost/api/v1/expressions' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
```
#### Пример успешного ответа:
```json
{
  "expressions": [
    {
      "id": "dc953f2c-e63e-4546-ae7d-db6a6c23bec7",
      "status": "COMPLETED",
      "result": 82
    },
    {
      "id": "57147543-87b2-4b87-8896-77407b4cb914",
      "status": "COMPLETED",
      "result": 84
    }
  ]
}
```

###  GET **localhost/api/v1/expressions/id**

Отправьте GET запрос, чтобы узнать информацию о конкретном выражении по его уникальному идентификатору(id).
#### CURL:
```bash
curl --location 'localhost/api/v1/expressions/id' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer YOUR_ACCESS_TOKEN' \

```
Данная команда выведет подробную информацию о том как считалось отправленное вами выражение(это действительно интересно, пусть и ужасно). 
```json
{
   "expression": {
      "id": "dc953f2c-e63e-4546-ae7d-db6a6c23bec7",
      "status": "COMPLETED",
      "result": 82
   }
}
```

###  POST **localhost/internal/task**
Сервер не позволяет пользователю извне вмешиваться в задачи, но существует подобный путь для POST-запросов.

#### Пример запроса:
```json
{
   "task": {
      "id": "4",
      "arg1": "2",
      "arg2": "3",
      "operation": "*",
      "operation_time": 5000
   }
}
```
#### Пример ответа:
```json
{
    "id": "4",
    "result": 6
}
```

###  POST localhost/api/v1/register
Есть возможность зарегистрироваться, вот пример запроса:
```bash
{"login":"admin","password":"admin123"}
```
#### CURL:
```bash
curl --location 'localhost/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login":"testuser","password":"testuser"
}'
```
Пример ответа:
```bash
{
    "message": "Пользователь успешно зарегистрирован",
    "success": "true"
}
```

###  POST **localhost/api/v1/login**
Можно и авторизоваться на сервере. Пример запроса:
```bash
{"login":"admin","password":"admin123"}
```
#### CURL:
```bash
curl --location 'localhost/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login":"testuser","password":"testuser"
}'
```
Пример ответа:
```bash
{
    "success": true,
    "login": "admin",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFkbWluIiwiZXhwIjoxNzQ2NDg4MDI0LCJuYmYiOjE3NDY0MDE2MjQsImlhdCI6MTc0NjQwMTYyNH0.iRcUvn0X2tylQCPM9JYmTJYoojqGSnmihmgeV9nzBXA"
}
```
Токен можете использовать для дальнейшей авторизации в заголовках.

---

### Команды для использования:

#### Пример использования с `curl`:
```bash
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
--data '{
  "expression": "2+2*2"
}'
```

YOUR_ACCESS_TOKEN замените на уже имеющийся JWT(Получите его авторизовавшись через веб, либо любым другим удобным способом). Если всё сделаете правильно, то можно увидеть стандартный ответ вроде:
```bash
{"id":"17464008-0300-8932-0005-853018633061"}
```
#### Остальные роуты запросов для curl можно взять из инструкции выше

---

## ⚠️ Коды HTTP

### В проекте существует несколько HTTP кодов ответа сервера:

```css
200 - успешно
201 - выражение принято для вычисления, либо же создано
304 - сообщает о том, что нет необходимости передавать ресурсы
401 - пользователь не авторизован
404 - что-то не найдено :)
405 - нет доступа к методу
409 - конфликт с имеющимися данными на сервере
500 - что-то пошло не так
```
---

## 🛠 Тестирование

В проекте есть несколько тестов. Чтобы запустить тесты, выполните:

```bash
go test ./...
```

Если все прошло успешно, вы увидите вывод вроде:
```
ok      djanGO/handlers         1.276s
ok      djanGO/handlers/tests   3.428s
ok      djanGO/integration      0.540s
ok      djanGO/storage          0.903s
ok      djanGO/utils            0.707s

```

---

## gRPC интеграция

Как можно было догадаться, проект обладает gRPC клиентом и сервером :)

Ниже описанная инструкция скорее для разработчиков, а также она имеет место быть только на системах **Linux/OSX**, поскольку я использую инструмент make, но для Windows-юзеров в папке scripts оставлены скрипты cmd и Powershell(в зависимости от того, что нравится) для сброки grpc сервера и воркера. 

### Установка и настройка

1. Для генерации Go-кода из proto-файлов вам потребуется установить `protoc` и плагины для Go:

Установка protoc (пример для macOS с Homebrew)
```bash
brew install protobuf
```

Установка плагинов Go для protoc
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

2. Генерация Go-кода из proto-файлов:

```bash
make proto
```

### Сборка и запуск

1. Сборка сервера и gRPC воркера:

```bash
make build
```

2. Запуск сервера с включенным gRPC:

```bash
make server
```

3. Запуск gRPC воркера в отдельном терминале:

```bash
make worker
```

4. Альтернативно, запуск и сервера, и воркера одновременно:

```bash
make run
```

### Переменные окружения

Чуть подробнее о переменных окружения. Там, откуда вы запускаете сервер, может находиться файл dotENV(.env), в котором могут быть установлены настройки, отличающиеся от стандартных на сервере
```
# Порт для HTTP сервера
PORT=8080

# Включить gRPC сервер (true/false)
ENABLE_GRPC=true

# Адрес и порт для gRPC сервера
GRPC_ADDR=:50051

# Количество gRPC воркеров
WORKERS_COUNT=2

# Вычислительная мощность сервера (количество локальных воркеров)
COMPUTING_POWER=2
```

В репозитории есть пример файла `.env.example` для справки.
