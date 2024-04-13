# Официальный образ Go
FROM golang:1.22 as builder

# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

# Копируем модули зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код проекта в контейнер
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -v -o bank-service ./cmd

# Образ alpine для запуска приложения (меньший размер)
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем исполняемый файл в новый образ
COPY --from=builder /app/bank-service .

# Открываем порт, который использует приложение
EXPOSE 3000

# Запускаем приложение при старте контейнера
CMD ["./bank-service"]
