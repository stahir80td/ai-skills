using MongoDB.Bson;
using MongoDB.Bson.Serialization.Attributes;

namespace AiPatterns.Domain.Models;

/// <summary>
/// User profile for MongoDB storage - demonstrates document/flexible schema data
/// </summary>
public class UserProfile
{
    [BsonId]
    [BsonRepresentation(BsonType.String)]
    public Guid Id { get; set; }

    [BsonElement("email")]
    public string Email { get; set; } = string.Empty;

    [BsonElement("firstName")]
    public string FirstName { get; set; } = string.Empty;

    [BsonElement("lastName")]
    public string LastName { get; set; } = string.Empty;

    [BsonElement("preferences")]
    public UserPreferences Preferences { get; set; } = new();

    [BsonElement("addresses")]
    public List<Address> Addresses { get; set; } = new();

    [BsonElement("paymentMethods")]
    public List<PaymentMethod> PaymentMethods { get; set; } = new();

    [BsonElement("tags")]
    public List<string> Tags { get; set; } = new();

    [BsonElement("metadata")]
    public Dictionary<string, object> Metadata { get; set; } = new();

    [BsonElement("createdAt")]
    public DateTime CreatedAt { get; set; }

    [BsonElement("updatedAt")]
    public DateTime UpdatedAt { get; set; }

    [BsonElement("lastLoginAt")]
    public DateTime? LastLoginAt { get; set; }

    public static UserProfile Create(string email, string firstName, string lastName)
    {
        return new UserProfile
        {
            Id = Guid.NewGuid(),
            Email = email,
            FirstName = firstName,
            LastName = lastName,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };
    }

    public void UpdatePreferences(UserPreferences newPreferences)
    {
        Preferences = newPreferences;
        UpdatedAt = DateTime.UtcNow;
    }

    public void AddAddress(Address address)
    {
        Addresses.Add(address);
        UpdatedAt = DateTime.UtcNow;
    }

    public void UpdateLastLogin()
    {
        LastLoginAt = DateTime.UtcNow;
        UpdatedAt = DateTime.UtcNow;
    }
}

public class UserPreferences
{
    [BsonElement("theme")]
    public string Theme { get; set; } = "system";

    [BsonElement("language")]
    public string Language { get; set; } = "en";

    [BsonElement("timezone")]
    public string Timezone { get; set; } = "UTC";

    [BsonElement("notifications")]
    public NotificationSettings Notifications { get; set; } = new();

    [BsonElement("privacy")]
    public PrivacySettings Privacy { get; set; } = new();

    [BsonElement("categories")]
    public List<string> FavoriteCategories { get; set; } = new();
}

public class NotificationSettings
{
    [BsonElement("email")]
    public bool EmailEnabled { get; set; } = true;

    [BsonElement("sms")]
    public bool SmsEnabled { get; set; } = false;

    [BsonElement("push")]
    public bool PushEnabled { get; set; } = true;

    [BsonElement("frequency")]
    public string Frequency { get; set; } = "daily";
}

public class PrivacySettings
{
    [BsonElement("profileVisible")]
    public bool ProfileVisible { get; set; } = true;

    [BsonElement("dataSharing")]
    public bool DataSharingEnabled { get; set; } = false;

    [BsonElement("analytics")]
    public bool AnalyticsEnabled { get; set; } = true;
}

public class Address
{
    [BsonElement("type")]
    public string Type { get; set; } = "shipping"; // shipping, billing

    [BsonElement("street")]
    public string Street { get; set; } = string.Empty;

    [BsonElement("city")]
    public string City { get; set; } = string.Empty;

    [BsonElement("state")]
    public string State { get; set; } = string.Empty;

    [BsonElement("zipCode")]
    public string ZipCode { get; set; } = string.Empty;

    [BsonElement("country")]
    public string Country { get; set; } = string.Empty;

    [BsonElement("isDefault")]
    public bool IsDefault { get; set; } = false;
}

public class PaymentMethod
{
    [BsonElement("type")]
    public string Type { get; set; } = string.Empty; // card, paypal, bank

    [BsonElement("provider")]
    public string Provider { get; set; } = string.Empty;

    [BsonElement("lastFour")]
    public string LastFour { get; set; } = string.Empty;

    [BsonElement("expiryMonth")]
    public int? ExpiryMonth { get; set; }

    [BsonElement("expiryYear")]
    public int? ExpiryYear { get; set; }

    [BsonElement("isDefault")]
    public bool IsDefault { get; set; } = false;

    [BsonElement("isActive")]
    public bool IsActive { get; set; } = true;
}