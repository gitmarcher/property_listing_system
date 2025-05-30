# Property Listing API Documentation

This document provides detailed information about the available API endpoints in the Property Listing system based on actual implementation.

## Base URL
```
http://localhost:3000/api
```

## Authentication
Most endpoints require authentication using JWT token. Include the token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

## API Endpoints

### Authentication

#### Register User
- **URL**: `/auth/register`
- **Method**: `POST`
- **Auth Required**: No
- **Body**:
```json
{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890"
}
```
- **Success Response** (201):
```json
{
    "success": true,
    "message": "User registered successfully",
    "data": {
        "user": {
            "id": "507f1f77bcf86cd799439011",
            "email": "user@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "phone": "+1234567890",
            "created_at": "2024-03-20T10:00:00Z"
        },
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
}
```
- **Error Responses**:
  - 400: Invalid request body, missing required fields, password too short
  - 409: User with email already exists
  - 500: Server error

#### Login User
- **URL**: `/auth/login`
- **Method**: `POST`
- **Auth Required**: No
- **Body**:
```json
{
    "email": "user@example.com",
    "password": "password123"
}
```
- **Success Response** (200):
```json
{
    "success": true,
    "message": "Login successful",
    "data": {
        "user": {
            "id": "507f1f77bcf86cd799439011",
            "email": "user@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "phone": "+1234567890",
            "created_at": "2024-03-20T10:00:00Z"
        },
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
}
```
- **Error Responses**:
  - 400: Invalid request body, missing email/password
  - 401: Invalid email or password
  - 500: Server error

### Properties (Public)

#### Get All Properties with Filtering
- **URL**: `/properties`
- **Method**: `GET`
- **Auth Required**: No
- **Query Parameters**:
  - `page` (default: 1): Page number for pagination
  - `limit` (default: 10, max: 100): Number of items per page
  - `min_price`: Minimum price filter
  - `max_price`: Maximum price filter
  - `state`: State filter (case-insensitive regex)
  - `city`: City filter (case-insensitive regex)
  - `type`: Property type filter (case-insensitive regex)
  - `bedrooms`: Number of bedrooms
  - `bathrooms`: Number of bathrooms
  - `verified`: Filter by verification status (true/false)
  - `furnished`: Furnished status filter (case-insensitive regex)
  - `sort_by`: Sort field (price, rating, areaSqFt, bedrooms, bathrooms)
  - `sort_order`: Sort order (asc/desc, default: asc)
- **Success Response** (200):
```json
{
    "success": true,l
    "data": [
        {
            "id": "PROP1001",
            "title": "Beautiful 2BR Apartment",
            "type": "Apartment",
            "price": 250000,
            "state": "California",
            "city": "San Francisco",
            "areaSqFt": 1200,
            "bedrooms": 2,
            "bathrooms": 2,
            "amenities": ["parking", "gym"],
            "furnished": "fully",
            "availableFrom": "2024-04-01",
            "tags": ["modern", "downtown"],
            "listingType": "sale",
            "rating": 4.5,
            "isVerified": true,
            "created_at": "2024-03-20T10:00:00Z"
        }
    ],
    "meta": {
        "page": 1,
        "limit": 10,
        "total": 50,
        "total_pages": 5
    }
}
```

#### Get Property by ID
- **URL**: `/properties/:id`
- **Method**: `GET`
- **Auth Required**: No
- **Success Response** (200):
```json
{
    "success": true,
    "data": {
        "id": "PROP1001",
        "title": "Beautiful 2BR Apartment",
        "type": "Apartment",
        "price": 250000,
        "state": "California",
        "city": "San Francisco",
        "areaSqFt": 1200,
        "bedrooms": 2,
        "bathrooms": 2,
        "amenities": ["parking", "gym"],
        "furnished": "fully",
        "availableFrom": "2024-04-01",
        "tags": ["modern", "downtown"],
        "listingType": "sale",
        "rating": 4.5,
        "isVerified": true,
        "created_at": "2024-03-20T10:00:00Z"
    }
}
```
- **Error Responses**:
  - 400: Property ID is required
  - 404: Property not found

#### Search Properties
- **URL**: `/properties/search`
- **Method**: `GET`
- **Auth Required**: No
- **Query Parameters**:
  - `q` (required): Search query
  - `page` (default: 1): Page number
  - `limit` (default: 10, max: 100): Items per page
