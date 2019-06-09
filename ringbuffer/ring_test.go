package ringbuffer

import (
	"testing"
	"time"
)

func TestBuffer(t *testing.T) {
	buf := New(3)
	buf.Push([]byte("A."))
	buf.Push([]byte("B."))
	buf.Push([]byte("C."))
	buf.Push([]byte("D."))
	buf.Push([]byte("E."))
	buf.print()

	data := buf.Pop()
	if "E." != string(data) {
		t.Error("wrong data " + string(data) + " expect E")
	}
	buf.Pop()
	buf.Pop()
	buf.print()

	go func() {
		time.Sleep(1e9)
		buf.Push([]byte("F."))
	}()
	data = buf.Pop()
	if "F." != string(data) {
		t.Error("wrong data " + string(data) + " expect F")
	}

}
