package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime"

	"github.com/alecthomas/kingpin"
	foundation "github.com/estafette/estafette-foundation"
	zerolog "github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

var (
	appgroup  string
	app       string
	version   string
	branch    string
	revision  string
	buildDate string
	goVersion = runtime.Version()
)

var (
	paramsYAML = kingpin.Flag("params-yaml", "Extension parameters, created from custom properties.").Envar("ESTAFETTE_EXTENSION_CUSTOM_PROPERTIES_YAML").Required().String()

	releaseTargetName = kingpin.Flag("release-target-name", "Name of the release target, which is used by convention to resolve the credentials.").Envar("ESTAFETTE_RELEASE_NAME").String()
	credentialsJSON   = kingpin.Flag("credentials", "GKE credentials configured at service level, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_KUBERNETES_ENGINE").String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// init log format from envvar ESTAFETTE_LOG_FORMAT
	foundation.InitLoggingFromEnv(appgroup, app, version, branch, revision, buildDate)

	zerolog.Info().Msg("Unmarshalling parameters / custom properties...")
	var params params
	err := yaml.Unmarshal([]byte(*paramsYAML), &params)
	if err != nil {
		log.Fatal("Failed unmarshalling parameters: ", err)
	}

	zerolog.Info().Msg("Setting defaults for parameters that are not set in the manifest...")
	params.SetDefaults(*releaseTargetName)

	if *credentialsJSON == "" {
		log.Fatal("Credentials of type kubernetes-engine are not injected; configure this extension as trusted and inject credentials of type kubernetes-engine")
	}

	log.Printf("Unmarshalling injected credentials...")
	var credentials []GKECredentials
	err = json.Unmarshal([]byte(*credentialsJSON), &credentials)
	if err != nil {
		log.Fatal("Failed unmarshalling injected credentials: ", err)
	}

	log.Printf("Checking if credential %v exists...", params.Credentials)
	credential := GetCredentialsByName(credentials, params.Credentials)
	if credential == nil {
		log.Fatalf("Credential with name %v does not exist.", params.Credentials)
	}

	log.Printf("Retrieving service account email from credentials...")
	var keyFileMap map[string]interface{}
	err = json.Unmarshal([]byte(credential.AdditionalProperties.ServiceAccountKeyfile), &keyFileMap)
	if err != nil {
		log.Fatal("Failed unmarshalling service account keyfile: ", err)
	}
	var saClientEmail string
	if saClientEmailIntfc, ok := keyFileMap["client_email"]; !ok {
		log.Fatal("Field client_email missing from service account keyfile")
	} else {
		if t, aok := saClientEmailIntfc.(string); !aok {
			log.Fatal("Field client_email not of type string")
		} else {
			saClientEmail = t
		}
	}

	log.Printf("Storing gke credential %v on disk...", params.Credentials)
	err = ioutil.WriteFile("/key-file.json", []byte(credential.AdditionalProperties.ServiceAccountKeyfile), 0600)
	if err != nil {
		log.Fatal("Failed writing service account keyfile: ", err)
	}

	log.Printf("Authenticating to google cloud")
	foundation.RunCommandWithArgs("gcloud", []string{"auth", "activate-service-account", saClientEmail, "--key-file", "/key-file.json"})

	log.Printf("Setting gcloud account to %v", saClientEmail)
	foundation.RunCommandWithArgs("gcloud", []string{"config", "set", "account", saClientEmail})

	log.Printf("Setting gcloud project")
	foundation.RunCommandWithArgs("gcloud", []string{"config", "set", "project", credential.AdditionalProperties.Project})

	log.Printf("Getting gke credentials for cluster %v", credential.AdditionalProperties.Cluster)
	clustersGetCredentialsArsgs := []string{"container", "clusters", "get-credentials", credential.AdditionalProperties.Cluster}
	if credential.AdditionalProperties.Zone != "" {
		clustersGetCredentialsArsgs = append(clustersGetCredentialsArsgs, "--zone", credential.AdditionalProperties.Zone)
	} else if credential.AdditionalProperties.Region != "" {
		clustersGetCredentialsArsgs = append(clustersGetCredentialsArsgs, "--region", credential.AdditionalProperties.Region)
	} else {
		log.Fatal("Credentials have no zone or region; at least one of them has to be defined")
	}
	foundation.RunCommandWithArgs("gcloud", clustersGetCredentialsArsgs)

	zerolog.Info().Msgf("Forwarding port %v to port %v on service %v in namespace %v...", params.LocalPort, params.ServicePort, params.Service, params.Namespace)
	foundation.RunCommand("kubectl port-forward service/%v %v:%v --address=0.0.0.0 -n %v", params.Service, params.LocalPort, params.ServicePort, params.Namespace)

	// wait for SIGTERM
	sigs, wg := foundation.InitGracefulShutdownHandling()
	foundation.HandleGracefulShutdown(sigs, wg)
}
