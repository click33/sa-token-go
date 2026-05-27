package listener

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_RegisterAndTrigger(t *testing.T) {
	m := NewManager()
	called := false
	m.RegisterFunc(EventLogin, func(data *EventData) {
		called = true
		assert.Equal(t, EventLogin, data.Event)
		assert.Equal(t, "user1", data.LoginID)
	})
	m.TriggerSync(&EventData{Event: EventLogin, LoginID: "user1"})
	assert.True(t, called)
}

func TestManager_RegisterFunc(t *testing.T) {
	m := NewManager()
	var count int32
	m.RegisterFunc(EventLogin, func(data *EventData) {
		atomic.AddInt32(&count, 1)
	})
	m.TriggerSync(&EventData{Event: EventLogin})
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestManager_Priority(t *testing.T) {
	m := NewManager()
	var order []int
	var mu sync.Mutex

	m.RegisterFuncWithConfig(EventLogin, func(data *EventData) {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
	}, ListenerConfig{Priority: 0, Async: false})

	m.RegisterFuncWithConfig(EventLogin, func(data *EventData) {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
	}, ListenerConfig{Priority: 10, Async: false})

	m.RegisterFuncWithConfig(EventLogin, func(data *EventData) {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
	}, ListenerConfig{Priority: 5, Async: false})

	m.TriggerSync(&EventData{Event: EventLogin})
	assert.Equal(t, []int{3, 2, 1}, order)
}

func TestManager_AsyncTrigger(t *testing.T) {
	m := NewManager()
	var count int32
	m.RegisterFunc(EventLogin, func(data *EventData) {
		atomic.AddInt32(&count, 1)
	})
	m.Trigger(&EventData{Event: EventLogin})
	m.Wait()
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestManager_SyncTrigger(t *testing.T) {
	m := NewManager()
	var count int32
	m.RegisterFuncWithConfig(EventLogin, func(data *EventData) {
		atomic.AddInt32(&count, 1)
	}, ListenerConfig{Async: false})
	m.TriggerSync(&EventData{Event: EventLogin})
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestManager_Unregister(t *testing.T) {
	m := NewManager()
	called := false
	id := m.RegisterFunc(EventLogin, func(data *EventData) {
		called = true
	})
	assert.True(t, m.Unregister(id))
	m.TriggerSync(&EventData{Event: EventLogin})
	assert.False(t, called)
}

func TestManager_WildcardEvent(t *testing.T) {
	m := NewManager()
	var events []Event
	var mu sync.Mutex
	m.RegisterFunc(EventAll, func(data *EventData) {
		mu.Lock()
		events = append(events, data.Event)
		mu.Unlock()
	})
	m.TriggerSync(&EventData{Event: EventLogin})
	m.TriggerSync(&EventData{Event: EventLogout})
	assert.Contains(t, events, EventLogin)
	assert.Contains(t, events, EventLogout)
}

func TestManager_PanicRecovery(t *testing.T) {
	m := NewManager()
	var recovered any
	m.SetPanicHandler(func(event Event, data *EventData, r any) {
		recovered = r
	})
	m.RegisterFuncWithConfig(EventLogin, func(data *EventData) {
		panic("test panic")
	}, ListenerConfig{Async: false})
	m.TriggerSync(&EventData{Event: EventLogin})
	assert.Equal(t, "test panic", recovered)
}

func TestManager_Filter(t *testing.T) {
	m := NewManager()
	called := false
	m.AddFilter(func(data *EventData) bool {
		return data.LoginID != "blocked"
	})
	m.RegisterFunc(EventLogin, func(data *EventData) {
		called = true
	})
	m.TriggerSync(&EventData{Event: EventLogin, LoginID: "blocked"})
	assert.False(t, called)

	m.TriggerSync(&EventData{Event: EventLogin, LoginID: "allowed"})
	assert.True(t, called)
}

func TestManager_EnableDisableEvent(t *testing.T) {
	m := NewManager()
	called := false
	m.RegisterFunc(EventLogin, func(data *EventData) {
		called = true
	})

	m.DisableEvent(EventLogin)
	assert.False(t, m.IsEventEnabled(EventLogin))
	m.TriggerSync(&EventData{Event: EventLogin})
	assert.False(t, called)

	m.EnableEvent(EventLogin)
	assert.True(t, m.IsEventEnabled(EventLogin))
	m.TriggerSync(&EventData{Event: EventLogin})
	assert.True(t, called)
}

func TestManager_Stats(t *testing.T) {
	m := NewManager()
	m.EnableStats(true)
	m.RegisterFunc(EventLogin, func(data *EventData) {})
	m.RegisterFunc(EventLogout, func(data *EventData) {})

	m.TriggerSync(&EventData{Event: EventLogin})
	m.TriggerSync(&EventData{Event: EventLogin})
	m.TriggerSync(&EventData{Event: EventLogout})

	stats := m.GetStats()
	assert.Equal(t, int64(3), stats.TotalTriggered)
	assert.Equal(t, int64(2), stats.EventCounts[EventLogin])
	assert.Equal(t, int64(1), stats.EventCounts[EventLogout])
}

func TestManager_Clear(t *testing.T) {
	m := NewManager()
	m.RegisterFunc(EventLogin, func(data *EventData) {})
	m.RegisterFunc(EventLogout, func(data *EventData) {})
	assert.Equal(t, 2, m.Count())

	m.Clear()
	assert.Equal(t, 0, m.Count())
}

func TestManager_Count(t *testing.T) {
	m := NewManager()
	assert.Equal(t, 0, m.Count())

	m.RegisterFunc(EventLogin, func(data *EventData) {})
	assert.Equal(t, 1, m.Count())

	m.RegisterFunc(EventLogin, func(data *EventData) {})
	assert.Equal(t, 2, m.Count())

	m.RegisterFunc(EventLogout, func(data *EventData) {})
	assert.Equal(t, 3, m.Count())
	assert.Equal(t, 2, m.CountForEvent(EventLogin))
	assert.Equal(t, 1, m.CountForEvent(EventLogout))
}

func TestManager_GetListenerIDs(t *testing.T) {
	m := NewManager()
	id1 := m.RegisterFunc(EventLogin, func(data *EventData) {})
	id2 := m.RegisterFunc(EventLogin, func(data *EventData) {})

	ids := m.GetListenerIDs(EventLogin)
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
}

func TestManager_GetAllEvents(t *testing.T) {
	m := NewManager()
	m.RegisterFunc(EventLogin, func(data *EventData) {})
	m.RegisterFunc(EventLogout, func(data *EventData) {})

	events := m.GetAllEvents()
	assert.Len(t, events, 2)
	assert.Contains(t, events, EventLogin)
	assert.Contains(t, events, EventLogout)
}

func TestManager_HasListeners(t *testing.T) {
	m := NewManager()
	assert.False(t, m.HasListeners(EventLogin))

	m.RegisterFunc(EventLogin, func(data *EventData) {})
	assert.True(t, m.HasListeners(EventLogin))
	assert.False(t, m.HasListeners(EventLogout))
}

func TestEventData_String(t *testing.T) {
	data := &EventData{
		Event:   EventLogin,
		LoginID: "user1",
		Device:  "web",
	}
	s := data.String()
	require.Contains(t, s, "login")
	require.Contains(t, s, "user1")
}
