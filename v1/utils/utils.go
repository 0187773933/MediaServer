package utils

import (
	"time"
	"io"
	"crypto/rand"
	ulid "github.com/oklog/ulid/v2"
)

var entropy io.Reader

func GenULID() ( result string ) {
	if entropy == nil {
		entropy = ulid.Monotonic( rand.Reader , 0 )
	}
	id , _ := ulid.New( ulid.Timestamp( time.Now() ) , entropy )
	result = id.String()
	return
}