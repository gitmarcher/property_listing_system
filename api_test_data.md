# Property Listing API - JSON Request Bodies for Testing

## Authentication Endpoints

### 1. Register User 1 (John Doe)
POST /api/auth/register
```json
{
    "email": "john.doe@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+91 0123456789"
}
```

### 2. Register User 2 (Jane Smith) 
POST /api/auth/register
```json
{
    "email": "jane.smith@example.com",
    "password": "password456",
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "+91 1234567890"
}
```

### 3. Login User 1 (John Doe)
POST /api/auth/login
```json
{
    "email": "john.doe@example.com",
    "password": "password123"
}
```

### 4. Login User 2 (Jane Smith)
POST /api/auth/login
```json
{
    "email": "jane.smith@example.com",
    "password": "password456"
}
```

## Listing Management Endpoints

### 5. Create Property Listing - Apartment
PUT /api/listings
```json
{
    "title": "Modern 2BR Downtown Apartment",
    "type": "Apartment",
    "price": 2500,
    "state": "Banglore",
    "city": "Karnataka",
    "areaSqFt": 1200,
    "bedrooms": 2,
    "bathrooms": 2,
    "amenities": ["parking", "gym", "pool", "wifi"],
    "furnished": "fully",
    "availableFrom": "2024-04-01",
    "tags": ["modern", "metro", "luxury"],
    "listingType": "rent"
}
```


### 8. Update Property Listing (Partial Update)
PATCH /api/listings/:id
```json
{
    "price": 2700,
    "furnished": "semi",
    "tags": ["modern", "metro", "luxury", "updated"]
}
```

### 9. Update Property Listing (Multiple Fields)
PATCH /api/listings/:id
```json
{
    "title": "Luxury 2BR Downtown Apartment - Updated",
    "price": 2800,
    "amenities": ["parking", "gym", "pool", "wifi", "concierge"],
    "availableFrom": "2024-04-15"
}
```

## Recommendation Endpoints

### 10. Send Recommendation - John to Jane
POST /api/recommendations/send
```json
{
    "property_id": "PROP1001",
    "recipient_email": "jane.smith@example.com",
    "message": "Hi Jane! I found this amazing downtown apartment that I think you'd love. It has all the amenities you were looking for and it's in a great location!"
}
```

### 11. Send Recommendation - Jane to John
POST /api/recommendations/send
```json
{
    "property_id": "PROP1002", 
    "recipient_email": "john.doe@example.com",
    "message": "Hey John! Check out this beautiful family home in Austin. I remember you mentioned wanting to move to Texas - this could be perfect for you!"
}
```

### 12. Send Recommendation - Simple Message
POST /api/recommendations/send
```json
{
    "property_id": "PROP1003",
    "recipient_email": "jane.smith@example.com",
    "message": "Perfect studio apartment for city living!"
}
```

### 13. Send Recommendation - No Message (Optional)
POST /api/recommendations/send
```json
{
    "property_id": "PROP1001",
    "recipient_email": "john.doe@example.com",
    "message": ""
}
```

## Additional Test Users for Extended Testing

### 14. Register User 3 (Mike Johnson)
POST /api/auth/register
```json
{
    "email": "mike.johnson@example.com",
    "password": "securepass789",
    "first_name": "Mike",
    "last_name": "Johnson",
    "phone": "+1555123456"
}
```

### 15. Register User 4 (Sarah Wilson)
POST /api/auth/register
```json
{
    "email": "sarah.wilson@example.com",
    "password": "mypassword321",
    "first_name": "Sarah",
    "last_name": "Wilson",
    "phone": "+1444987654"
}
```

## Property Listing Examples for Different Categories

### 16. Luxury Condo
PUT /api/listings
```json
{
    "title": "Luxury Oceanview Condo",
    "type": "Condo",
    "price": 750000,
    "state": "Florida",
    "city": "Miami",
    "areaSqFt": 1800,
    "bedrooms": 3,
    "bathrooms": 2,
    "amenities": ["ocean_view", "pool", "gym", "valet_parking", "spa"],
    "furnished": "fully",
    "availableFrom": "2024-06-01",
    "tags": ["luxury", "oceanview", "resort_style"],
    "listingType": "sale"
}
```

### 17. Affordable Townhouse
PUT /api/listings
```json
{
    "title": "Spacious Townhouse with Yard",
    "type": "Townhouse",
    "price": 1400,
    "state": "Ohio",
    "city": "Columbus",
    "areaSqFt": 1600,
    "bedrooms": 3,
    "bathrooms": 2,
    "amenities": ["yard", "garage", "basement"],
    "furnished": "unfurnished",
    "availableFrom": "2024-04-01",
    "tags": ["affordable", "family_friendly", "quiet"],
    "listingType": "rent"
}
```

### 18. Commercial Property
PUT /api/listings
```json
{
    "title": "Prime Commercial Office Space",
    "type": "Commercial",
    "price": 5000,
    "state": "Illinois",
    "city": "Chicago",
    "areaSqFt": 3000,
    "bedrooms": 0,
    "bathrooms": 2,
    "amenities": ["parking", "elevator", "conference_rooms", "wifi"],
    "furnished": "semi",
    "availableFrom": "2024-05-01",
    "tags": ["commercial", "office", "downtown"],
    "listingType": "rent"
}
```

## Notes for Testing:

1. **User Registration Flow:**
   - Register John Doe first
   - Register Jane Smith second
   - Use these users to test recommendation functionality

2. **Property Creation:**
   - Create properties using John's account (after login)
   - Note the property IDs returned (PROP1001, PROP1002, etc.)
   - Use these property IDs in recommendation requests

3. **Recommendation Testing:**
   - Login as John and send recommendation to Jane
   - Login as Jane and send recommendation to John
   - Check sent/received recommendations for both users

4. **Authentication Headers:**
   - After login, use the returned JWT token in Authorization header
   - Format: "Authorization: Bearer <jwt_token>"

5. **Property IDs:**
   - Property IDs are auto-generated starting from PROP1000
   - Use actual returned IDs from CREATE responses in recommendations

6. **Required vs Optional Fields:**
   - All listing fields in examples 5-7 are required
   - Update requests (examples 8-9) can include any subset of fields
   - Recommendation message is optional (can be empty string) 