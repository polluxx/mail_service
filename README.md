# Implemented tasks #
1. API service. See routes section.
2. Cursor-based pagination for getting messages
3. Old messages become expired in one hour
4. Concurrent access is supported through multi-threading
5. Solution that stores all data entirely in-memory and in-process
6. Support receiving messages via SMTP. See SMTP listener section.
7. Unit tests are written to test inmemory database

# Possible ways to improve performance #
1. Improve inmemory database.
*  Now golang chanels are used to sync data between threads.
*  There is another approach, it`s using mutex for thread synchronization.
*  Way with mutex may be faster because we can operate with Read and Write locks.
*  It means that the same resource can be read in parallel
2. Hardware upgrade
*  RAM and memory bus may be upgrated

# Routes #

* POST /mailboxes
* POST /mailboxes/{email address}/messages
* GET /mailboxes/{email address}/messages: Cursor pagination with "?maxId={maxId}" param
* GET /mailboxes/{email address}/messages/{message id}:
* DELETE /mailboxes/{email address}
* DELETE /mailboxes/{email address}/messages/{message id}

# SMTP listener #

* Working on 127.0.0.1:2525 (if in docker container - docker_host:2525)

# How to start app with Docker #

* install docker-engine on your machine
* in app directory "sudo docker build ."
* "sudo docker run {here must be docker image ID} -P"
* docker host must be http://172.17.0.1/ (or increment by 1 if some docker image was already run)
* app api must be 127.0.0.1:8080 (or with docker - docker_host:8080)
* to test - http://docker_host:8080/mailboxes/email_3@some.domain/messages

# How to start app on localhost #

* install latest go (1.6 v)
* set GOPATH (export GOPATH=$HOME/{your_go_folder})
* copy content to GOPATH
* in GOPATH folder "go build"
* start builded app (in current folder)