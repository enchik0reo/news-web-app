# Использует образ Golang в качестве базового образа для сборки приложения. Образ будет помечен как build.
FROM golang:1.21.6 as build

# Устанавливает рабочий каталог внутри контейнера
WORKDIR /src

# Копирование исходного кода
COPY ../../ /src

# Сборка первого приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrator ./migrations/migrator/main.go

# Второй этап: создание образа для выполнения приложения
FROM scratch

# Копирование SSL сертификатов
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копирование доп файлов
COPY --from=build /src/configs src/configs
COPY --from=build /src/.env ./
COPY --from=build /src/migrations src/migrations

# Копирование исполняемых файлов из первого этапа
COPY --from=build /bin/migrator /bin/migrator

# Команда для запуска по умолчанию
CMD ["/bin/migrator"]