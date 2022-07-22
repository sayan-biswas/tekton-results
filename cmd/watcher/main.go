/*
Copyright 2020 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/injection"
	"log"
	"os"
	"strings"
	"time"

	cred "github.com/tektoncd/results/pkg/watcher/grpc"
	"github.com/tektoncd/results/pkg/watcher/reconciler"
	"github.com/tektoncd/results/pkg/watcher/reconciler/pipelinerun"
	"github.com/tektoncd/results/pkg/watcher/reconciler/taskrun"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	_ "knative.dev/pkg/system/testing"
)

const (
	// Service Account token path. See https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#accessing-the-api-from-a-pod
	podTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

var (
	apiServer               = flag.String("api-server", "localhost:50051", "Address of API server to report to")
	authMode                = flag.String("auth-mode", "token", "Authentication mode to use when making requests. If not set, no additional credentials will be used in the request. Valid values: [google]")
	authToken               = flag.String("auth-token", "", "Authentication token to use in requests. If not specified, on-cluster configuration is assumed.")
	authServiceAccount      = flag.String("auth-service-account", "tekton-results:tekton-results", "Service account to use for authentication. Format: [namespace:serviceaccount].")
	crdUpdate               = flag.Bool("crd-update", false, "Enable/Disables Tekton CRD annotation update on reconcile.")
	completedRunGracePeriod = flag.Duration("completed_run_grace_period", 0, "Grace period duration before Runs should be deleted. If 0, Runs will not be deleted. If < 0, Runs will be deleted immediately.")
	tlsCert                 = flag.String("tls-cert", "/etc/tls/tls.crt", "Certificate to connect the API server")
	tlsOverride             = flag.String("tls-override", "", "TLS server name override")
)

func main() {
	flag.Parse()
	// TODO: Enable leader election.
	ctx := sharedmain.WithHADisabled(signals.NewContext())

	conn, err := connectToAPIServer(ctx, *apiServer, *authMode)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	results := rpb.NewResultsClient(conn)

	config := injection.ParseAndGetRESTConfigOrDie()

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create kubernetes client: %v", err)
	}

	sa := strings.Split(*authServiceAccount, ":")

	cfg := &reconciler.Config{
		Auth: reconciler.Auth{
			AuthMode:                *authMode,
			ServiceAccount:          sa[1],
			ServiceAccountNamespace: sa[0],
		},
		KubeClient:                   clientset,
		DisableAnnotationUpdate:      *crdUpdate,
		CompletedResourceGracePeriod: *completedRunGracePeriod,
	}

	sharedmain.MainWithConfig(ctx, "watcher", config,
		func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			return pipelinerun.NewControllerWithConfig(ctx, results, cfg)
		}, func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			return taskrun.NewControllerWithConfig(ctx, results, cfg)
		},
	)
}

func connectToAPIServer(ctx context.Context, apiServer string, authMode string) (*grpc.ClientConn, error) {
	// Load TLS certs
	creds, err := credentials.NewClientTLSFromFile(*tlsCert, "")
	if err != nil {
		log.Fatal("error loading TLS certificate: ", err)
	}

	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}
	// Add in additional credentials to requests if desired.
	switch authMode {
	case "google":
		opts = append(opts,
			grpc.WithAuthority(apiServer),
			grpc.WithTransportCredentials(creds),
			grpc.WithDefaultCallOptions(grpc.PerRPCCredentials(cred.Google())),
		)
	case "token":
		if t := *authToken; t != "" {
			opts = append(opts,
				grpc.WithDefaultCallOptions(grpc.PerRPCCredentials(oauth.TokenSource{
					TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t}),
				})),
				grpc.WithTransportCredentials(creds),
			)
		} else {
			return nil, errors.New("token parameter must be provide when auth mode is set to 'token'")
		}
	case "service-account":
		if sa := strings.Split(*authServiceAccount, ":"); len(sa) == 2 {
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			return nil, errors.New("invalid service account name passed in parameter. Format - Namespace:ServiceAccountName")
		}
	case "insecure":
		opts = append(opts, grpc.WithInsecure())
	}

	log.Printf("dialing %s...\n", apiServer)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, apiServer, opts...)
}

func loadCerts() (*x509.CertPool, error) {
	// Setup TLS certs to the server.
	f, err := os.Open("/etc/tls/tls.crt")
	if err != nil {
		log.Println("no local cluster cert found, defaulting to system pool...")
		return x509.SystemCertPool()
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read TLS cert file: %v", err)
	}

	certs, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("error loading cert pool: %v", err)
	}
	if ok := certs.AppendCertsFromPEM(b); !ok {
		return nil, fmt.Errorf("unable to add cert to pool")
	}
	return certs, nil
}
