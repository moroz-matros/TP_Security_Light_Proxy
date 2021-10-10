#!/bin/sh
openssl req -new -key ./gen_cert/cert.key -subj "/CN=$1" -sha256 | openssl x509 -req -days 3650 -CA ./gen_cert/ca.crt -CAkey ./gen_cert/ca.key -set_serial "$2" > ./gen_cert/certs/"$1".crt