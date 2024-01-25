# NOTE:
# Commands below assume HTTP
# If want to use with TLS (if enabled on server), add "-k" to disable cert check
# AND remember to change port - default will be 8443
# i.e:
# curl -k -X POST -H "Bearer somebadvalue" "localhost:8443/jwt/login"

# Get JWT token
curl -X POST -H "Content-Type: application/json" -d '{"username":"user", "password":"pass"}' "localhost:8080/jwt/login"

# Invalid Bearer
curl -X POST -H "Bearer somebadvalue" "localhost:8080/jwt/login"


# "protected"  "/" group using authMiddleware
# That is: /ping, /json

# Valid Bearer
curl -H "Authorization: Bearer bearer_secret" "localhost:8080/ping"
curl -H "Authorization: Bearer bearer_secret" "localhost:8080/json"

# Some invalid Bearer
curl -H "Authorization: Bearer wrongone" "localhost:8080/ping"
curl -H "Authorization: Bearer wrongone" "localhost:8080/json"

# "protected"  "/jwt" group
# That is: /jwt/login (jwtLoginHandler),  /jwt/secret (jwtVerifyMiddleware)


# Get fresh JWT token from /jwt/login
curl -X POST -H "Content-Type: application/json" -d '{"username":"user", "password":"pass"}' "localhost:8080/jwt/login"


# Access /jwt/secret endpoint
# Requires valid JWT, should fail otherwise

# Expired token case
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDM5ODQwNjcsInVzZXJuYW1lIjoidXNlciJ9.xZ6oeKkbsMjVv4UvGhJ5eNjU_MInCK2fjF9k_HMnK6Y" http://localhost:8080/jwt/secret

# Invalid token
curl -H "Authorization: Bearer faketokenlol.eyJleHAiOjE3MDM5ODQwNjcsInVzZXJuYW1lIjoidXNlciJ9.xZ6oeKkbsMjVv4UvGhJ5eNjU_MInCK2fjF9k_HMnK6Y" http://localhost:8080/jwt/secret



# unrpotected group of paths 
# That is: /unprotected, /render, /special, /ws

# No login is required for these


curl -i "http://localhost:8080/unprotected"

# /render requires some URL parameters to work
curl -i "http://localhost:8080/render?mandatory1=lmao&mandatory2=12&optional2=true"


# Special allows to pass arbitrary command via URL parameters
# Example for "id" command
# More complex commands i.e. with spaces should be properly URL-encoded

curl -i "http://localhost:8080/special?lol=id"

# ?start=true param launches reverse shell. Requires tool like nc listening on 
# In code hardcoded to "127.0.0.1:9000"
# nc -n -l -p 9000

curl -i "http://localhost:8080/special?start=true"