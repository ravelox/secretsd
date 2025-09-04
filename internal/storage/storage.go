package storage

type SecretVersion struct {
	VersionID  string
	Ciphertext []byte
	WrappedDEK []byte
	KEKID      string
	CreatedAt  int64
}

type Store interface {
	Put(path string, v SecretVersion) (string, error)
	Get(path, version string) (*SecretVersion, error)
}
