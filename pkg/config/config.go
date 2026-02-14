package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	Username = field.StringField(
		"username",
		field.WithRequired(true),
		field.WithDescription("The DockerHub username used to connect to the DockerHub API."),
		field.WithDisplayName("Username"),
	)
	AccessToken = field.StringField(
		"access-token",
		field.WithDescription("The DockerHub Personal Access Token used to connect to the DockerHub API."),
		field.WithIsSecret(true),
		field.WithDisplayName("Access Token"),
	)
	Password = field.StringField(
		"password",
		field.WithDescription("The DockerHub password used to connect to the DockerHub API."),
		field.WithIsSecret(true),
		field.WithDisplayName("Password"),
	)
	Orgs = field.StringSliceField(
		"orgs",
		field.WithDescription("Limit syncing to specific organizations by providing organization slugs."),
		field.WithDisplayName("Organizations"),
	)
	BaseURL = field.StringField(
		"base-url",
		field.WithDescription("Override the DockerHub API URL (for testing)"),
		field.WithDisplayName("Base URL"),
		field.WithHidden(true),
	)

	// FieldRelationships defines relationships between the fields.
	FieldRelationships = []field.SchemaFieldRelationship{
		field.FieldsMutuallyExclusive(AccessToken, Password),
		field.FieldsAtLeastOneUsed(AccessToken, Password),
	}
)

//go:generate go run ./gen
var Configuration = field.NewConfiguration([]field.SchemaField{
	Username,
	AccessToken,
	Password,
	Orgs,
	BaseURL,
}, field.WithConstraints(FieldRelationships...))

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid.
func ValidateConfig(cfg *Dockerhub) error {
	return nil
}
