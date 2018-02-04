package parameters

import (
	"path"
	"testing"
	"github.com/boltdb/bolt"
	"ParsinServer/glb"
)

// It's using bolt to check LoadParameters
func BenchmarkLoadParameters(b *testing.B) {
	var ps FullParameters = *NewFullParameters()
	db, err := bolt.Open(path.Join("data", "testdb.db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte("resources"))
			if b == nil {
				glb.Error.Println("Resources dont exist")
				return ""
			}
			v := b.Get([]byte("fullParameters"))
			ps = LoadParameters(v)
			return nil
		})
		if err != nil {
			glb.Error.Println(err)
		}

	}
}

