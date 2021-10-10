openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=yngwie proxy CA" -config san.cnf -config san.cnf -extensions v3_req
openssl genrsa -out cert.key 2048
mkdir certs/