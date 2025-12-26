docker build -t myapp .
-t = tag
It gives your image a name (and optionally a version)

You can also specify versions:
docker build -t myapp:1.0 .


docker run -d nginx
-d = detached

Runs the container in the background

Your terminal is free again


docker run -p 8080:80 nginx
-p (port mapping)
it means --- host_port : container_port
So:
localhost:8080  →  container:80


docker run --name my-nginx nginx
--name — give the container a friendly name
Instead of:
f3a1c92d7b4a
You get:
my-nginx

then you can log and stop by name:
docker stop my-nginx
docker logs my-nginx



docker run -it ubuntu bash
-it (interactive + terminal)
This is actually two flags combined:

Flag	Meaning
-i	    Keep STDIN open
-t	    Allocate a terminal

docker run --rm hello-world
--rm — auto delete container

docker ps
ps = “process status”
shows running process

docker ps -a
all process


## What Docker Compose is
Docker Compose lets you define multiple containers in one file and run them together.
Instead of many `docker run` commands → one `docker-compose.yml`

---

## Example: Backend + Database

version: "3.9"

services:
  backend:
    build: .
    container_name: my-backend
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql://postgres:password@db:5432/appdb
    depends_on:
      - db

  db:
    image: postgres:15
    container_name: my-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: appdb
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:

---

# Explanation (Line by Line)

## version: "3.9"
- Docker Compose file format version
- NOT Docker version
- 3.9 is common and stable

---

## services
- Top-level section
- Each service = one container

---

## backend (service)
- Service name
- Acts as hostname inside Docker network
- Other containers can reach it via `backend`

### build: .
- Builds image from Dockerfile in current directory
- Equivalent to: docker build .

### container_name: my-backend
- Explicit container name
- Easier than auto-generated names
- Equivalent to: docker run --name my-backend

### ports:
  - "8000:8000"
- Port mapping
- host_port : container_port
- localhost:8000 → container:8000
- Equivalent to: docker run -p 8000:8000

### environment:
- Environment variables inside container
- Used by backend to connect to database
- db = service name of database (Docker DNS)
js reads it like:
const dbUrl = process.env.DATABASE_URL;
env file is not included inside the docker file
docker automatically loads in env variables if they are in the same folder
e.g.
.env:
POSTGRES_PASSWORD=password
docker-compose.yml:
environment:
  POSTGRES_PASSWORD=${POSTGRES_PASSWORD}


### depends_on:
- Backend starts after db container starts
- Does NOT wait for db to be ready
- Only controls startup order

---

## db (service)
- Database container
- Service name = db
- Hostname inside Docker network

### image: postgres:15
- Pulls image from Docker Hub
- Uses Postgres version 15
- Equivalent to: docker pull postgres:15

### container_name: my-postgres
- Human-friendly name
- Useful for logs, exec, inspect

### environment (Postgres setup)
- POSTGRES_USER: postgres
- POSTGRES_PASSWORD: password
- POSTGRES_DB: appdb
- Required by official Postgres image

### volumes:
  - postgres_data:/var/lib/postgresql/data
- Persists database data
- Data survives container restarts

---

## volumes (top-level)
- Named volume managed by Docker
- Stored outside containers
- Automatically created if missing

---

# How to Run

docker compose up
- Start all services

docker compose up -d
- Start all services in background

docker compose ps
- Show running containers

docker compose down
- Stop and remove containers

docker compose down -v
- Stop containers AND delete volumes (database data lost)

---
Correct database URL (WORKS)
postgresql://postgres:password@db:5432/appdb
Incorrect database URLs (FAIL)
postgresql://postgres:password@localhost:5432/appdb ❌
postgresql://postgres:password@127.0.0.1:5432/appdb ❌

Compose
services:
  frontend:
  backend:

Inside frontend code

✅ Works:
fetch("http://backend:8000/api")

❌ Fails:
fetch("http://localhost:8000/api")

remember to use retry logic when connecting to other services, bc depends-on only creates one service after creating another, but the another service might not be ready by the time the one service is ready.


