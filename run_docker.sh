docker build -t terriyaki-go .
docker run --env-file .env -p 8080:8080 terriyaki-go