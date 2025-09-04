package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func main() {
	key := flag.String("key", "", "remote secret key, e.g. db/password")
	out := flag.String("out", "", "output file path")
	base := flag.String("base", "http://127.0.0.1:8080", "secretsd base URL")
	flag.Parse()
	if *key == "" || *out == "" { fmt.Println("usage: csi-smoketest --key db/password --out ./out/DB_PASSWORD"); os.Exit(2) }

	u, _ := url.Parse(*base + "/v1/get")
	q := u.Query(); q.Set("key", *key); u.RawQuery = q.Encode()
	cl := &http.Client{Timeout: 5 * time.Second}
	resp, err := cl.Get(u.String())
	if err != nil { fmt.Println("GET:", err); os.Exit(1) }
	defer resp.Body.Close()
	if resp.StatusCode != 200 { b, _ := io.ReadAll(resp.Body); fmt.Println("status:", resp.Status, string(b)); os.Exit(1) }
	var r struct{ Value, Version string }
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil { fmt.Println("json:", err); os.Exit(1) }
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil { fmt.Println("mkdir:", err); os.Exit(1) }
	if err := os.WriteFile(*out, []byte(r.Value), fs.FileMode(0o600)); err != nil { fmt.Println("write:", err); os.Exit(1) }
	fmt.Printf("wrote %s (version %s)\n", *out, r.Version)
}
