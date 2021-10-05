# TP_Security_Light_Proxy

Light HTTP-Proxy. 

start server with 

```
go run main.go
```

try work with 
```
curl -x http://127.0.0.1:8080 http://mail.ru
```

to build in Docker
```
 sudo docker build -t proxy .
 sudo docker run -p 8080:8080 -d --name=proxy_container proxy
```

