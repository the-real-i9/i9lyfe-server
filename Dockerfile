# Stage 1: Build the Node.js app
FROM node:20.10.0-alpine as build

# Copy your Node.js app code
COPY . /app

# Install dependencies
WORKDIR /app
RUN npm install

# Stage 2: Build the PostgreSQL sidecar
FROM alpine:latest

# Copy the PostgreSQL data directory (if needed)
COPY --from=build /app/data /var/lib/postgresql/data

# Install psql
RUN apk add --no-cache psql

# Set environment variables for database credentials
ENV POSTGRES_USER myuser
ENV POSTGRES_PASSWORD mypassword
ENV POSTGRES_DB mydatabase

# Run psql to execute your SQL file
RUN psql -U postgres -c "CREATE DATABASE $POSTGRES_DB;" && \
    psql -U postgres -d $POSTGRES_DB -c "ALTER USER $POSTGRES_USER WITH PASSWORD '$POSTGRES_PASSWORD';" && \
    psql -U $POSTGRES_USER -d $POSTGRES_DB -f init.sql

# Start the PostgreSQL container
CMD ["postgres", "-D", "/var/lib/postgresql/data"]

# Stage 3: Combine the Node.js app and PostgreSQL
FROM build

# Copy the PostgreSQL container
COPY --from=postgres /var/lib/postgresql/data /var/lib/postgresql/data

# Start the Node.js app
CMD ["npm", "start"]
