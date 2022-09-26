// Copyright 2021 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	authnv1 "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"strings"
)

// KCP is a Kubernetes KCP based auth checker. This uses the Kubernetes
// TokenReview and SubjectAccessReview APIs to defer auth decisions to the
// cluster.
// Users should pass in `token` metadata through the gRPC context.
// This checks KCP permissions in the `results.tekton.dev` group, and assumes
// checks are done at the namespace

type claims struct {
	ClusterName string `json:"kubernetes.io/serviceaccount/clusterName,omitempty"`
}

type KCP struct {
	config *rest.Config
}

func NewKCPAuth() *KCP {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting kube config: %v", err)
	}
	cluster, err := getClusterName(config.BearerToken)
	if err != nil {
		log.Fatalf("Error getting cluster name: %v", err)
	}
	config.APIPath = fmt.Sprintf("clusters/%s", cluster)
	return &KCP{
		config,
	}
}

func (kcp *KCP) Check(ctx context.Context, cluster, namespace, resource, verb string) error {

	// Get token from authorization header
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "Metadata not found")
	}
	v := md.Get("authorization")
	if len(v) == 0 {
		return status.Error(codes.Unauthenticated, "Authorization header not found")
	}
	s := strings.SplitN(v[0], " ", 2)
	if len(s) < 2 {
		return status.Error(codes.Unauthenticated, "Invalid token")
	}
	token := s[1]

	// No check required for pod's SA account
	if token == kcp.config.BearerToken {
		return nil
	}

	// Check workspace access for the user
	client, err := kubernetes.NewForConfig(&rest.Config{
		Host:            kcp.config.Host + "/clusters/" + cluster,
		TLSClientConfig: kcp.config.TLSClientConfig,
		BearerToken:     token,
	})
	if err != nil {
		log.Printf("Error creating cluster clientset: %v", err)
		return status.Error(codes.Aborted, "Internal Server Error")
	}
	authn := client.AuthenticationV1()
	tr, err := authn.TokenReviews().Create(ctx, &authnv1.TokenReview{
		Spec: authnv1.TokenReviewSpec{
			Token: token,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error creating TokenReview: %v", err)
		return status.Error(codes.Aborted, "Error validating token")
	}
	if !tr.Status.Authenticated {
		return status.Error(codes.Unauthenticated,
			fmt.Sprintf("User %s doesn't have access to workspace %s",
				tr.Status.User.Username, cluster,
			),
		)
	}

	// Check resource access for user
	client, err = kubernetes.NewForConfig(kcp.config)
	if err != nil {
		log.Printf("Error creating cluster clientset: %v", err)
		return status.Error(codes.Aborted, "Internal Server Error")
	}
	authz := client.AuthorizationV1()
	sar, err := authz.SubjectAccessReviews().Create(ctx, &authzv1.SubjectAccessReview{
		Spec: authzv1.SubjectAccessReviewSpec{
			User:   tr.Status.User.Username,
			Groups: []string{"tekton.dev"},
			ResourceAttributes: &authzv1.ResourceAttributes{
				Group:     "results.tekton.dev",
				Resource:  resource,
				Verb:      verb,
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error creating SubjectAccessReview: ", err)
		return status.Error(codes.Aborted, "Internal Server Error")
	}
	if !sar.Status.Allowed {
		return status.Error(
			codes.PermissionDenied,
			fmt.Sprintf("User %s doesn't have access to %s %s in workspace %s",
				tr.Status.User.Username, verb, resource, cluster,
			),
		)
	}

	return nil
}

func (c claims) Valid() error {
	return nil
}

func getClusterName(saToken string) (string, error) {
	jwtParser := jwt.NewParser()
	token, _, err := jwtParser.ParseUnverified(saToken, &claims{})
	if err != nil {
		return "", err
	}
	if c, ok := token.Claims.(*claims); ok {
		return c.ClusterName, nil
	}
	return "", errors.New("clusterName claim not found in token")
}
