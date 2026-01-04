using Core.Infrastructure.SqlServer;
using Core.Infrastructure.Kafka;
using Core.Infrastructure.MongoDB;
using Core.Infrastructure.KeyVault;
using Core.Infrastructure;
using Core.Logger;
using Core.Config;
using Microsoft.Extensions.Logging;
using Xunit.Abstractions;

namespace Core.Tests
{
    public class EnhancedInfrastructureTests
    {
        private readonly ITestOutputHelper _output;
        private readonly ServiceLogger _logger;

        public EnhancedInfrastructureTests(ITestOutputHelper output)
        {
            _output = output;
            var loggerConfig = new LoggerConfig
            {
                ServiceName = "test",
                Environment = "test",
                LogLevel = "Debug",
                OutputFormat = "json"
            };
            _logger = new ServiceLogger(loggerConfig);
        }

        #region SqlServerClient Tests

        [Fact]
        public void SqlServerConfig_DefaultValues()
        {
            // Arrange & Act
            var config = new SqlServerConfig();

            // Assert
            Assert.Equal("", config.ConnectionString);
            Assert.Equal(60, config.CommandTimeout);
            Assert.True(config.EnableConnectionPooling);
            Assert.Equal(25, config.MaxPoolSize);
            Assert.Equal(5, config.MaxIdleConnections);
            Assert.Equal(5, config.ConnectionLifetimeMinutes);
            Assert.Equal(5, config.HealthCheckTimeoutSeconds);
        }

        [Fact]
        public void SqlServerConfig_Validate_RequiresConnectionString()
        {
            // Arrange
            var config = new SqlServerConfig { ConnectionString = "" };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("ConnectionString is required", result.Errors);
        }

