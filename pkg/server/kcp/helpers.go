package kcp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/kcp-dev/logicalcluster"
	"google.golang.org/grpc/metadata"
)

// NamespaceLocator stores a logical cluster and namespace and is used
// as the source for the mapped namespace name in a physical cluster.
type NamespaceLocator struct {
	LogicalCluster logicalcluster.Name `json:"logical-cluster"`
	Namespace      string              `json:"namespace"`
}

// PhysicalClusterNamespaceName encodes the NamespaceLocator to a new
// namespace name for use on a physical cluster. The encoding is repeatable.
func PhysicalClusterNamespaceName(c string, n string) (string, error) {
	if c == "" {
		return n, nil
	}
	nl := NamespaceLocator{
		LogicalCluster: logicalcluster.New(c),
		Namespace:      n,
	}
	b, err := json.Marshal(nl)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum224(b)
	return fmt.Sprintf("kcp%x", hash), nil
}

func GetLogicalClusterName(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	v := md.Get("cluster")
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
