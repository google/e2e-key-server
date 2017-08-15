// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/keytransparency/cmd/keytransparency-client/grpcc"
	"github.com/google/keytransparency/core/authentication"
	"github.com/google/keytransparency/core/client/kt"

	"github.com/google/trillian"
	"github.com/google/trillian/crypto/keys/der"
	"github.com/google/trillian/crypto/keys/pem"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"

	kpb "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
	gauth "github.com/google/keytransparency/impl/google/authentication"
	spb "github.com/google/keytransparency/impl/proto/keytransparency_v1_service"
	_ "github.com/google/trillian/merkle/coniks"    // Register coniks
	_ "github.com/google/trillian/merkle/objhasher" // Register objhasher
	_ "github.com/spf13/viper/remote"               // Enable remote configs
)

var (
	cfgFile string
	verbose bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "keytransparency-client",
	Short: "A client for interacting with the key transparency server",
	Long: `The key transparency client retrieves and sets keys in the 
key transparency server.  The client verifies all cryptographic proofs the
server provides to ensure that account data is accurate.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			grpcc.Vlog = log.New(os.Stdout, "", log.LstdFlags)
			kt.Vlog = log.New(os.Stdout, "", log.LstdFlags)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.keytransparency.yaml)")

	RootCmd.PersistentFlags().String("kt-url", "35.184.134.53:8080", "URL of Key Transparency server")
	RootCmd.PersistentFlags().String("kt-cert", "genfiles/server.crt", "Path to public key for Key Transparency")
	RootCmd.PersistentFlags().Bool("autoconfig", true, "Fetch config info from the server's /v1/domain/info")
	RootCmd.PersistentFlags().Bool("insecure", false, "Skip TLS checks")

	RootCmd.PersistentFlags().String("vrf", "genfiles/vrf-pubkey.pem", "path to vrf public key")

	RootCmd.PersistentFlags().String("log-key", "genfiles/trillian-log.pem", "Path to public key PEM for Trillian Log server")
	RootCmd.PersistentFlags().String("map-key", "genfiles/trillian-map.pem", "Path to public key PEM for Trillian Map server")

	RootCmd.PersistentFlags().String("client-secret", "", "Path to client_secret.json file for user creds")
	RootCmd.PersistentFlags().String("service-key", "", "Path to service_key.json file for anonymous creds")
	RootCmd.PersistentFlags().String("fake-auth-userid", "", "userid to present to the server as identity for authentication. Only succeeds if fake auth is enabled on the server side.")

	// Global flags for use by subcommands.
	RootCmd.PersistentFlags().DurationP("timeout", "t", 3*time.Minute, "Time to wait before operations timeout")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print in/out and verification steps")
	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		log.Fatalf("%v", err)
	}
}

// initConfig reads in config file and ENV variables if set.
// initConfig is run during a command's preRun().
func initConfig() {
	viper.AutomaticEnv() // Read in environment variables that match.

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Failed reading config file: %v: %v", viper.ConfigFileUsed(), err)
		}
	} else {
		viper.SetConfigName(".keytransparency")
		viper.AddConfigPath("$HOME")
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

}

// getTokenFromWeb uses config to request a Token.  Returns the retrieved Token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// TODO: replace state token with something random to prevent CSRF.
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}

	tok, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func getCreds(ctx context.Context, clientSecretFile string) (credentials.PerRPCCredentials, error) {
	b, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, gauth.RequiredScopes...)
	if err != nil {
		return nil, err
	}

	tok, err := getTokenFromWeb(ctx, config)
	if err != nil {
		return nil, err
	}
	return oauth.NewOauthAccess(tok), nil
}

func getServiceCreds(serviceKeyFile string) (credentials.PerRPCCredentials, error) {
	b, err := ioutil.ReadFile(serviceKeyFile)
	if err != nil {
		return nil, err
	}
	return oauth.NewServiceAccountFromKey(b, gauth.RequiredScopes...)
}

func transportCreds(ktURL string) (credentials.TransportCredentials, error) {
	ktCert := viper.GetString("kt-cert")
	insecure := viper.GetBool("insecure")

	host, _, err := net.SplitHostPort(ktURL)
	if err != nil {
		return nil, err
	}

	switch {
	case insecure: // Impatient insecure.
		return credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		}), nil

	case ktCert != "": // Custom CA Cert.
		return credentials.NewClientTLSFromFile(ktCert, host)

	default: // Use the local set of root certs.
		return credentials.NewClientTLSFromCert(nil, host), nil
	}
}

// userCreds returns PerRPCCredentials. Only one type of credential
// should exist in an RPC call. Fake credentials have the highest priority, followed
// by Client credentials and Service Credentials.
func userCreds(ctx context.Context, useClientSecret bool) (credentials.PerRPCCredentials, error) {
	fakeUserID := viper.GetString("fake-auth-userid")    // Fake user creds.
	clientSecretFile := viper.GetString("client-secret") // Real user creds.
	serviceKeyFile := viper.GetString("service-key")     // Anonymous user creds.
	if !useClientSecret {
		clientSecretFile = ""
	}

	switch {
	case fakeUserID != "":
		return authentication.GetFakeCredential(fakeUserID), nil
	case clientSecretFile != "":
		return getCreds(ctx, clientSecretFile)
	case serviceKeyFile != "":
		return getServiceCreds(serviceKeyFile)
	default:
		return nil, nil
	}
}

func dial(ctx context.Context, ktURL string, useClientSecret bool) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	transportCreds, err := transportCreds(ktURL)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithTransportCredentials(transportCreds))

	userCreds, err := userCreds(ctx, useClientSecret)
	if err != nil {
		return nil, err
	}
	if userCreds != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(userCreds))
	}

	cc, err := grpc.Dial(ktURL, opts...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// GetClient connects to the server and returns a key transparency verification
// client.
func GetClient(useClientSecret bool) (*grpcc.Client, error) {
	ctx := context.Background()
	ktURL := viper.GetString("kt-url")
	cc, err := dial(ctx, ktURL, useClientSecret)
	if err != nil {
		return nil, fmt.Errorf("Error Dialing: %v", err)
	}

	config, err := config(ctx, cc)
	if err != nil {
		return nil, fmt.Errorf("Error reading config: %v", err)
	}

	return grpcc.NewFromConfig(cc, config)
}

// config selects a source for and returns the client configuration.
func config(ctx context.Context, cc *grpc.ClientConn) (*kpb.GetDomainInfoResponse, error) {
	autoConfig := viper.GetBool("autoconfig")
	switch {
	case autoConfig:
		ktClient := spb.NewKeyTransparencyServiceClient(cc)
		return ktClient.GetDomainInfo(ctx, &kpb.GetDomainInfoRequest{})
	default:
		return readConfigFromDisk()
	}
}

func readConfigFromDisk() (*kpb.GetDomainInfoResponse, error) {
	vrfPubFile := viper.GetString("vrf")
	logPEMFile := viper.GetString("log-key")
	mapPEMFile := viper.GetString("map-key")

	// Log PubKey.
	logPubKey, err := pem.ReadPublicKeyFile(logPEMFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open log public key %v: %v", logPEMFile, err)
	}
	logPubPB, err := der.ToPublicProto(logPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize log public key: %v", err)
	}

	// VRF PubKey
	vrfPubKey, err := pem.ReadPublicKeyFile(vrfPubFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %s. %v", vrfPubFile, err)
	}
	vrfPubPB, err := der.ToPublicProto(vrfPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize vrf public key: %v", err)
	}

	// MapPubKey.
	mapPubKey, err := pem.ReadPublicKeyFile(mapPEMFile)
	if err != nil {
		return nil, fmt.Errorf("error reading map public key %v: %v", mapPEMFile, err)
	}
	mapPubPB, err := der.ToPublicProto(mapPubKey)
	if err != nil {
		return nil, fmt.Errorf("error seralizeing map public key: %v", err)
	}

	return &kpb.GetDomainInfoResponse{
		Log: &trillian.Tree{
			HashStrategy: trillian.HashStrategy_OBJECT_RFC6962_SHA256,
			PublicKey:    logPubPB,
		},
		Map: &trillian.Tree{
			HashStrategy: trillian.HashStrategy_CONIKS_SHA512_256,
			PublicKey:    mapPubPB,
		},
		Vrf: vrfPubPB,
	}, nil
}