        [Fact]
        public void SqlServerConfig_Validate_RequiresMinimumTimeout()
        {
            // Arrange
            var config = new SqlServerConfig 
            { 
                ConnectionString = "Server=test",
                CommandTimeout = 30 
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("CommandTimeout must be at least 60 seconds", result.Errors);
        }

        [Fact]
        public void SqlServerConfig_Validate_ValidConfiguration()
        {
            // Arrange
            var config = new SqlServerConfig 
            { 
                ConnectionString = "Server=test;Database=test",
                CommandTimeout = 120,
                MaxPoolSize = 50,
                HealthCheckTimeoutSeconds = 10
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.True(result.IsValid);
            Assert.Empty(result.Errors);
        }

        [Fact]
        public void SqlServerClient_ThrowsOnInvalidConfig()
        {
            // Arrange
            var config = new SqlServerConfig { ConnectionString = "" };

            // Act & Assert
            var exception = Assert.Throws<InvalidOperationException>(() => 
                new SqlServerClient(config, _logger));
            
            Assert.Contains("Invalid configuration", exception.Message);
        }

        #endregion

        #region KafkaClient Tests

        [Fact]
        public void KafkaConfig_DefaultValues()
        {
            // Arrange & Act
            var config = new KafkaConfig();

            // Assert
            Assert.Equal("localhost:9092", config.BootstrapServers);
            Assert.Equal(Confluent.Kafka.SecurityProtocol.Plaintext, config.SecurityProtocol);
            Assert.True(config.EnableIdempotence);
            Assert.Equal(30000, config.MessageTimeoutMs);
            Assert.Equal(60000, config.RequestTimeoutMs);
            Assert.Equal(60000, config.SessionTimeoutMs);
            Assert.Equal(20000, config.HeartbeatIntervalMs);
            Assert.Equal(10, config.HealthCheckTimeoutSeconds);
        }

        [Fact]
        public void KafkaConfig_Validate_RequiresBootstrapServers()
        {
            // Arrange
            var config = new KafkaConfig { BootstrapServers = "" };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("BootstrapServers is required", result.Errors);
        }

        [Fact]
        public void KafkaConfig_Validate_RequiresMinimumTimeouts()
        {
            // Arrange
            var config = new KafkaConfig 
            { 
                BootstrapServers = "localhost:9092",
                RequestTimeoutMs = 30000,
                SessionTimeoutMs = 30000
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("RequestTimeoutMs must be at least 60 seconds", result.Errors);
            Assert.Contains("SessionTimeoutMs must be at least 60 seconds", result.Errors);
        }

        [Fact]
        public void KafkaConfig_Validate_ValidConfiguration()
        {
            // Arrange
            var config = new KafkaConfig 
            { 
                BootstrapServers = "localhost:9092",
                RequestTimeoutMs = 90000,
                SessionTimeoutMs = 75000,
                HealthCheckTimeoutSeconds = 15
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.True(result.IsValid);
            Assert.Empty(result.Errors);
        }

        [Fact]
        public void KafkaProducer_ThrowsOnInvalidConfig()
        {
            // Arrange
            var config = new KafkaConfig { BootstrapServers = "" };

            // Act & Assert
            var exception = Assert.Throws<InvalidOperationException>(() => 
                new KafkaProducer(config, _logger));
            
            Assert.Contains("Invalid configuration", exception.Message);
        }

        #endregion

        #region MongoClient Tests

        [Fact]
        public void MongoConfig_DefaultValues()
        {
            // Arrange & Act
            var config = new MongoConfig();

            // Assert
            Assert.Equal("mongodb://localhost:27017", config.ConnectionString);
            Assert.Equal("myapp", config.DatabaseName);
            Assert.Equal(TimeSpan.FromSeconds(60), config.ConnectionTimeout);
            Assert.Equal(TimeSpan.FromSeconds(60), config.ServerSelectionTimeout);
            Assert.Equal(TimeSpan.FromSeconds(60), config.SocketTimeout);
            Assert.Equal(100, config.MaxConnectionPoolSize);
            Assert.Equal(10, config.HealthCheckTimeoutSeconds);
        }

        [Fact]
        public void MongoConfig_Validate_RequiresConnectionString()
        {
            // Arrange
            var config = new MongoConfig { ConnectionString = "", DatabaseName = "test" };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("ConnectionString is required", result.Errors);
        }

        [Fact]
        public void MongoConfig_Validate_RequiresDatabaseName()
        {
            // Arrange
            var config = new MongoConfig { ConnectionString = "mongodb://test", DatabaseName = "" };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("DatabaseName is required", result.Errors);
        }

        [Fact]
        public void MongoConfig_Validate_RequiresMinimumTimeouts()
        {
            // Arrange
            var config = new MongoConfig 
            { 
                ConnectionString = "mongodb://test",
                DatabaseName = "test",
                ConnectionTimeout = TimeSpan.FromSeconds(30),
                ServerSelectionTimeout = TimeSpan.FromSeconds(30),
                SocketTimeout = TimeSpan.FromSeconds(30)
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("ConnectionTimeout must be at least 60 seconds", result.Errors);
            Assert.Contains("ServerSelectionTimeout must be at least 60 seconds", result.Errors);
            Assert.Contains("SocketTimeout must be at least 60 seconds", result.Errors);
        }

        [Fact]
        public void MongoConfig_Validate_ValidConfiguration()
        {
            // Arrange
            var config = new MongoConfig 
            { 
                ConnectionString = "mongodb://localhost:27017",
                DatabaseName = "testdb",
                ConnectionTimeout = TimeSpan.FromSeconds(120),
                ServerSelectionTimeout = TimeSpan.FromSeconds(90),
                SocketTimeout = TimeSpan.FromSeconds(75),
                MaxConnectionPoolSize = 200,
                HealthCheckTimeoutSeconds = 8
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.True(result.IsValid);
            Assert.Empty(result.Errors);
        }

        [Fact]
        public void MongoClient_ThrowsOnInvalidConfig()
        {
            // Arrange
            var config = new MongoConfig { ConnectionString = "", DatabaseName = "test" };

            // Act & Assert
            var exception = Assert.Throws<InvalidOperationException>(() => 
                new MongoClient(config, _logger));
            
            Assert.Contains("Invalid configuration", exception.Message);
        }

        #endregion

        #region KeyVaultClient Tests

        [Fact]
        public void KeyVaultConfig_DefaultValues()
        {
            // Arrange & Act
            var config = new KeyVaultConfig();

            // Assert
            Assert.Equal("", config.VaultUrl);
            Assert.Equal(KeyVaultAuthMethod.ManagedIdentity, config.AuthMethod);
            Assert.Equal(TimeSpan.FromMinutes(5), config.CacheTtl);
            Assert.True(config.EnableCaching);
            Assert.Equal(10, config.HealthCheckTimeoutSeconds);
        }

        [Fact]
        public void KeyVaultConfig_Validate_RequiresVaultUrl()
        {
            // Arrange
            var config = new KeyVaultConfig { VaultUrl = "" };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("VaultUrl is required", result.Errors);
        }

        [Fact]
        public void KeyVaultConfig_Validate_ServicePrincipalRequiresCredentials()
        {
            // Arrange
            var config = new KeyVaultConfig 
            { 
                VaultUrl = "https://test.vault.azure.net/",
                AuthMethod = KeyVaultAuthMethod.ServicePrincipal,
                ClientId = "",
                ClientSecret = "",
                TenantId = ""
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("ClientId is required for ServicePrincipal authentication", result.Errors);
            Assert.Contains("ClientSecret is required for ServicePrincipal authentication", result.Errors);
            Assert.Contains("TenantId is required for ServicePrincipal authentication", result.Errors);
        }

        [Fact]
        public void KeyVaultConfig_Validate_ValidManagedIdentityConfiguration()
        {
            // Arrange
            var config = new KeyVaultConfig 
            { 
                VaultUrl = "https://test.vault.azure.net/",
                AuthMethod = KeyVaultAuthMethod.ManagedIdentity,
                CacheTtl = TimeSpan.FromMinutes(10),
                EnableCaching = true,
                HealthCheckTimeoutSeconds = 15
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.True(result.IsValid);
            Assert.Empty(result.Errors);
        }

        [Fact]
        public void KeyVaultClient_ThrowsOnInvalidConfig()
        {
            // Arrange
            var config = new KeyVaultConfig { VaultUrl = "" };

            // Act & Assert
            var exception = Assert.Throws<InvalidOperationException>(() => 
                new KeyVaultClient(config, _logger));
            
            Assert.Contains("Invalid configuration", exception.Message);
        }

        #endregion

        #region ScyllaDBClient Tests

        [Fact]
        public void ScyllaConfig_DefaultValues()
        {
            // Arrange & Act
            var config = new ScyllaConfig { Hosts = new[] { "localhost" } };

            // Assert
            Assert.Single(config.Hosts);
            Assert.Equal("localhost", config.Hosts[0]);
            Assert.Equal(9042, config.Port);
            Assert.Equal(60000, config.ConnectionTimeoutMs);
            Assert.Equal(60000, config.ReadTimeoutMs);
            Assert.Equal(60000, config.WriteTimeoutMs);
            Assert.Equal(8, config.MaxConnectionsPerHost);
            Assert.Equal(128, config.MaxRequestsPerConnection);
            Assert.True(config.EnableCompression);
            Assert.Equal(10, config.HealthCheckTimeoutSeconds);
        }

        [Fact]
        public void ScyllaConfig_Validate_RequiresHosts()
        {
            // Arrange
            var config = new ScyllaConfig { Hosts = new string[0] };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("At least one host is required", result.Errors);
        }

        [Fact]
        public void ScyllaConfig_Validate_RequiresValidPort()
        {
            // Arrange
            var config = new ScyllaConfig 
            { 
                Hosts = new[] { "localhost" },
                Port = -1 
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("Port must be between 1 and 65535", result.Errors);
        }

        [Fact]
        public void ScyllaConfig_Validate_RequiresMinimumTimeouts()
        {
            // Arrange
            var config = new ScyllaConfig 
            { 
                Hosts = new[] { "localhost" },
                ConnectionTimeoutMs = 30000,
                ReadTimeoutMs = 30000,
                WriteTimeoutMs = 30000
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.False(result.IsValid);
            Assert.Contains("ConnectionTimeoutMs must be at least 60 seconds", result.Errors);
            Assert.Contains("ReadTimeoutMs must be at least 60 seconds", result.Errors);
            Assert.Contains("WriteTimeoutMs must be at least 60 seconds", result.Errors);
        }

        [Fact]
        public void ScyllaConfig_Validate_ValidConfiguration()
        {
            // Arrange
            var config = new ScyllaConfig 
            { 
                Hosts = new[] { "localhost", "host2" },
                Port = 9042,
                Keyspace = "testkeyspace",
                ConnectionTimeoutMs = 90000,
                ReadTimeoutMs = 120000,
                WriteTimeoutMs = 75000,
                MaxConnectionsPerHost = 16,
                MaxRequestsPerConnection = 256,
                EnableCompression = false,
                HealthCheckTimeoutSeconds = 20
            };

            // Act
            var result = config.Validate();

            // Assert
            Assert.True(result.IsValid);
            Assert.Empty(result.Errors);
        }

        [Fact]
        public void ScyllaDBClient_ThrowsOnInvalidConfig()
        {
            // Arrange
            var config = new ScyllaConfig { Hosts = new string[0] };

            // Act & Assert
            var exception = Assert.Throws<InvalidOperationException>(() => 
                new ScyllaDBClient(config, _logger));
            
            Assert.Contains("Invalid configuration", exception.Message);
        }

        [Fact]
        public void ScyllaDBClient_ValidConfig_InitializesSuccessfully()
        {
            // Arrange
            var config = new ScyllaConfig 
            { 
                Hosts = new[] { "localhost" },
                Keyspace = "test",
                ConnectionTimeoutMs = 75000
            };

            // Act & Assert - should not throw
            var client = new ScyllaDBClient(config, _logger);
            Assert.NotNull(client);
            
            // Cleanup
            client.Dispose();
        }

        #endregion

        #region ConnectionDiagnostics Tests

        [Fact]
        public void ConnectionDiagnostics_DefaultValues()
        {
            // Arrange & Act
            var diagnostics = new ConnectionDiagnostics();

            // Assert
            Assert.False(diagnostics.IsHealthy);
            Assert.Equal("", diagnostics.DatabaseName);
            Assert.Equal("", diagnostics.ServerVersion);
            Assert.Equal(TimeSpan.Zero, diagnostics.ConnectionTime);
            Assert.Null(diagnostics.ErrorMessage);
        }

        [Fact]
        public void ConnectionDiagnostics_HealthyState()
        {
            // Arrange & Act
            var diagnostics = new ConnectionDiagnostics
            {
                IsHealthy = true,
                DatabaseName = "TestDB",
                ServerVersion = "5.0.0",
                ConnectionTime = TimeSpan.FromMilliseconds(150)
            };

            // Assert
            Assert.True(diagnostics.IsHealthy);
            Assert.Equal("TestDB", diagnostics.DatabaseName);
            Assert.Equal("5.0.0", diagnostics.ServerVersion);
            Assert.Equal(TimeSpan.FromMilliseconds(150), diagnostics.ConnectionTime);
            Assert.Null(diagnostics.ErrorMessage);
        }

        [Fact]
        public void ConnectionDiagnostics_UnhealthyState()
        {
            // Arrange & Act
            var diagnostics = new ConnectionDiagnostics
            {
                IsHealthy = false,
                ConnectionTime = TimeSpan.FromSeconds(5),
                ErrorMessage = "Connection timeout"
            };

            // Assert
            Assert.False(diagnostics.IsHealthy);
            Assert.Equal(TimeSpan.FromSeconds(5), diagnostics.ConnectionTime);
            Assert.Equal("Connection timeout", diagnostics.ErrorMessage);
        }

        #endregion

        #region Error Code Tests

        [Fact]
        public void InfrastructureClients_UseStandardizedErrorCodes()
        {
            // Test SqlServer client
            var sqlConfig = new SqlServerConfig { ConnectionString = "" };
            var sqlException = Assert.Throws<InvalidOperationException>(() => 
                new SqlServerClient(sqlConfig, _logger));
            Assert.Contains("Invalid configuration", sqlException.Message);
            _output.WriteLine("✓ SqlServer client throws InvalidOperationException on invalid config");

            // Test Kafka client
            var kafkaConfig = new KafkaConfig { BootstrapServers = "" };
            var kafkaException = Assert.Throws<InvalidOperationException>(() => 
                new KafkaProducer(kafkaConfig, _logger));
            Assert.Contains("Invalid configuration", kafkaException.Message);
            _output.WriteLine("✓ Kafka client throws InvalidOperationException on invalid config");

            // Test MongoDB client
            var mongoConfig = new MongoConfig { ConnectionString = "", DatabaseName = "test" };
            var mongoException = Assert.Throws<InvalidOperationException>(() => 
                new MongoClient(mongoConfig, _logger));
            Assert.Contains("Invalid configuration", mongoException.Message);
            _output.WriteLine("✓ MongoDB client throws InvalidOperationException on invalid config");

            // Test ScyllaDB client
            var scyllaConfig = new ScyllaConfig { Hosts = new string[0] };
            var scyllaException = Assert.Throws<InvalidOperationException>(() => 
                new ScyllaDBClient(scyllaConfig, _logger));
            Assert.Contains("Invalid configuration", scyllaException.Message);
            _output.WriteLine("✓ ScyllaDB client throws InvalidOperationException on invalid config");
        }

        #endregion
    }
}