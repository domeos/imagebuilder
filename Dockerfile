FROM docker:1.8 

COPY dockerize /code/dockerize
COPY dockerize /user/local/bin/dockerize
COPY dockerize /user/local/bin/domeize
COPY imagebuilder /user/local/bin/build

ENTRYPOINT  ["build"]
