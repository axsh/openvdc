package backend

import "github.com/boltdb/bolt"

type Bolt struct {
	db       *bolt.DB
	basePath string
}

func NewBoltBackend() *Bolt {
	return &Bolt{
		basePath: "/openvdc",
	}
}

func (b *Bolt) Connect(dest []string) error {
	if b.db != nil {
		return ErrConnectionExists
	}
	db, err := bolt.Open(dest[0], 0644, nil)
	if err != nil {
		return err
	}
	b.db = db
	return nil
}

func (b *Bolt) Close() error {
	if b.db == nil {
		return ErrConnectionNotReady
	}
	defer func() {
		b.db = nil
	}()
	return b.db.Close()
}

func (b *Bolt) Create(key string, value []byte) error {
	return nil
}

func (b *Bolt) Update(key string, value []byte) error {
	return nil
}

func (b *Bolt) CreateWithID(key string, value []byte) (string, error) {
	return "", nil
}

func (b *Bolt) Find(key string) ([]byte, error) {
	return nil, nil
}

func (b *Bolt) Delete(key string) error {
	return nil
}

func (b *Bolt) Keys(parentKey string) (KeyIterator, error) {
	return nil, nil
}

func (b *Bolt) FindLastKey(prefixKey string) (string, error) {
	return "", nil
}
