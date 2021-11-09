package xchannel

import (
	"testing"
	"time"
)

func TestBoundless(t *testing.T) {
	for name, tc := range map[string]struct {
		bufferSize     int
		prematureClose bool
	}{
		"0":                 {},
		"10":                {bufferSize: 10},
		"prematureClose":    {prematureClose: true},
		"prematureClose,10": {bufferSize: 10, prematureClose: true},
	} {
		t.Run(name, func(t *testing.T) {
			b := NewBoundless(tc.bufferSize)

			for i := 0; i < 200; i++ {
				timeout := time.NewTimer(time.Second)
				select {
				case b.In() <- i:
				case <-timeout.C:
					t.Fatal("timed out")
				}
			}
			for i := 0; i < 100; i++ {
				timeout := time.NewTimer(time.Second)
				select {
				case elem, ok := <-b.Out():
					if !ok {
						t.Fatal("closed")
					}
					if i != elem.(int) {
						t.Errorf("expected %v, got %v.", i, elem)
					}
				case <-timeout.C:
					t.Fatal("timed out")
				}
			}
			if tc.prematureClose {
				b.Close()
			}
			i := 100
			for {
				if !tc.prematureClose && i == 200 {
					b.Close()
				}
				timeout := time.NewTimer(time.Second)
				select {
				case elem, ok := <-b.Out():
					if i == 200 {
						if ok {
							t.Fatal("not closed")
						}
						return
					}
					if !ok {
						t.Fatal("closed")
					}
					if i != elem.(int) {
						t.Errorf("expected %v, got %v.", i, elem)
					}
					i++
				case <-timeout.C:
					t.Fatal("timed out")
				}
			}
		})
	}
}
