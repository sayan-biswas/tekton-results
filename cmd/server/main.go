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
	"flag"
	"fmt"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tektoncd/results/pkg/server/api/v1alpha2"
	"github.com/tektoncd/results/pkg/server/auth"
	rpb "github.com/tektoncd/results/proto/results/v1alpha2"
	_ "go.uber.org/automaxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"net/http"
	"path"
)

var (
	dbHost      = flag.String("db-host", "", "Database host")
	dbPort      = flag.String("db-port", "5432", "Database port")
	dbName      = flag.String("db-name", "", "Database name")
	dbUser      = flag.String("db-user", "", "Database user")
	dbPassword  = flag.String("db-password", "", "Database password")
	dbSSL       = flag.String("db-ssl", "disable", "Enable/Disable database SSL mode")
	apiAuth     = flag.Bool("api-auth", false, "Enable/Disable API auth")
	grpcPort    = flag.String("grpc-port", "50051", "GRPC API Port")
	restPort    = flag.String("rest-port", "8080", "REST API Port")
	promPort    = flag.String("prometheus-port", "9090", "Prometheus Port")
	tlsPath     = flag.String("tls-path", "/etc/tls", "TLS cert and key path")
	tlsOverride = flag.String("tls-override", "", "TLS server name override")
)

func main() {
	flag.Parse()

	// Connect to the database.
	if *dbHost == "" || *dbName == "" || *dbUser == "" || *dbPassword == "" {
		log.Fatal("Must provide database host, name, user, password flag")
	}
	dbURI := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", *dbHost, *dbUser, *dbPassword, *dbName, *dbPort, *dbSSL)
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// Load server TLS cert
	creds, err := credentials.NewServerTLSFromFile(path.Join(*tlsPath, "tls.crt"), path.Join(*tlsPath, "tls.key"))
	if err != nil {
		log.Fatal("Error loading TLS key pair: ", err)
	}

	// Create auth client
	var authChecker auth.Checker
	authChecker = auth.AllowAll{}
	if *apiAuth {
		authChecker = auth.NewKCPAuth()
	}

	// Register API api(s)
	v1a2, err := v1alpha2.New(db, v1alpha2.WithAuth(authChecker))
	if err != nil {
		log.Fatal("Error creating GRPC server: ", err)
	}
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StreamInterceptor(prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(prometheus.UnaryServerInterceptor),
	)
	rpb.RegisterResultsServer(s, v1a2)

	// Allow service reflection - required for grpc_cli ls to work.
	reflection.Register(s)

	// Set up health checks.
	hs := health.NewServer()
	hs.SetServingStatus("tekton.result.v1alpha2.Results", hpb.HealthCheckResponse_SERVING)
	hpb.RegisterHealthServer(s, hs)

	// Prometheus metrics
	prometheus.Register(s)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("Prometheus server Listening on: %s", *promPort)
		if err := http.ListenAndServe(":"+*promPort, promhttp.Handler()); err != nil {
			log.Fatal("Error running Prometheus HTTP handler: ", err)
		}
	}()

	// Listen on port and serve.
	lis, err := net.Listen("tcp", ":"+*grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Serve gRPC server
	go func() {
		log.Printf("gRPC server Listening on: %s", *grpcPort)
		log.Fatal(s.Serve(lis))
	}()

	// Load client TLS cert
	creds, err = credentials.NewClientTLSFromFile(path.Join(*tlsPath, "tls.crt"), *tlsOverride)
	if err != nil {
		log.Fatal("Error loading TLS certificate: ", err)
	}

	// Register gRPC server endpoint for gRPC gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	err = rpb.RegisterResultsHandlerFromEndpoint(ctx, mux, ":"+*grpcPort, opts)
	if err != nil {
		log.Fatal("Error registering gRPC server endpoint: ", err)
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	log.Printf("REST server Listening on: %s", *restPort)
	log.Fatal(http.ListenAndServeTLS(":"+*restPort, path.Join(*tlsPath, "tls.crt"), path.Join(*tlsPath, "tls.key"), mux))
}
