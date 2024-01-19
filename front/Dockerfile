# Этап 1: создание образа для установки зависимостей и сборки приложения
FROM node:21.6.0-slim as build
WORKDIR /app
COPY package.json .
COPY package-lock.json .
RUN npm install
COPY . .
RUN npm run build

# Этап 2: создание образа для запуска приложения
FROM nginx:1.24.0-bullseye-perl
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]