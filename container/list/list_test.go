package list

import (
    "testing"
)

func TestListBasic(t *testing.T) {
    l := New[int]()
    l.Append(1, 2, 3)
    if l.Len() != 3 {
        t.Errorf("Len = %d, want 3", l.Len())
    }
    if v := l.Get(0); v != 1 {
        t.Errorf("Get(0) = %d, want 1", v)
    }
    l.Set(1, 42)
    if v := l.Get(1); v != 42 {
        t.Errorf("Set/Get(1) = %d, want 42", v)
    }
}