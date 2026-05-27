package pool

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenewPoolManagerWithConfig(t *testing.T) {
	cfg := &RenewPoolConfig{
		MinSize:       10,
		MaxSize:       50,
		ScaleUpRate:   0.8,
		ScaleDownRate: 0.3,
		CheckInterval: time.Minute,
		Expiry:        10 * time.Second,
		NonBlocking:   true,
	}
	mgr, err := NewRenewPoolManagerWithConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, mgr)

	_, capacity, _ := mgr.Stats()
	assert.True(t, capacity >= 10)
	mgr.Stop()
}

func TestNewRenewPoolManagerWithConfig_NilConfig(t *testing.T) {
	mgr, err := NewRenewPoolManagerWithConfig(nil)
	require.NoError(t, err)
	require.NotNil(t, mgr)
	mgr.Stop()
}

func TestRenewPoolManager_Submit(t *testing.T) {
	cfg := &RenewPoolConfig{
		MinSize:       20,
		MaxSize:       50,
		ScaleUpRate:   0.8,
		ScaleDownRate: 0.3,
		CheckInterval: time.Minute,
		Expiry:        10 * time.Second,
		NonBlocking:   false,
	}
	mgr, err := NewRenewPoolManagerWithConfig(cfg)
	require.NoError(t, err)
	defer mgr.Stop()

	var count int32
	for i := 0; i < 10; i++ {
		err := mgr.Submit(func() {
			atomic.AddInt32(&count, 1)
		})
		assert.NoError(t, err)
	}

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(10), atomic.LoadInt32(&count))
}

func TestRenewPoolManager_Stats(t *testing.T) {
	cfg := &RenewPoolConfig{
		MinSize:       5,
		MaxSize:       20,
		ScaleUpRate:   0.8,
		ScaleDownRate: 0.3,
		CheckInterval: time.Minute,
		Expiry:        10 * time.Second,
		NonBlocking:   true,
	}
	mgr, err := NewRenewPoolManagerWithConfig(cfg)
	require.NoError(t, err)
	defer mgr.Stop()

	running, capacity, usage := mgr.Stats()
	assert.True(t, capacity >= 5)
	assert.True(t, running >= 0)
	assert.True(t, usage >= 0)
}

func TestDefaultRenewPoolConfig(t *testing.T) {
	cfg := DefaultRenewPoolConfig()
	assert.Equal(t, DefaultMinSize, cfg.MinSize)
	assert.Equal(t, DefaultMaxSize, cfg.MaxSize)
	assert.Equal(t, DefaultScaleUpRate, cfg.ScaleUpRate)
	assert.Equal(t, DefaultScaleDownRate, cfg.ScaleDownRate)
	assert.Equal(t, DefaultCheckInterval, cfg.CheckInterval)
	assert.Equal(t, DefaultExpiry, cfg.Expiry)
}
