package memcached

type Item struct {
	key     string
	flags   uint32
	exptime uint32
	data    []byte
}
