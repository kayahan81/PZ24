---
# Практическое задание 24

## ЭФМО-02-25 

## Алиев Каяхан Командар оглы
---
# Тема работы
CI/CD pipeline: тесты, сборка, упаковка Docker-образа и деплой.

## Цели занятия
Сделать автоматическую проверку качества и сборку проекта при каждом push/merge, а также подготовить базовый конвейер доставки (build → package → publish → deploy).

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
├───.github
│   └───workflows
│           ci.yml
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
│   │   │   ├───middleware
│   │   │   └───service
│   │   └───pkg
│   │       └───authpb
│   │               auth.pb.go
│   │               auth_grpc.pb.go
│   │
│   └───tasks
│       │   .dockerignore
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
│           │       tasks_test.go
│           │
│           ├───http
│           ├───middleware
│           │       auth.go
│           │       auth_cookie.go
│           │       csrf.go
│           │       metric.go
│           │       security_headers.go
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
    ├───csrf
    │       generator.go
    │
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
git clone https://github.com/kayahan81/pz24
```
## 2) Перейти в папку pz19:
```Powershell
cd pz24
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
## Деплой
<img width="1196" height="290" alt="image" src="https://github.com/user-attachments/assets/41c24587-8e16-4cc2-a105-2ccf7515b69f" />
<img width="1204" height="213" alt="image" src="https://github.com/user-attachments/assets/df673ec6-74fc-4f73-a9be-f94b848816c1" />

## Публикация образа по тегу
<img width="1161" height="105" alt="image" src="https://github.com/user-attachments/assets/9fe1a3ad-0464-4719-af81-3807db3aa322" />
<img width="376" height="139" alt="image" src="https://github.com/user-attachments/assets/c8db422b-4346-4bdb-b076-4a3fdd2444b1" />



# Отчёт
1.	Файл pipeline (ci.yml или .gitlab-ci.yml)
```yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, master, develop]
    tags: ['v*']
  pull_request:
    branches: [main, master]

env:
  GO_VERSION: '1.25'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/tasks

jobs:
  # ============================================
  # JOB 1: Тесты и сборка
  # ============================================
  test-and-build:
    name: Test & Build
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v ./...
      
      - name: Build Auth
        run: go build -o /dev/null ./services/auth/cmd/auth
      
      - name: Build Tasks
        run: go build -o /dev/null ./services/tasks/cmd/tasks

  # ============================================
  # JOB 2: Сборка Docker образа
  # ============================================
  docker-build:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: test-and-build
    if: github.event_name == 'push' && (github.ref == 'refs/heads/main' || github.ref == 'refs/heads/master')
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./services/tasks/Dockerfile
          load: true
          tags: techip-tasks:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Verify image
        run: docker images | grep techip-tasks

  # ============================================
  # JOB 3: Публикация в Registry (при теге)
  # ============================================
  docker-push:
    name: Push to GHCR
    runs-on: ubuntu-latest
    needs: docker-build
    if: startsWith(github.ref, 'refs/tags/')
    
    permissions:
      contents: read
      packages: write
      
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=semver,pattern={{version}}
            type=sha,prefix={{date:YYYYMMDDHHmmss}}-
            type=raw,value=latest
      
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./services/tasks/Dockerfile
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```
2.	Описание шагов pipeline (что и в каком порядке делается)
test-and-build — проверка тестов и компиляции
docker-build — сборка Docker образа
3.	Скрин/лог успешного прогона: тесты + build + docker build
<img width="1351" height="448" alt="image" src="https://github.com/user-attachments/assets/9f4fc85e-3601-4f5e-8f57-8fe3ea225603" />
4.	Если есть push в registry — указать, куда публикуется образ и как формируется тег
<img width="1173" height="231" alt="image" src="https://github.com/user-attachments/assets/2020bd51-e885-4055-9451-15857452b4b8" />



# Ответы на вопросы
1.	Чем CI отличается от CD?
CI — непрерывная интеграция (тесты, сборка), CD — непрерывная доставка (деплой)
2.	Почему go test должен запускаться в pipeline?
go test должен запускаться в pipeline, чтобы поймать ошибки до того, как код попадёт в основную ветку
3.	Что такое секреты CI и почему их нельзя хранить в репозитории?
Секреты CI — это зашифрованные переменные, их нельзя хранить в репозитории, чтобы злоумышленник не украл токены/ключи
4.	Почему важно версионировать docker-образы?
Версионировать образы важно, чтобы можно было откатиться к предыдущей версии и понять, что именно развёрнуто на сервере
5.	Какие риски у автоматического деплоя без ручного контроля?
Можно случайно задеплоить сломанную версию, если тесты пропустили баг; нужен ручной контроль для критичных систем
