# TP_Security_Light_Proxy

Light HTTP-Proxy. Proxy HTTP and HTTPS requests

HTTPS part:
gen new cert with 
```
cd gen_cert
./gen_ca.sh
```
return to parent folder with
```
cd ..
```
add cert to your system with
```
sudo cp gen_cert/ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

If you want to use with a browser, don't forget to add the cert to your browser. In Chrome:
```Settings -> Privacy and security -> Security -> Manage certificates -> Authorities -> Import```
Choose the generated file.

start server with 

```
go run main.go
```

try work http with 
```
curl -x http://127.0.0.1:8080 http://mail.ru
```
You can launch Chrome and start working in it.



