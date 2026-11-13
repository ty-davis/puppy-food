# Puppy feeding schedule

A local puppy-feeding schedule with a Go and SQLite backend and a Vue/TypeScript frontend.

## Run locally

```sh
make install
make dev
```

The frontend is available at <http://localhost:5173>. The Go API listens on <http://localhost:8080>.

## Build and run

```sh
make build
cd backend && go run ./cmd/server
```

The SQLite database is created at `backend/data/puppy-food.db` by default. Set `DATABASE_PATH` to use another path. The server serves `frontend/dist` when it exists.

## API schema and client

The API schema is generated from the endpoint definitions in Go and saved in `backend/openapi.json`. The frontend SDK is generated from that JSON schema with HeyAPI.

```sh
make generate
```

## Deploy with Docker

Docker builds the Vue app and Go server into one container. The server hosts the frontend and API at the same address, and the SQLite database lives in a persistent Docker volume.

### 1. Install Docker

Install Docker Engine and the Docker Compose plugin on the machine that will host the app. Check that both commands work:

```sh
docker --version
docker compose version
```

### 2. Set the host port

Create a local `.env` file from the example:

```sh
cp .env.example .env
```

Edit `.env` if needed:

```env
APP_BIND_ADDRESS=0.0.0.0
APP_PORT=8080
```

`APP_PORT` is the port people use in their browser. To use port `8090`, change it to `8090`. `APP_BIND_ADDRESS=0.0.0.0` makes it reachable on the server’s network interfaces; use `127.0.0.1` to keep it accessible only from the server itself.

### 3. Build and start

From this project directory, run:

```sh
docker compose up -d --build
```

The first build downloads the base images and dependencies, so it can take a few minutes. Later starts are faster. Check the container with:

```sh
docker compose ps
docker compose logs -f puppy-food
```

Open `http://localhost:8080` on the server. If you changed `APP_PORT`, use that port instead.

### 4. Open it from another device on your LAN

Find the server’s LAN IP address (for example, with `hostname -I` on Linux), then open `http://SERVER-LAN-IP:8080` on the phone or computer. Substitute your configured `APP_PORT` if it is not `8080`. The server’s firewall must allow that TCP port.

The frontend automatically sends API requests to the same origin it was loaded from. You do **not** need to set a new frontend URL or rebuild the app when its LAN IP or host port changes. In this setup the Go server serves both the frontend and `/api` endpoints.

### 5. Stop, update, and keep your data

Stop the app without removing its data:

```sh
docker compose down
```

Start it again with `docker compose up -d`. To install code changes, pull/copy the updated project and run `docker compose up -d --build` again. The `puppy-food-data` volume keeps the database across container rebuilds and normal `down`/`up` cycles.

**Do not run `docker compose down -v` unless you intend to delete the database volume and all feeding data.** Back up the database volume regularly if the schedule matters to you.

### Network safety

This app has no authentication. Anyone who can reach its address can change schedules and feeding completion states. Keep it on a trusted LAN, and do not forward its port from your router or expose it directly to the public internet.

### Container settings

The Compose file sets `LISTEN_ADDRESS`, `DATABASE_PATH`, and `WEB_DIST` for the container. `APP_PORT` only changes the host-side port mapping; the app continues listening on port `8080` inside the container. The database is stored in the `puppy-food-data` volume mounted at `/data`.
