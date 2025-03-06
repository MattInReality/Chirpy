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