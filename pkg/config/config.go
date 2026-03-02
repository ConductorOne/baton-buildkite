package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	accessTokenField = field.StringField(
		"api-token",
		field.WithDisplayName("API token"),
		field.WithDescription("API token to authenticate buildkite"),
		field.WithIsSecret(true),
		field.WithRequired(true),
	)

	orgField = field.StringField(
		"organization",
		field.WithDisplayName("Organization"),
		field.WithDescription("Buildkite Organization"),
		field.WithRequired(true),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	[]field.SchemaField{
		accessTokenField,
		orgField,
	},
	field.WithConnectorDisplayName("Buildkite v2"),
	field.WithHelpUrl("/docs/baton/buildkite-v2"),
	field.WithIconUrl("/static/app-icons/buildkite.svg"),
)