- **Success Response** (200):
```json
{
    "success": true,
    "data": [
        {
            "id": "PROP1001",
            "title": "Beautiful House",
            "type": "House",
            "price": 350000,
            "state": "Texas",
            "city": "Austin",
            "areaSqFt": 2000,
            "bedrooms": 3,
            "bathrooms": 2,
            "amenities": ["garden", "garage"],
            "furnished": "semi",
            "availableFrom": "2024-05-01",
            "tags": ["family", "quiet"],
            "listingType": "sale",
            "rating": 4.8,
            "isVerified": true,
            "created_at": "2024-03-20T10:00:00Z"
        }
    ],
    "message": "Search completed successfully",
    "meta": {
        "page": 1,
        "limit": 10,
        "total": 25,
        "total_pages": 3
    }
}
```
- **Error Responses**:
  - 400: Search query is required
  - 500: Search failed

### Listings (Requires Authentication)

#### Get User's Listings
- **URL**: `/listings`
- **Method**: `GET`
- **Auth Required**: Yes
- **Query Parameters**:
  - `page` (default: 1): Page number
  - `limit` (default: 10, max: 100): Items per page
- **Success Response** (200):
```json
{
    "success": true,
    "data": [
        {
            "id": "PROP1001",
            "title": "My Property Listing",
            "type": "Apartment",
            "price": 180000,
            "state": "New York",
            "city": "Brooklyn",
            "areaSqFt": 900,
            "bedrooms": 1,
            "bathrooms": 1,
            "amenities": ["wifi", "heating"],
            "furnished": "unfurnished",
            "availableFrom": "2024-06-01",
            "tags": ["cozy", "affordable"],
            "listingType": "rent",
            "createdBy": "507f1f77bcf86cd799439011",
            "created_at": "2024-03-20T10:00:00Z",
            "updated_at": "2024-03-20T10:00:00Z",
            "isVerified": false,
            "rating": 0
        }
    ],
    "meta": {
        "page": 1,
        "limit": 10,
        "total": 5,
        "total_pages": 1
    }
}
```

#### Create Listing
- **URL**: `/listings`
- **Method**: `PUT`
- **Auth Required**: Yes
- **Body**:
```json
{
    "title": "Beautiful 2BR Apartment",
    "type": "Apartment",
    "price": 250000,
    "state": "California",
    "city": "San Francisco",
    "areaSqFt": 1200,
    "bedrooms": 2,
    "bathrooms": 2,
    "amenities": ["parking", "gym"],
    "furnished": "fully",
    "availableFrom": "2024-04-01",
    "tags": ["modern", "downtown"],
    "listingType": "sale"
}
```
- **Success Response** (201):
```json
{
    "success": true,
    "message": "Listing created successfully",
    "data": {
        "id": "PROP1002",
        "title": "Beautiful 2BR Apartment",
        "type": "Apartment",
        "price": 250000,
        "state": "California",
        "city": "San Francisco",
        "areaSqFt": 1200,
        "bedrooms": 2,
        "bathrooms": 2,
        "amenities": ["parking", "gym"],
        "furnished": "fully",
        "availableFrom": "2024-04-01",
        "tags": ["modern", "downtown"],
        "listingType": "sale",
        "createdBy": "507f1f77bcf86cd799439011",
        "created_at": "2024-03-20T10:00:00Z",
        "updated_at": "2024-03-20T10:00:00Z",
        "isVerified": false,
        "rating": 0
    }
}
```
- **Required Fields**: title, type, price, state, city, areaSqFt, bedrooms, bathrooms, furnished, availableFrom, listingType
- **listingType**: Must be either "rent" or "sale"

#### Update Listing
- **URL**: `/listings/:id`
- **Method**: `PATCH`
- **Auth Required**: Yes
- **Body** (all fields optional):
```json
{
    "title": "Updated Title",
    "price": 275000,
    "furnished": "semi"
}
```
- **Success Response** (200):
```json
{
    "success": true,
    "message": "Listing updated successfully",
    "data": {
        "id": "PROP1002",
        "title": "Updated Title",
        "price": 275000,
        "furnished": "semi",
        "updated_at": "2024-03-20T11:00:00Z"
    }
}
```
- **Error Responses**:
  - 400: Listing ID is required, invalid request body
  - 403: You don't have permission to update this listing
  - 404: Listing not found

#### Delete Listing
- **URL**: `/listings/:id`
- **Method**: `DELETE`
- **Auth Required**: Yes
- **Success Response** (200):
```json
{
    "success": true,
    "message": "Listing deleted successfully"
}
```
- **Error Responses**:
  - 400: Listing ID is required
  - 403: You don't have permission to delete this listing
  - 404: Listing not found

### Favorites (Requires Authentication)

