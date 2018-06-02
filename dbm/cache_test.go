package dbm

//
//import (
//	"testing"
//	"ParsinServer/algorithms/parameters"
//	"ParsinServer/glb"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestUserCache(t *testing.T) {
//	SetUserCache("zack", []string{"bob", "bill", "jane"})
//	users, _ := GetUserCache("zack")
//	assert.Equal(t, users, []string{"bob", "bill", "jane"})
//}
//
//func TestResetCache(t *testing.T) {
//	SetUserCache("zack", []string{"bob", "bill", "jane"})
//	ResetCache("userCache")
//	_, ok := GetUserCache("zack")
//	assert.Equal(t, ok, false)
//}
//
//func BenchmarkSetCache(b *testing.B) {
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		SetUserCache("zack", []string{"bob", "bill", "jane"})
//	}
//}
//
//func BenchmarkResetCache(b *testing.B) {
//	SetUserCache("zack", []string{"bob", "bill", "jane"})
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		ResetCache("userCache)")
//	}
//}
//
//// BenchmarkCache needs to have precomputed parameters for testdb (run Optimize after loading testdb.sh)
//func BenchmarkGetPSCache(b *testing.B) {
//	var err error
//	//db, err := boltOpen(path.Join("data", "testdb.db"), 0600, nil)
//	//if err != nil {
//	//	glb.Error.Println(err)
//	//}
//	//err = db.View(func(tx *bolt.Tx) error {
//	//	// Assume bucket exists and has keys
//	//	b := tx.Bucket([]byte("resources"))
//	//	if b == nil {
//	//		return fmt.Errorf("Resources dont exist")
//	//	}
//	//	v := b.Get([]byte("fullParameters"))
//	//	ps = parameters.LoadParameters(v)
//	//	return nil
//	//})
//	//if err != nil {
//	//	Error.Println(err)
//	//}
//	//db.Close()
//
//	var ps parameters.FullParameters
//	err = GetCompressedResourceInBucket("fullParameters",&ps,"resources","testdb")
//
//	if err != nil {
//		glb.Error.Println(err)
//	}
//
//
//
//	SetPsCache("testdb", ps)
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		GetPsCache("testdb")
//	}
//
//}
//
//// BenchmarkCache needs to have precomputed parameters for testdb (run Optimize after loading testdb.sh)
//func BenchmarkSetPSCache(b *testing.B) {
//	var err error
//	//var ps parameters.FullParameters
//	//db, err := boltOpen(path.Join("data", "testdb.db"), 0600, nil)
//	//if err != nil {
//	//	Error.Println(err)
//	//}
//	//err = db.View(func(tx *bolt.Tx) error {
//	//	// Assume bucket exists and has keys
//	//	b := tx.Bucket([]byte("resources"))
//	//	if b == nil {
//	//		return fmt.Errorf("Resources dont exist")
//	//	}
//	//	v := b.Get([]byte("fullParameters"))
//	//	ps = parameters.LoadParameters(v)
//	//	return nil
//	//})
//	//if err != nil {
//	//	Error.Println(err)
//	//}
//	//db.Close()
//	var ps parameters.FullParameters
//	err = GetCompressedResourceInBucket("fullParameters",&ps,"resources","testdb")
//
//	if err != nil {
//		glb.Error.Println(err)
//	}
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		SetPsCache("testdb", ps)
//	}
//
//}
