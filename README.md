

# MYSQL Docker

```
docker run -d -it --name mysql -e MYSQL_ROOT_PASSWORD=12345678 -p 3306:3306 mysql:9.0.1
```

# REDIS Docker
```
docker run -d -it --name redis -p 6379:6379 redis:7.4.0
docker run -d -it --name redis -p 6380:6380 redis:7.4.0
docker run -d -it --name redis -p 6381:6381 redis:7.4.0
docker run -d -it --name redis -p 6382:6382 redis:7.4.0
docker run -d -it --name redis -p 6383:6383 redis:7.4.0
```