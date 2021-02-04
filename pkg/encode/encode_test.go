package encode

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDurationText(t *testing.T) {
	var (
		input  = []byte("10s")
		output = time.Second * 10
		d      Duration
	)
	if err := d.UnmarshalText(input); err != nil {
		t.FailNow()
	}
	if int64(output) != int64(d) {
		t.FailNow()
	}

	enc, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, input, enc)
}

func TestConcurrentMapReadAndMapWrite(t *testing.T) {
	m := make(map[int]string)
	for i := 0; i < 3; i++ {
		m[i] = fmt.Sprintf("data%d", i)
	}

	//for i := 0;i<100;i++ {
	//	go t.Logf(m[2][:])
	//}
	for i := 0; i < 60; i++ {
		go func(idx int) {
			m[2] = fmt.Sprintf("data%d", idx)
		}(i)
	}
	time.Sleep(1 * time.Second)
}
