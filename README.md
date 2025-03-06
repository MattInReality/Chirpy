# Chirpy API

A RESTful API for the Chirpy social media platform. The project is a guided project built during a course from the great
team at [boot.dev](https://boot.dev). Many of the problems are solved my way but the general structure of the project
is aligned with the course material.

## Development
- Server runs on port 8080 by default
- Requires PostgreSQL database
- Uses environment variables for configuration:
    - `DB_URL`: Database connection string
    - `JWT_SECRET`: Secret for JWT signing
    - `POLKA_KEY`: API key for webhook authentication
    - `PLATFORM`: Platform environment setting

## Getting Started
1. Set up environment variables
2. Ensure PostgreSQL is running
3. Start the server:
   ```bash
   go run main.go
   ```
4. Server will be available at `http://localhost:8080`

## License
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## API Endpoints

### Health Check
**GET `/api/healthz`**
- Checks if the API is running
- Returns 200 OK when server is operational

### Authentication

#### Create Account
**POST `/api/users`**
```json
{
    "email": "user@example.com",
    "password": "yourpassword"
}
```
- Creates a new user account
- Email must be valid
- Returns user information

#### Login
**POST `/api/login`**
- Authenticates user credentials
- Returns authentication token

#### Update User
**PUT `/api/users`**
- Updates user information
- Requires authentication

#### Token Management
**POST `/api/refresh`**
- Refreshes authentication token

**POST `/api/revoke`**
- Revokes refresh token

### Chirps

#### Create Chirp
**POST `/api/chirps`**
- Creates a new chirp post
- Requires authentication

#### List Chirps
**GET `/api/chirps`**
- Retrieves all chirps
- Accepts query params for user_id and sorting
```aiignore
GET /api/chirps?sort=asc
GET /api/chirps?sort=desc
GET /api/chirps?sort=desc&user_id=thisisauserid
```

#### Get Single Chirp
**GET `/api/chirps/{chirpID}`**
- Retrieves a specific chirp by ID

#### Delete Chirp
**DELETE `/api/chirps/{chirpID}`*## Development
- Server runs on port 8080 by default
- Requires PostgreSQL database
- Uses environment variables for configuration:
    - `DB_URL`: Database connection string
    - `JWT_SECRET`: Secret for JWT signing
    - `POLKA_KEY`: API key for webhook authentication
    - `PLATFORM`: Platform environment setting

## Getting Started
1. Set up environment variables
2. Ensure PostgreSQL is running
3. Start the server:
   ```bash
   go run main.go
   ```
4. Server will be available at `http://localhost:8080`

## License
[Add your license information here]*
- Deletes a specific chirp
- Requires authentication
- User can only delete their own chirps

### Admin Controls

#### View Metrics
**GET `/admin/metrics`**
- Displays system metrics
- Shows total visit count
- Returns HTML format

#### Reset System
**POST `/admin/reset`**
- Resets all users
- Only available in development environment
- Platform must be set to "dev"

### Webhooks

#### Polka Webhook
**POST `/api/polka/webhooks`**
- Handles Polka webhook notifications
- Requires valid API key

## Authentication
The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:
# Chirpy API

A RESTful API for the Chirpy social media platform.

## API Endpoints

### Health Check
**GET `/api/healthz`**
- Checks if the API is running
- Returns 200 OK when server is operational

### Authentication

#### Create Account
**POST `/api/users`**
```json
{
    "email": "user@example.com",
    "password": "yourpassword"
}
```
- Creates a new user account
- Email must be valid
- Returns user information

#### Login
**POST `/api/login`**
- Authenticates user credentials
- Returns authentication token

#### Update User
**PUT `/api/users`**
- Updates user information
- Requires authentication

#### Token Management
**POST `/api/refresh`**
- Refreshes authentication token

**POST `/api/revoke`**
- Revokes refresh token

### Chirps

#### Create Chirp
**POST `/api/chirps`**
- Creates a new chirp post
- Requires authentication

#### List Chirps
**GET `/api/chirps`**
- Retrieves all chirps

#### Get Single Chirp
**GET `/api/chirps/{chirpID}`**
- Retrieves a specific chirp by ID

#### Delete Chirp
**DELETE `/api/chirps/{chirpID}`**
- Deletes a specific chirp
- Requires authentication
- User can only delete their own chirps

### Admin Controls

#### View Metrics
**GET `/admin/metrics`**
- Displays system metrics
- Shows total visit count
- Returns HTML format

#### Reset System
**POST `/admin/reset`**
- Resets all users
- Only available in development environment
- Platform must be set to "dev"

### Webhooks

#### Polka Webhook
**POST `/api/polka/webhooks`**
- Handles Polka webhook notifications
- Requires valid API key

## Authentication
The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

