# Calculator Server

HTTP-сервис для выполнения вычислений с использованием нативных библиотек C и Rust.

## Запуск

Для сборки и запуска сервиса выполните:

docker compose up --build -d

После запуска сервис будет доступен на порту 8080.

## API

Выполнение расчёта
POST http://localhost:8080/api/v1/calc?num={value}

Пример:

curl -X POST "http://localhost:8080/api/v1/calc?num=10"

## Метрики

GET http://localhost:8080/api/v1/metrics

Метрики возвращаются в формате Prometheus.

Пример:

curl "http://localhost:8080/api/v1/metrics"

Метрики включают:

RPS HTTP-запросов за каждую секунду последней минуты;
p95 времени выполнения функции на C;
p99 времени выполнения функции на C;
p95 времени выполнения функции на Rust;
p99 времени выполнения функции на Rust.

## Остановка

Остановить сервис:

docker compose down
