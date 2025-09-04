
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	v1 "github.com/yourname/secretsd/api/gen/secretsd/v1"
	"github.com/yourname/secretsd/internal/crypto"
	"github.com/yourname/secretsd/internal/crypto/kek"
	"github.com/yourname/secretsd/internal/storage"
)

type Server struct {
	Store storage.Store
	Env   *crypto.Envelope
}

func main() {
	kekFile := env("KEK_FILE", "./.dev/kek.bin")
	fk, err := kek.NewFileKEK(kekFile)
	if err != nil { log.Fatalf("kek: %v", err) }
	envlp := crypto.NewEnvelope(fk)

	dsn := env("PG_DSN", "postgres://secretsd:secretsd@127.0.0.1:5432/secretsd?sslmode=disable")
	store, err := storage.NewPostgres(dsn)
	if err != nil { log.Fatalf("postgres: %v", err) }
	defer func(db *sql.DB){ _ = db.Close() }(store.UnderlyingDB())

	s := &Server{Store: store, Env: envlp}

	// HTTP mux
	http.HandleFunc("/v1/put", s.putHandler)
	http.HandleFunc("/v1/get", s.getHandler)

	// Start HTTP
	addr := env("ADDR", ":8080")
	go func() {
		log.Printf("HTTP listening on %s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	// Start gRPC
	grpcAddr := env("GRPC_ADDR", ":8081")
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil { log.Fatalf("grpc listen: %v", err) }
	gs := grpc.NewServer()
	v1.RegisterSecretsServer(gs, &grpcSvc{srv: s})
	log.Printf("gRPC listening on %s", grpcAddr)
	log.Fatal(gs.Serve(lis))
}

func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "use POST", http.StatusMethodNotAllowed); return }
	var req struct{ Path, Value string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad json", http.StatusBadRequest); return }
	ct, wdek, kid, err := s.Env.Encrypt([]byte(req.Value))
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	ver, err := s.Store.Put(req.Path, storage.SecretVersion{Ciphertext: ct, WrappedDEK: wdek, KEKID: kid})
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	writeJSON(w, map[string]any{"version": ver})
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("key")
	version := r.URL.Query().Get("version")
	v, err := s.Store.Get(path, version)
	if err != nil { http.Error(w, err.Error(), http.StatusNotFound); return }
	pt, err := s.Env.Decrypt(v.Ciphertext, v.WrappedDEK, v.KEKID)
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	writeJSON(w, map[string]any{"value": string(pt), "version": v.VersionID})
}

type grpcSvc struct{ srv *Server }

func (g *grpcSvc) Put(ctx context.Context, in *v1.PutRequest) (*v1.PutResponse, error) {
	ct, wdek, kid, err := g.srv.Env.Encrypt([]byte(in.Value))
	if err != nil { return nil, err }
	ver, err := g.srv.Store.Put(in.Path, storage.SecretVersion{Ciphertext: ct, WrappedDEK: wdek, KEKID: kid})
	if err != nil { return nil, err }
	return &v1.PutResponse{Version: ver}, nil
}

func (g *grpcSvc) Get(ctx context.Context, in *v1.GetRequest) (*v1.GetResponse, error) {
	v, err := g.srv.Store.Get(in.Path, in.Version)
	if err != nil { return nil, err }
	pt, err := g.srv.Env.Decrypt(v.Ciphertext, v.WrappedDEK, v.KEKID)
	if err != nil { return nil, err }
	return &v1.GetResponse{Value: string(pt), Version: v.VersionID}, nil
}

func writeJSON(w http.ResponseWriter, v any) { w.Header().Set("Content-Type", "application/json"); _ = json.NewEncoder(w).Encode(v) }
func env(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }
