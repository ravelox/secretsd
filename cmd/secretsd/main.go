package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"secretsd/internal/crypto"
	"secretsd/internal/crypto/kek"
	"secretsd/internal/storage"
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
	http.HandleFunc("/v1/put", s.putHandler)
	http.HandleFunc("/v1/get", s.getHandler)

	addr := env("ADDR", ":8080")
	log.Printf("secretsd listening on %s  KEK=%s  PG=%s", addr, kekFile, dsn)
	log.Fatal(http.ListenAndServe(addr, nil))
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

func writeJSON(w http.ResponseWriter, v any) { w.Header().Set("Content-Type", "application/json"); _ = json.NewEncoder(w).Encode(v) }
func env(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }
