# Использует образ Golang в качестве базового образа для сборки приложения. Образ будет помечен как build.
FROM golang:1.21.6 as build

# Устанавливает рабочий каталог внутри контейнера
WORKDIR /src

# Копирование исходного кода
COPY ../../ /src

# Сборка первого приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/apisrv ./app/apiService/cmd/api_service/main.go

# Второй этап: создание образа для выполнения приложения
FROM scratch

# Копирование SSL сертификатов
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копирование configs и .env
COPY --from=build /src/configs src/configs
COPY --from=build /src/.env ./

# Копирование исполняемого файла из первого этапа
COPY --from=build /bin/apisrv /bin/apisrv

# Команда по умолчанию
CMD ["/bin/apisrv"]