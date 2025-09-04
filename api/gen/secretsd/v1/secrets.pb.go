// Code generated manually for bootstrap; replace with protoc output.

package v1

type PutRequest struct {
	Path  string
	Value string
}
type PutResponse struct {
	Version string
}

type GetRequest struct {
	Path    string
	Version string
}
type GetResponse struct {
	Value   string
	Version string
}

type SecretsClient interface {
	Put(*PutRequest) (*PutResponse, error)
	Get(*GetRequest) (*GetResponse, error)
}

type UnimplementedSecretsServer struct{}
