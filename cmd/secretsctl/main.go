package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type client struct{ base string; hc *http.Client }

func newClient() *client { return &client{base: env("SECRETSD_URL", "http://127.0.0.1:8080"), hc: &http.Client{Timeout: 10 * time.Second}} }

func (c *client) put(path, val string) error {
	body := map[string]any{"path": path, "value": val}
	b, _ := json.Marshal(body)
	resp, err := c.hc.Post(c.base+"/v1/put", "application/json", strings.NewReader(string(b)))
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 { return fmt.Errorf("put %s: %s", path, resp.Status) }
	return nil
}

func migrateDotenv(c *client, file, prefix string) error {
	f, err := os.Open(file); if err != nil { return err }
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") { continue }
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 { continue }
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), "\"'")
		path := joinPath(prefix, key)
		if err := c.put(path, val); err != nil { return fmt.Errorf("%s: %w", path, err) }
		fmt.Printf("put %s\n", path)
	}
	return s.Err()
}

func migrateJSON(c *client, file, prefix string) error {
	b, err := os.ReadFile(file); if err != nil { return err }
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil { return err }
	for k, v := range m {
		path := joinPath(prefix, k)
		if err := c.put(path, v); err != nil { return fmt.Errorf("%s: %w", path, err) }
		fmt.Printf("put %s\n", path)
	}
	return nil
}

func joinPath(prefix, key string) string {
	p := strings.Trim(prefix, "/")
	k := strings.Trim(key, "/")
	if p == "" { return k }
	return p+"/"+k
}

func main() {
	if len(os.Args) < 2 { usage(); os.Exit(2) }
	sub := os.Args[1]
	switch sub {
	case "put":
		fs := flag.NewFlagSet("put", flag.ExitOnError)
		path := fs.String("path", "", "secret path (e.g. db/password)")
		val := fs.String("value", "", "secret value")
		_ = fs.Parse(os.Args[2:])
		if *path == "" || *val == "" { fs.Usage(); os.Exit(2) }
		c := newClient(); err := c.put(*path, *val); check(err)
		fmt.Println("ok")
	case "migrate":
		if len(os.Args) < 3 { migrateUsage(); os.Exit(2) }
		mode := os.Args[2]
		switch mode {
		case "dotenv":
			fs := flag.NewFlagSet("migrate dotenv", flag.ExitOnError)
			file := fs.String("file", ".env", ".env file path")
			prefix := fs.String("prefix", "", "path prefix, e.g. app/dev")
			_ = fs.Parse(os.Args[3:])
			c := newClient(); err := migrateDotenv(c, *file, *prefix); check(err)
		case "json":
			fs := flag.NewFlagSet("migrate json", flag.ExitOnError)
			file := fs.String("file", "secrets.json", "flat JSON file path")
			prefix := fs.String("prefix", "", "path prefix, e.g. app/dev")
			_ = fs.Parse(os.Args[3:])
			c := newClient(); err := migrateJSON(c, *file, *prefix); check(err)
		default:
			migrateUsage(); os.Exit(2)
		}
	default:
		usage(); os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `secretsctl <command> [flags]
Commands:
  put --path <p> --value <v>
  migrate dotenv --file .env --prefix app/dev
  migrate json   --file secrets.json --prefix app/dev

Env:
  SECRETSD_URL   base URL (default http://127.0.0.1:8080)
`)
}

func migrateUsage() { fmt.Fprintln(os.Stderr, "migrate {dotenv|json} ...") }

func env(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }

func check(err error) { if err != nil { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) } }
