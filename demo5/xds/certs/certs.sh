#!/bin/bash

openssl genrsa -out envoy-root-ca.key 4096

openssl req -new -x509 -days 365 -key envoy-root-ca.key -subj "/CN=Root CA" -out envoy-root-ca.crt

openssl req -nodes -new -keyout envoy-intermediate-ca.key -out envoy-intermediate-ca.csr -subj "/CN=Intermediate CA"

openssl x509 -days 365 -req -in envoy-intermediate-ca.csr -CAcreateserial -CA envoy-root-ca.crt -CAkey envoy-root-ca.key -out envoy-intermediate-ca.crt

openssl req -nodes -new -keyout envoy-proxy-server.key -out envoy-proxy-server.csr -subj "/CN=http.domain.com"

openssl x509 -days 365 -req -in envoy-proxy-server.csr -CAcreateserial -CA envoy-intermediate-ca.crt -CAkey envoy-intermediate-ca.key -out envoy-proxy-server.crt

openssl req -nodes -new -keyout envoy-proxy-client.key -out envoy-proxy-client.csr -subj "/CN=http.domain.com"

openssl x509 -days 365 -req -in envoy-proxy-client.csr -CAcreateserial -CA envoy-intermediate-ca.crt -CAkey envoy-intermediate-ca.key -out envoy-proxy-client.crt

cat envoy-root-ca.crt envoy-intermediate-ca.crt > envoy-intermediate-and-envoy-root-ca-chain.crt