package commands

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerBlockData(t *testing.T) {
	f, err := os.Open("testdata/server_blockdata.bin")
	assert.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()

	fmt.Print(f)

	payload := make([]byte, 64)
	count, err := f.Read(payload)
	assert.NoError(t, err)
	assert.True(t, count > 0)

	pkg := &ServerBlockData{}
	err = pkg.UnmarshalPacket(payload[0x37:])
	assert.NoError(t, err)

	assert.Equal(t, int16(32), pkg.Pos.PosX)
	assert.Equal(t, int16(-2), pkg.Pos.PosY)
	assert.Equal(t, int16(12), pkg.Pos.PosZ)
}
