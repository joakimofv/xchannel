package xchannel

import (
	"fmt"
	"testing"
	"time"
)

func TestBoundless(t *testing.T) {
	for _, bufferSize := range []int{10, 0} {
		t.Run(fmt.Sprintf("%v", bufferSize), func(t *testing.T) {
			b := NewBoundless(bufferSize)

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
			b.Close()
			i := 100
			for {
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
