package models

import (
	"time"

	"github.com/google/uuid"
)

// UserProfile for MongoDB storage - demonstrates document/flexible schema data
type UserProfile struct {
	ID             uuid.UUID              `json:"id" bson:"_id"`
	Email          string                 `json:"email" bson:"email"`
	FirstName      string                 `json:"firstName" bson:"firstName"`
	LastName       string                 `json:"lastName" bson:"lastName"`
	Preferences    UserPreferences        `json:"preferences" bson:"preferences"`
	Addresses      []Address              `json:"addresses" bson:"addresses"`
	PaymentMethods []PaymentMethod        `json:"paymentMethods" bson:"paymentMethods"`
	Tags           []string               `json:"tags" bson:"tags"`
	Metadata       map[string]interface{} `json:"metadata" bson:"metadata"`
	CreatedAt      time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt" bson:"updatedAt"`
	LastLoginAt    *time.Time             `json:"lastLoginAt,omitempty" bson:"lastLoginAt,omitempty"`
}

// UserPreferences holds user preference settings
type UserPreferences struct {
	Theme              string               `json:"theme" bson:"theme"`
	Language           string               `json:"language" bson:"language"`
	Timezone           string               `json:"timezone" bson:"timezone"`
	Notifications      NotificationSettings `json:"notifications" bson:"notifications"`
	Privacy            PrivacySettings      `json:"privacy" bson:"privacy"`
	FavoriteCategories []string             `json:"favoriteCategories" bson:"favoriteCategories"`
}

// NotificationSettings holds notification preferences
type NotificationSettings struct {
	Email       bool `json:"email" bson:"email"`
	Push        bool `json:"push" bson:"push"`
	SMS         bool `json:"sms" bson:"sms"`
	Marketing   bool `json:"marketing" bson:"marketing"`
	OrderStatus bool `json:"orderStatus" bson:"orderStatus"`
}

// PrivacySettings holds privacy preferences
type PrivacySettings struct {
	ProfilePublic     bool `json:"profilePublic" bson:"profilePublic"`
	ShowOnlineStatus  bool `json:"showOnlineStatus" bson:"showOnlineStatus"`
	AllowDataSharing  bool `json:"allowDataSharing" bson:"allowDataSharing"`
	AllowTracking     bool `json:"allowTracking" bson:"allowTracking"`
}

// Address represents a user address
type Address struct {
	ID         uuid.UUID `json:"id" bson:"id"`
	Label      string    `json:"label" bson:"label"` // home, work, etc.
	Street     string    `json:"street" bson:"street"`
	City       string    `json:"city" bson:"city"`
	State      string    `json:"state" bson:"state"`
	PostalCode string    `json:"postalCode" bson:"postalCode"`
	Country    string    `json:"country" bson:"country"`
	IsDefault  bool      `json:"isDefault" bson:"isDefault"`
}

// PaymentMethod represents a saved payment method
type PaymentMethod struct {
	ID           uuid.UUID `json:"id" bson:"id"`
	Type         string    `json:"type" bson:"type"` // card, paypal, etc.
	Last4        string    `json:"last4" bson:"last4"`
	ExpiryMonth  int       `json:"expiryMonth,omitempty" bson:"expiryMonth,omitempty"`
	ExpiryYear   int       `json:"expiryYear,omitempty" bson:"expiryYear,omitempty"`
	IsDefault    bool      `json:"isDefault" bson:"isDefault"`
	BillingEmail string    `json:"billingEmail,omitempty" bson:"billingEmail,omitempty"`
}

// NewUserProfile creates a new user profile with defaults
func NewUserProfile(email, firstName, lastName string) *UserProfile {
	now := time.Now().UTC()
	return &UserProfile{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Preferences: UserPreferences{
			Theme:    "system",
			Language: "en",
			Timezone: "UTC",
			Notifications: NotificationSettings{
				Email:       true,
				Push:        true,
				SMS:         false,
				Marketing:   false,
				OrderStatus: true,
			},
			Privacy: PrivacySettings{
				ProfilePublic:     false,
				ShowOnlineStatus:  true,
				AllowDataSharing:  false,
				AllowTracking:     false,
			},
			FavoriteCategories: []string{},
		},
		Addresses:      []Address{},
		PaymentMethods: []PaymentMethod{},
		Tags:           []string{},
		Metadata:       make(map[string]interface{}),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdatePreferences updates the user's preferences
func (u *UserProfile) UpdatePreferences(prefs UserPreferences) {
	u.Preferences = prefs
	u.UpdatedAt = time.Now().UTC()
}

// AddAddress adds a new address to the user profile
func (u *UserProfile) AddAddress(addr Address) {
	if addr.ID == uuid.Nil {
		addr.ID = uuid.New()
	}
	u.Addresses = append(u.Addresses, addr)
	u.UpdatedAt = time.Now().UTC()
}

// UpdateLastLogin updates the last login timestamp
func (u *UserProfile) UpdateLastLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// GetFullName returns the user's full name
func (u *UserProfile) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UpdatePreferencesRequest represents the request to update user preferences
type UpdatePreferencesRequest struct {
	Preferences UserPreferences `json:"preferences"`
}
