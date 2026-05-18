---
# Практическое задание 21

## ЭФМО-02-25 

## Алиев Каяхан Командар оглы
---
# Тема работы
HTTPS/TLS и защита от SQL-инъекций в серверном приложении.

## Цели занятия
Научиться включать защищённый транспорт (HTTPS) и устранять уязвимости SQL-инъекций, используя корректные методы работы с базой данных.

## Структура проекта
```
C:.
│   .gitattributes
│   .gitignore
│   go.mod
│   go.sum
│   README.md
│   testdata.bat
│
├───.vs
│   │   ProjectSettings.json
│   │   slnx.sqlite
│   │   VSWorkspaceState.json
│   │
│   └───tech-ip-sem2
│       ├───FileContentIndex
│       │       2019e4a7-05d4-4380-9757-192646eef486.vsidx
│       │
│       └───v17
├───deploy
│   ├───monitoring
│   │       docker-compose.yml
│   │       prometheus.yml
│   │
│   └───tls
│           cert.pem
│           docker-compose.yml
│           key.pem
│           nginx.conf
│
├───docs
│       pz17_api.md
│       pz17_diagram.md
│
├───img
├───proto
│       auth.proto
│
├───services
│   ├───auth
│   │   ├───cmd
│   │   │   └───auth
│   │   │           main.go
│   │   │
│   │   ├───internal
│   │   │   ├───config
│   │   │   ├───grpc
│   │   │   │       server.go
│   │   │   │
│   │   │   ├───handler
│   │   │   │       auth.go
│   │   │   │
│   │   │   ├───http
│   │   │   └───service
│   │   └───pkg
│   │       └───authpb
│   │               auth.pb.go
│   │               auth_grpc.pb.go
│   │
│   └───tasks
│       │   Dockerfile
│       │
│       ├───cmd
│       │   └───tasks
│       │           main.go
│       │
│       └───internal
│           ├───client
│           │   ├───authclient
│           │   │       client.go
│           │   │
│           │   └───authgrpc
│           │           client.go
│           │
│           ├───handler
│           │       tasks.go
│           │
│           ├───http
│           ├───middleware
│           │       auth.go
│           │       metric.go
│           │
│           ├───migration
│           │       001_create_tasks.sql
│           │
│           ├───models
│           │       tasks.go
│           │
│           ├───repository
│           │       postgres.go
│           │
│           ├───service
│           └───storage
│                   memory.go
│
└───shared
    ├───httpx
    │       client.go
    │
    ├───logger
    │       logger.go
    │
    └───middleware
            accesslog.go
            requestid.go
```

## Коды статуса:
-	200 OK — успешный ответ
-	201 Created — ресурс создан
-	204 No Content — успешно, без тела
-	400 Bad Request — неверные данные
-	404 Not Found — ресурс не найден
-	422 Unprocessable Entity — некорректные данные по смыслу
-	500 Internal Server Error — внутренняя ошибка

# Примечания по конфигурации и требования

Для запуска требуется:

Go: версия 1.25.1

<img width="841" height="232" alt="Установка Git и Go" src="https://github.com/user-attachments/assets/8e01d831-5a7f-4376-8348-9052b240aec9" />


# Команды запуска/сборки
## 1) Клонировать данный репозиторий в удобную для вас папку:
```Powershell
git clone https://github.com/kayahan81/pz21
```
## 2) Перейти в папку pz19:
```Powershell
cd pz21
```
## 3) Загрузка зависимостей:
```Powershell
go mod tidy
```
## 4) Команда запуска
В первом окне
```Powershell
cd services/auth
$env:AUTH_PORT="8081"
$env:AUTH_GRPC_PORT="50051"
$env:ENV="development"
go run ./cmd/auth
```
Во втором окне
```Powershell
cd deploy/tls
docker-compose up -d
```

# Проверка работоспособности
В качестве СУБД была выбрана PostgreSQL
Сервис обработки задач разворачивается на docker
## Создаём и просматриваем задачи
<img width="754" height="754" alt="image" src="https://github.com/user-attachments/assets/956aca19-6ce6-4e1c-a206-ba979a6dc794" />

## Нормальный поиск по названию
<img width="704" height="639" alt="image" src="https://github.com/user-attachments/assets/35636f51-dd20-4462-b33d-dece2f8917e3" />

