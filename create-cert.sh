# Generate a private key
openssl genpkey -algorithm RSA -out server.key

# Generate a certificate signing request (CSR)
openssl req -new -key server.key -out server.csr

# Generate a self-signed certificate
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt
