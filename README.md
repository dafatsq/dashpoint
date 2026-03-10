# DashPoint POS

DashPoint is a Point of Sale (POS) application featuring a robust Go backend and a modern Next.js frontend. 

This guide will help you set up and run the DashPoint application locally on your machine.

## Prerequisites

Before you begin, ensure you have the following installed on your system:
- **[Git](https://git-scm.com/)**: For cloning the repository.
- **[Docker & Docker Compose](https://www.docker.com/products/docker-desktop/)**: Required to easily run the PostgreSQL database (and optionally the backend API).
- **[Node.js](https://nodejs.org/)** (v20+ recommended) & **npm**: Required for running the Next.js frontend.
- **[Go](https://go.dev/)** (v1.21+): Only required if you wish to run the backend natively outside of Docker.

## Getting Started

Follow these step-by-step instructions to get the application running on your local machine.

### 1. Clone the Repository

First, clone the repository to your local machine and navigate into the project directory:

```bash
git clone https://github.com/your-username/dashpoint.git
cd dashpoint
```

*(Replace `your-username` with the actual GitHub username where the repository is hosted).*

### 2. Set Up Environment Variables

The backend relies on environment variables for configuration. A template file is provided.

Copy the `.env.example` file to create your own `.env` file:

```bash
cp .env.example .env
```

The default values in `.env` are already configured for local development and will work seamlessly with the provided Docker Compose setup.

### 3. Start the Database and Backend API

The easiest way to run the backend and its PostgreSQL database is using Docker Compose. The `docker-compose.yml` file is configured to spin up both the `db` (Postgres) and the `backend` (Go API) services, automatically running database migrations on startup.

Run the following command in the root of the project:

```bash
docker-compose up -d
```

*Note: The `-d` flag runs the containers in the background (detached mode). To view the logs, you can run `docker-compose logs -f`.*

The backend API will now be accessible at `http://localhost:8080`.

*(Alternatively, to run the Go backend natively: Start only the database with `docker-compose up -d db`. Then, open a new terminal, navigate to the `backend` directory, run `go mod download`, and start the server with `go run cmd/server/main.go`)*.

### 4. Start the Frontend (Next.js)

With the backend and database running, you can now start the frontend web application.

Open a new terminal window, navigate to the `frontend` directory, install the dependencies, and start the development server:

```bash
cd frontend
npm install
npm run dev
```

The Next.js development server will start, typically on port 3000. 

### 5. Access the Application

Open your favorite web browser and navigate to:

**[http://localhost:3000](http://localhost:3000)**

You should now see the DashPoint frontend, successfully communicating with your local backend and database!

---

## Troubleshooting

- **Database Connection Issues:** Ensure Docker is running. If the backend fails to connect to the database, verify that the `db` container is fully initialized and healthy.
- **Port Conflicts:** If ports `8080` (Backend), `5432` (Postgres), or `3000` (Frontend) are already in use by other applications, you will need to stop those applications or modify the respective configurations (`docker-compose.yml`, `.env`, or frontend start scripts).