## Демонстрация уязвимости на учебном стенде
Видно, что по запросу выдаются все данные. Это плохо
<img width="714" height="912" alt="image" src="https://github.com/user-attachments/assets/5e9406b2-e03e-4133-85f9-2831eac5de50" />  

Если же попробовать тот же запрос в нормальном поиске - результат будет null
<img width="703" height="537" alt="image" src="https://github.com/user-attachments/assets/1e720a71-3a68-41a2-a6ca-dddc0650121b" />

# Отчёт
- Был выбран NGINX как TLS-терминатор, потому что отделяет шифрование от логики приложения
- Команды генерации сертификата (сертификат был добавлен в gitignore):
```Powershell
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout key.pem \
  -out cert.pem \
  -days 365 \
  -subj "/CN=localhost"
```
- Конфиг NGINX:
```
events {}

http {
    log_format main '$remote_addr - $remote_user [$time_local] '
                    '"$request" $status $body_bytes_sent '
                    '"$http_x_request_id"';

    access_log /var/log/nginx/access.log main;

    resolver 127.0.0.11 valid=30s;

    server {
        listen 8443 ssl;
        server_name localhost;

        ssl_certificate     /etc/nginx/tls/cert.pem;
        ssl_certificate_key /etc/nginx/tls/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;

        location / {
            set $upstream tasks:8082;
            proxy_pass http://$upstream;
            
            proxy_set_header Host $host;
            proxy_set_header X-Forwarded-Proto https;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Request-ID $http_x_request_id;
            proxy_set_header Authorization $http_authorization;
        }
    }
}
```
- Описание БД:
Создаются база данных tasks_db и таблица tasks с полями id, title, description, due_date, done, created_at
- Демонстрация SQLi
fmt.Sprintf("SELECT id, title, description, due_date, done, created_at FROM tasks WHERE title LIKE '%%%s%%'", keyword)

Исправленный код (параметризованный запрос с $1)
query := `SELECT id, title, description, due_date, done, created_at FROM tasks WHERE title LIKE $1`
rows, err := r.db.Query(query, "%"+keyword+"%")

keyword передаётся как параметр, база данных не воспринимает его как исполняемый SQL код.

Команда для проверки уязвимости в учебном стенде
<img width="714" height="912" alt="image" src="https://github.com/user-attachments/assets/5e9406b2-e03e-4133-85f9-2831eac5de50" />  



# Ответы на вопросы
1.	Какие свойства даёт TLS соединению?
TLS обеспечивает шифрование данных между клиентом и сервером, защиту от подслушивания и гарантию, что клиент общается с подлинным сервером, а не с злоумышленником.
2.	Почему самоподписанный сертификат не подходит для реального продакшна?
Самоподписанный сертификат не подходит для продакшна, потому что он не подтверждён доверенным центром сертификации, и браузеры/клиенты будут показывать предупреждение о небезопасном соединении, а также отсутствует возможность отзыва сертификата при компрометации ключа.
3.	В чём отличие TLS-терминации на NGINX от TLS в приложении?
При TLS-терминации на NGINX шифрование снимается на уровне прокси и дальше внутри сети трафик идёт по HTTP, что позволяет централизованно управлять сертификатами; при TLS в приложении каждый сервис сам отвечает за шифрование, что усложняет управление сертификатами и увеличивает нагрузку.
4.	Как возникает SQL-инъекция?
SQL-инъекция возникает, когда приложение склеивает строку запроса с пользовательским вводом, позволяя злоумышленнику подменить условие или выполнить произвольный SQL-код, например введя ' OR '1'='1.
5.	Почему параметризованный запрос защищает от SQLi?
Параметризованный запрос защищает от SQLi, потому что пользовательский ввод передаётся как параметр, а не как часть SQL-кода, поэтому база данных воспринимает его как данные, а не как исполняемые команды.
6.	Почему детали ошибок БД нельзя показывать клиенту?
Детали ошибок БД нельзя показывать клиенту, потому что они могут раскрыть структуру таблиц, имена полей или другую внутреннюю информацию, что поможет злоумышленнику подготовить более точную атаку, а в логах нужно фиксировать подробности для диагностики.
