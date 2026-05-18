# ============================================
# STAGE 1: Builder - компилируем бинарник
# ============================================
FROM golang:1.25-alpine AS builder

# Устанавливаем необходимые утилиты
RUN apk add --no-cache git ca-certificates

# Рабочая директория
WORKDIR /app

# Копируем go.mod и go.sum (из корня проекта, на уровень выше)
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь код (включая services/)
COPY . .

# Собираем бинарник (указываем правильный путь)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -o /app/bin/tasks ./services/tasks/cmd/tasks

# ============================================
# STAGE 2: Runner - минимальный образ
# ============================================
FROM alpine:latest

# Устанавливаем CA сертификаты
RUN apk --no-cache add ca-certificates

# Создаём непривилегированного пользователя
RUN adduser -D -u 1000 appuser

# Рабочая директория
WORKDIR /home/appuser

# Копируем бинарник из builder стадии
COPY --from=builder /app/bin/tasks .

# Меняем владельца
RUN chown -R appuser:appuser /home/appuser
USER appuser

# Порт
EXPOSE 8082

# Запуск
CMD ["./tasks"]