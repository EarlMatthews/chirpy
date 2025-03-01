# Chirpy

Chirpy is a simple HTTP server that serves static files and provides an API endpoint for validating chirps.

## Features

- Serves static files from the current directory.
- Provides an API endpoint for validating chirps.
- Validates chirps to ensure that they are not too long (max 140 characters) and that they do not contain any bad words ("kerfuffle", "sharbert", "fornax").
- Provides a health check endpoint.
- Provides a metrics endpoint that shows the number of times the fileserver has been hit.
- Provides a reset endpoint that resets the fileserver hits counter.

## Endpoints

- `/app/`: Serves static files from the current directory.
- `/app/assets/`: Serves static files from the assets directory.
- `GET /admin/healthz`: Returns a health check status.
- `GET /admin/metrics`: Returns the number of times the fileserver has been hit.
- `POST /admin/reset`: Resets the fileserver hits counter.
- `POST /api/validate_chirp`: Validates a chirp.

## Validation

The `POST /api/validate_chirp` endpoint validates chirps to ensure that they are not too long (max 140 characters) and that they do not contain any bad words ("kerfuffle", "sharbert", "fornax"). If a chirp is invalid, the endpoint returns a 400 error. If a chirp is valid, the endpoint returns a 200 OK.

The bad words are replaced with asterisks.

## Usage

To run the server, execute the following command:

```sh
go run main.go
```

Then, open your browser and navigate to `http://localhost:8888/app/`.

You can also use the API endpoint to validate chirps. For example, to validate the chirp "Hello, world!", you can send a POST request to `http://localhost:8888/api/validate_chirp` with the following JSON body:

```json
{
    "body": "Hello, world!"
}
```

The server will respond with the following JSON body:

```json
{
    "cleaned_body": "Hello, world!"
}
```