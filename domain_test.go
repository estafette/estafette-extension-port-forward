package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDefaults(t *testing.T) {

	t.Run("SetsLocalPortToServicePortIfEmpty", func(t *testing.T) {

		releaseTargetName := ""

		params := params{
			ServicePort: "443",
			LocalPort:   "",
		}

		// act
		params.SetDefaults(releaseTargetName)

		assert.Equal(t, "443", params.LocalPort)
	})

	t.Run("KeepsLocalPortIfSet", func(t *testing.T) {

		releaseTargetName := ""

		params := params{
			ServicePort: "443",
			LocalPort:   "8443",
		}

		// act
		params.SetDefaults(releaseTargetName)

		assert.Equal(t, "8443", params.LocalPort)
	})

	t.Run("SetsCredentialsToReleaseTargetNamePrefixedWithGKEIfEmpty", func(t *testing.T) {

		releaseTargetName := "development"

		params := params{
			Credentials: "",
		}

		// act
		params.SetDefaults(releaseTargetName)

		assert.Equal(t, "gke-development", params.Credentials)
	})

	t.Run("KeepsCredentialsIfSet", func(t *testing.T) {

		releaseTargetName := "development"

		params := params{
			Credentials: "gke-staging",
		}

		// act
		params.SetDefaults(releaseTargetName)

		assert.Equal(t, "gke-staging", params.Credentials)
	})
}
