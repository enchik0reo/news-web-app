# Использует образ Golang в качестве базового образа для сборки приложения. Образ будет помечен как build.
FROM golang:1.21.6 as build

# Устанавливает рабочий каталог внутри контейнера
WORKDIR /src

# Копирование исходного кода
COPY ../../ /src

# Сборка первого приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/authsrv ./app/authService/cmd/auth_service/main.go

# Второй этап: создание образа для выполнения приложения
FROM alpine:3.19.0

# Копирование SSL сертификатов
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копирование доп файлов
COPY --from=build /src/configs src/configs
COPY --from=build /src/.env ./
COPY --from=build /src/migrations src/migrations

# Копирование исполняемого файла из первого этапа
COPY --from=build /bin/authsrv /bin/authsrv

# Команда по умолчанию
CMD ["/bin/authsrv"]