#### Get User's Favorites
- **URL**: `/favorites`
- **Method**: `GET`
- **Auth Required**: Yes
- **Success Response** (200):
```json
{
    "success": true,
    "data": [
        {
            "id": "PROP1001",
            "title": "Favorite Property",
            "type": "House",
            "price": 350000,
            "state": "Texas",
            "city": "Austin",
            "areaSqFt": 2000,
            "bedrooms": 3,
            "bathrooms": 2,
            "amenities": ["garden", "garage"],
            "furnished": "semi",
            "availableFrom": "2024-05-01",
            "tags": ["family", "quiet"],
            "listingType": "sale",
            "rating": 4.8,
            "isVerified": true
        }
    ]
}
```

#### Add to Favorites
- **URL**: `/favorites/:propertyId`
- **Method**: `POST`
- **Auth Required**: Yes
- **Success Response** (200):
```json
{
    "success": true,
    "message": "Added to favorites successfully"
}
```
- **Error Responses**:
  - 400: Property ID is required, property already in favorites, invalid user ID
  - 404: Property not found

#### Remove from Favorites
- **URL**: `/favorites/:propertyId`
- **Method**: `DELETE`
- **Auth Required**: Yes
- **Success Response** (200):
```json
{
    "success": true,
    "message": "Removed from favorites successfully"
}
```
- **Error Responses**:
  - 400: Property ID is required, invalid user ID
  - 500: Failed to remove from favorites

### Recommendations (Requires Authentication)

#### Send Recommendation
- **URL**: `/recommendations/send`
- **Method**: `POST`
- **Auth Required**: Yes
- **Body**:
```json
{
    "property_id": "507f1f77bcf86cd799439012",
    "recipient_email": "friend@example.com",
    "message": "Check out this amazing property!"
}
```
- **Success Response** (200):
```json
{
    "_id": "507f1f77bcf86cd799439013",
    "property_id": "507f1f77bcf86cd799439012",
    "sender_id": "507f1f77bcf86cd799439011",
    "recipient_email": "friend@example.com",
    "message": "Check out this amazing property!",
    "status": "pending",
    "created_at": "2024-03-20T10:00:00Z",
    "updated_at": "2024-03-20T10:00:00Z"
}
```
- **Error Responses**:
  - 400: Invalid request body, invalid property/user ID
  - 401: Unauthorized

#### Get User Recommendations
- **URL**: `/recommendations`
- **Method**: `GET`
- **Auth Required**: Yes
- **Success Response** (200):
```json
[
    {
        "_id": "507f1f77bcf86cd799439013",
        "property_id": "507f1f77bcf86cd799439012",
        "sender_id": "507f1f77bcf86cd799439011",
        "recipient_email": "friend@example.com",
        "message": "Check out this amazing property!",
        "status": "pending",
        "created_at": "2024-03-20T10:00:00Z",
        "updated_at": "2024-03-20T10:00:00Z"
    }
]
```

## Error Response Format
All endpoints return errors in the following format:
```json
{
    "success": false,
    "message": "Error description"
}
```

## Example API Calls using cURL

### Register User
```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890"
  }'
```

### Login
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Get Properties with Filters
```bash
curl -X GET "http://localhost:3000/api/properties?min_price=200000&max_price=500000&city=austin&bedrooms=2&sort_by=price&sort_order=asc&page=1&limit=20"
```

### Search Properties
```bash
curl -X GET "http://localhost:3000/api/properties/search?q=apartment&page=1&limit=10"
```

### Create a Listing (Authenticated)
```bash
curl -X PUT http://localhost:3000/api/listings \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Beautiful 2BR Apartment",
    "type": "Apartment",
    "price": 250000,
    "state": "California",
    "city": "San Francisco",
    "areaSqFt": 1200,
    "bedrooms": 2,
    "bathrooms": 2,
    "amenities": ["parking", "gym"],
    "furnished": "fully",
    "availableFrom": "2024-04-01",
    "tags": ["modern", "downtown"],
    "listingType": "sale"
  }'
```

### Add to Favorites (Authenticated)
```bash
curl -X POST http://localhost:3000/api/favorites/PROP1001 \
  -H "Authorization: Bearer your_jwt_token"
```

### Send Recommendation (Authenticated)
```bash
curl -X POST http://localhost:3000/api/recommendations/send \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "property_id": "507f1f77bcf86cd799439012",
    "recipient_email": "friend@example.com",
    "message": "Check out this amazing property!"
  }'
```

## Notes
- Property IDs are generated automatically with format "PROP{number}" starting from PROP1000
- User IDs are MongoDB ObjectIDs
- JWT tokens expire after 7 days
- Listings start as unverified (isVerified: false) and with 0 rating
- Property search uses regex matching on title, state, city, type, amenities, and tags fields
- Pagination is available on most listing endpoints with reasonable limits 