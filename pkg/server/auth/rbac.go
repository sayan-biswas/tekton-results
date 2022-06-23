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
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"strings"
)

// RBAC is a Kubernetes RBAC based auth checker. This uses the Kubernetes
// TokenReview and SubjectAccessReview APIs to defer auth decisions to the
// cluster.
// Users should pass in `token` metadata through the gRPC context.
// This checks RBAC permissions in the `results.tekton.dev` group, and assumes
// checks are done at the namespace

type RBAC struct {
	*rest.Config
}

func NewRBAC(config *rest.Config) *RBAC {
	return &RBAC{
		config,
	}
}

func (c *RBAC) Check(ctx context.Context, cluster, namespace, resource, verb string) error {

	// Get token from authorization header
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "unable to get context metadata")
	}
	v := md.Get("authorization")
	if len(v) == 0 {
		return status.Error(codes.Unauthenticated, "Unable to find token")
	}
	s := strings.SplitN(v[0], " ", 2)
	if len(s) < 2 {
		return status.Error(codes.Unauthenticated, "Invalid bearer token")
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(&rest.Config{
		Host: fmt.Sprintf("%s/clusters/%s", c.Host, cluster),
		TLSClientConfig: rest.TLSClientConfig{
			CAFile: c.CAFile,
		},
		BearerToken: s[1],
	})
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	authz := clientset.AuthorizationV1()

	// Check RBAC rule
	ssar, err := authz.SelfSubjectAccessReviews().Create(ctx, &authzv1.SelfSubjectAccessReview{
		//ObjectMeta: metav1.ObjectMeta{
		//	Namespace: namespace,
		//},
		Spec: authzv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authzv1.ResourceAttributes{
				Group:     "results.tekton.dev",
				Resource:  resource,
				Verb:      verb,
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})

	if err != nil {
		log.Println("Error creating SelfSubjectAccessReview: ", err)
		return status.Error(codes.Unauthenticated, "Unauthorized")
	}
	if ssar.Status.Allowed {
		return nil
	}
	return status.Error(codes.Unauthenticated, "Unauthorized")
}

//type RBAC struct {
//	authn authnclient.AuthenticationV1Interface
//	authz authzclient.AuthorizationV1Interface
//}

//func NewRBAC(client kubernetes.Interface) *RBAC {
//	return &RBAC{
//		authn: client.AuthenticationV1(),
//		authz: client.AuthorizationV1(),
//	}
//}

//func (r *RBAC) Check(ctx context.Context, namespace, resource, verb string) error {
//	md, ok := metadata.FromIncomingContext(ctx)
//	if !ok {
//		return status.Error(codes.Unauthenticated, "unable to get context metadata")
//	}
//	v := md.Get("authorization")
//	if len(v) == 0 {
//		return status.Error(codes.Unauthenticated, "unable to find token")
//	}
//
//	for _, raw := range v {
//		// We expect tokens to be in the form "Bearer <token>". Parse the token out.
//		s := strings.SplitN(raw, " ", 2)
//		if len(s) < 2 {
//			log.Println("unknown auth token format")
//			continue
//		}
//		t := s[1]
//
//		// Authenticate the token by sending it to the API Server for review.
//		tr, err := r.authn.TokenReviews().Create(ctx, &authnv1.TokenReview{
//			Spec: authnv1.TokenReviewSpec{
//				Token: t,
//			},
//		}, metav1.CreateOptions{})
//		if err != nil {
//			log.Println(err)
//			continue
//		}
//		if !tr.Status.Authenticated {
//			continue
//		}
//
//		// Authorize the request by checking the RBAC permissions for the resource.
//		sar, err := r.authz.SubjectAccessReviews().Create(ctx, &authzv1.SubjectAccessReview{
//			Spec: authzv1.SubjectAccessReviewSpec{
//				User:   tr.Status.User.Username,
//				Groups: []string{"tekton.dev"},
//				ResourceAttributes: &authzv1.ResourceAttributes{
//					Namespace: namespace,
//					Group:     "results.tekton.dev",
//					Resource:  resource,
//					Verb:      verb,
//				},
//			},
//		}, metav1.CreateOptions{})
//		if err != nil {
//			log.Println(err)
//			continue
//		}
//		if sar.Status.Allowed {
//			return nil
//		}
//	}
//	// Return Unauthenticated - we don't know if we failed because of invalid
//	// token or unauthorized user, so this is safer to not leak any state.
//	return status.Error(codes.Unauthenticated, "permission denied")
//}