docker exec jobscanner-db psql -U jobscanner -d jobscanner -c "SELECT 1;"
“Run a PostgreSQL command inside a running Docker container and check if the database is responding.”

docker exec ...
Runs a command inside an already running container
Does not start a new container

jobscanner-db
container name

psql is the PostgreSQL command-line client
This works because the Postgres image includes psql

-U jobscanner
Connect as PostgreSQL user jobscanner

-d jobscanner
Connect to database named jobscanner

-c "SELECT 1;"
Run this SQL command
Exit immediately after

docker exec -it jobscanner-db psql -U jobscanner -d jobscanner
this makes it interactive so you can run SELECT 1 inside the psql database after running the above

docker run -e NODE_ENV=production myapp
run myapp image with process.env.NODE_ENV=production


docker run nginx:1.27.0-bookworm
run nginx with specific version


A digest is a cryptographic fingerprint of data.
Example (SHA-256)
Input:   hello
Digest:  2cf24dba5fb0a30e26e83b2ac5b9e29e
         1b161e5c1fa7425e73043362938b9824


use docker run -p nginx@sha256:6af7923abc2...
use digest to specific specific version as well


slim and alpine(minimal version of linux) are version that are much smaller containers


docker run -v mydata:/data <container and its commands that save data>
-v tells docker to save to a volume so data is saved



FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
Locally, go mod download is cheap because Go checks go.mod and uses its own module cache, so changing source files does not trigger re-downloads.
In Docker, caching is based on filesystem layers, so if you COPY . . before running go mod download, any file change invalidates the cache and forces dependencies to download again.
That’s why Dockerfiles copy go.mod/go.sum first and run go mod download before copying the rest of the source. So that when we change any source code in the future, docker could still use cache and not re download

every line of code in dockerfile is a layer that docker caches that makes up the image
when a line changes, the docker loses its cache for that line and every line below it and will rebuild

be careful bc if 1 line produces a different result another day, docker won't know it because it will still read from cache

docker build --no-cache .
tells docker to not cache

layers are immutable so event if a data is deleted in docker file, it still exists in the layers

EXPOSE 8000
doesn't do anything, just to tell dev that we are using 8000 port


CMD ["uvicorn", "mysite.main:app", "--host", "0.0.0.0", "--port", "8000"]
runs the command



the build process generates a lot of files that aren't needed when using it, we can separate the building and running commands in different stages

```
FROM python:3.12-slim AS builder

WORKDIR /app

COPY pyproject.toml requirements.txt ./
RUN pip wheel --no-cache-dir --no-deps --wheel-dir wheels -r requirements.txt
# no --no-cache-dir tells pip to not cache dependencies downloaded so it doesnt take up space in the image
COPY src src
RUN pip wheel --no-cache-dir --no-deps --wheel-dir wheels .

FROM python:3.12-slim AS runner

COPY --from=builder /app/wheels /wheels
RUN pip install --no-cache /wheels/* && rm -rf /wheels

EXPOSE 8000

CMD ["uvicorn", "mysite.main:app", "--host", "0.0.0.0", "--port", "8000"]
```
here, we have 2 build stages, builder and runner
we reference the files created by builder by using --from and we delete it after since we don't need it

if we used only 1 build stage, the building commands exist as layers in the image, taking up a lot of space. But with multi build stage, only the last build stage is contained in the container

docker build -t mysite-backend --target  runner .
uses the container called runner as the docker container

pull_policy: never, tells docker to never pull from docker hub for this docker container name
image: name of the image
container_name: name of the container

docker compose build
build image

docker compose up
runs the image as container

docker compose down
shut down container

we dont need a port for database because only backend talks to the database

env_file: replace "environment" with this and to tell docker the address of the .env file

volumes:
  - mongodb-data:/data/db
mongodb-data is the named Docker volume.
It is not a folder on your machine
It is managed entirely by Docker
Docker stores it somewhere like:
Linux: /var/lib/docker/volumes/mongodb-data/_data
macOS / Windows: inside Docker Desktop’s VM