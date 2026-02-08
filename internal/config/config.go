package config

import (
	"time"
)

// AuthType represents the authentication method
type AuthType string

const (
	AuthTypeToken AuthType = "token"
	AuthTypeOAuth AuthType = "oauth"
)

// Instance represents a Zendesk instance configuration
type Instance struct {
	Name           string `ini:"-"`
	Subdomain      string `ini:"subdomain"`
	AuthType       AuthType  `ini:"auth_type"`
	Email          string `ini:"email,omitempty"`
	APIToken       string `ini:"api_token,omitempty"`
	OAuthClientID  string `ini:"oauth_client_id,omitempty"`
	OAuthSecret    string `ini:"oauth_secret,omitempty"`
	OAuthToken     string `ini:"oauth_token,omitempty"`
	OAuthRefresh   string `ini:"oauth_refresh,omitempty"`
	OAuthExpiry    string `ini:"oauth_expiry,omitempty"` // Store as RFC3339 string
}

// GetOAuthExpiry returns the OAuth expiry as a time.Time
func (i *Instance) GetOAuthExpiry() (time.Time, error) {
	if i.OAuthExpiry == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, i.OAuthExpiry)
}

// SetOAuthExpiry sets the OAuth expiry from a time.Time
func (i *Instance) SetOAuthExpiry(t time.Time) {
	i.OAuthExpiry = t.Format(time.RFC3339)
}

// Config represents the entire CLI configuration
type Config struct {
	Current   string               `ini:"-"`
	Instances map[string]*Instance `ini:"-"`
}

// NewConfig creates a new empty configuration
func NewConfig() *Config {
	return &Config{
		Instances: make(map[string]*Instance),
	}
}

// GetCurrentInstance returns the currently active instance
func (c *Config) GetCurrentInstance() (*Instance, error) {
	if c.Current == "" {
		return nil, ErrNoCurrentInstance
	}

	instance, ok := c.Instances[c.Current]
	if !ok {
		return nil, ErrInstanceNotFound
	}

	return instance, nil
}

// AddInstance adds a new instance to the configuration
func (c *Config) AddInstance(instance *Instance) error {
	return c.AddInstanceWithSwitch(instance, false)
}

// AddInstanceAndSwitch adds a new instance and makes it current
func (c *Config) AddInstanceAndSwitch(instance *Instance) error {
	return c.AddInstanceWithSwitch(instance, true)
}

// AddInstanceWithSwitch adds a new instance with optional switching
func (c *Config) AddInstanceWithSwitch(instance *Instance, makeCurrent bool) error {
	if instance.Name == "" {
		return ErrInvalidInstanceName
	}

	if _, exists := c.Instances[instance.Name]; exists {
		return ErrInstanceExists
	}

	c.Instances[instance.Name] = instance

	// Make it current if first instance OR if explicitly requested
	if c.Current == "" || makeCurrent {
		c.Current = instance.Name
	}

	return nil
}

// RemoveInstance removes an instance from the configuration
func (c *Config) RemoveInstance(name string) error {
	if _, ok := c.Instances[name]; !ok {
		return ErrInstanceNotFound
	}

	delete(c.Instances, name)

	// If we removed the current instance, switch to another one
	if c.Current == name {
		if len(c.Instances) > 0 {
			// Pick the first available instance
			for n := range c.Instances {
				c.Current = n
				break
			}
		} else {
			c.Current = ""
		}
	}

	return nil
}

// SwitchInstance changes the current active instance
func (c *Config) SwitchInstance(name string) error {
	if _, ok := c.Instances[name]; !ok {
		return ErrInstanceNotFound
	}

	c.Current = name
	return nil
}
