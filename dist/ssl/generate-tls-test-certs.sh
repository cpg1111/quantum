#!/bin/bash

# Copyright (c) 2016 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

rm -rf $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/certs
mkdir -p $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/certs

rm -rf $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/csrs
mkdir -p $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/csrs

rm -rf $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/keys
mkdir -p $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/keys

rm -rf $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/data
mkdir -p $GOPATH/src/github.com/Supernomad/quantum/dist/ssl/data

pushd $GOPATH/src/github.com/Supernomad/quantum/dist/ssl 2>&1 > /dev/null

touch data/index.txt
echo '01' > data/serial

export SAN="IP:127.0.0.1, IP:172.18.0.4, IP:fd00:dead:beef::4"

# Create CA cert
yes | openssl req -config etcd-openssl.cnf -passout pass:quantum -new -x509 -extensions v3_ca -keyout keys/ca.key -out certs/ca.crt -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=ca.quantum.dev"

# Create and sign etcd server certificate
yes | openssl req -config etcd-openssl.cnf -new -nodes -keyout keys/etcd.quantum.dev.key -out csrs/etcd.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=etcd.quantum.dev"
yes | openssl ca -config etcd-openssl.cnf -passin pass:quantum -extensions etcd_server -keyfile keys/ca.key -cert certs/ca.crt -out certs/etcd.quantum.dev.crt -infiles csrs/etcd.quantum.dev.csr

# Create and sign etcd client certificates
yes | openssl req -config etcd-openssl.cnf -new -nodes -keyout keys/quantum0.quantum.dev.key -out csrs/quantum0.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=quantum0.quantum.dev"
yes | openssl ca -config etcd-openssl.cnf -passin pass:quantum -extensions etcd_client -keyfile keys/ca.key -cert certs/ca.crt -out certs/quantum0.quantum.dev.crt -infiles csrs/quantum0.quantum.dev.csr

yes | openssl req -config etcd-openssl.cnf -new -nodes -keyout keys/quantum1.quantum.dev.key -out csrs/quantum1.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=quantum1.quantum.dev"
yes | openssl ca -config etcd-openssl.cnf -passin pass:quantum -extensions etcd_client -keyfile keys/ca.key -cert certs/ca.crt -out certs/quantum1.quantum.dev.crt -infiles csrs/quantum1.quantum.dev.csr

yes | openssl req -config etcd-openssl.cnf -new -nodes -keyout keys/quantum2.quantum.dev.key -out csrs/quantum2.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=quantum2.quantum.dev"
yes | openssl ca -config etcd-openssl.cnf -passin pass:quantum -extensions etcd_client -keyfile keys/ca.key -cert certs/ca.crt -out certs/quantum2.quantum.dev.crt -infiles csrs/quantum2.quantum.dev.csr

popd 2>&1 > /dev/null
