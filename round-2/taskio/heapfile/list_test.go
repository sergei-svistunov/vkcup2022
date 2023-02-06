package heapfile

import "testing"

func TestListPushBack(t *testing.T) {
	l := &list{}

	item1 := &ioCachePage{}

	t.Run("First push", func(t *testing.T) {
		l.pushBack(item1)
		if l.head != item1 || l.tail != item1 {
			t.Fatal("Invalid head or tail")
		}
	})

	item2 := &ioCachePage{}
	t.Run("Second push", func(t *testing.T) {
		l.pushBack(item2)
		if l.head != item1 || l.tail != item2 {
			t.Fatal("Invalid head or tail")
		}
		if item1.next != item2 || item2.prev != item1 {
			t.Fatal("Invalid next or prev")
		}
	})

	item3 := &ioCachePage{}
	t.Run("Third push", func(t *testing.T) {
		l.pushBack(item3)
		if l.head != item1 || l.tail != item3 {
			t.Fatal("Invalid head or tail")
		}
		if item1.next != item2 || item2.next != item3 {
			t.Fatal("Invalid next")
		}
		if item3.prev != item2 || item2.prev != item1 {
			t.Fatal("Invalid prev")
		}
	})
}

func TestListRemove(t *testing.T) {
	t.Run("1 item list", func(t *testing.T) {
		l := &list{}
		item := &ioCachePage{}
		l.pushBack(item)
		l.remove(item)
		if l.head != nil || l.tail != nil {
			t.Fatal("Invalid head or tail")
		}
	})

	t.Run("2 items list: head", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)
		l.remove(l.head)
		if l.head != item2 || l.tail != item2 {
			t.Fatal("Invalid head or tail")
		}
		if item2.next != nil || item2.prev != nil {
			t.Fatal("Invalid next or prev")
		}
	})

	t.Run("2 items list: tail", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)
		l.remove(l.tail)
		if l.head != item1 || l.tail != item1 {
			t.Fatal("Invalid head or tail")
		}
		if item1.next != nil || item1.prev != nil {
			t.Fatal("Invalid next or prev")
		}
	})

	t.Run("3 items list: head", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		item3 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)
		l.pushBack(item3)
		l.remove(l.head)
		if l.head != item2 || l.tail != item3 {
			t.Fatal("Invalid head or tail")
		}
		if item2.prev != nil || item2.next != item3 || item3.prev != item2 || item3.next != nil {
			t.Fatal("Invalid next or prev")
		}
	})

	t.Run("3 items list: middle", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		item3 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)
		l.pushBack(item3)
		l.remove(item2)
		if l.head != item1 || l.tail != item3 {
			t.Fatal("Invalid head or tail")
		}
		if item1.prev != nil || item1.next != item3 || item3.prev != item1 || item3.next != nil {
			t.Fatal("Invalid next or prev")
		}
	})

	t.Run("3 items list: tail", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		item3 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)
		l.pushBack(item3)
		l.remove(l.tail)
		if l.head != item1 || l.tail != item2 {
			t.Fatal("Invalid head or tail")
		}
		if item1.prev != nil || item1.next != item2 || item2.prev != item1 || item2.next != nil {
			t.Fatal("Invalid next or prev")
		}
	})
}

func TestListMoveLeft(t *testing.T) {
	t.Run("1 item list", func(t *testing.T) {
		l := &list{}
		item := &ioCachePage{}
		l.pushBack(item)

		l.moveLeft(item)
		if l.head != item || l.tail != item {
			t.Fatal("Invalid head or tail")
		}
		if item.prev != nil || item.next != nil {
			t.Fatal("Invalid next or prev")
		}
	})

	t.Run("2 items list", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)

		l.moveLeft(item2)
		if l.head != item2 || l.tail != item1 {
			t.Fatal("Invalid head or tail")
		}
		if item2.prev != nil || item2.next != item1 || item1.prev != item2 || item1.next != nil {
			t.Fatal("Invalid next or prev")
		}
	})

	t.Run("3 items list: mid", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		item3 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)
		l.pushBack(item3)

		l.moveLeft(item2)
		if l.head != item2 || l.tail != item3 {
			t.Fatal("Invalid head or tail")
		}
		if item2.prev != nil || item2.next != item1 || item1.prev != item2 || item1.next != item3 || item3.prev != item1 || item3.next != nil {
			t.Fatal("Invalid next or prev")
		}
	})

	t.Run("3 items list: tail", func(t *testing.T) {
		l := &list{}
		item1 := &ioCachePage{}
		item2 := &ioCachePage{}
		item3 := &ioCachePage{}
		l.pushBack(item1)
		l.pushBack(item2)
		l.pushBack(item3)

		l.moveLeft(l.tail)
		if l.head != item1 || l.tail != item2 {
			t.Fatal("Invalid head or tail")
		}
		if item1.prev != nil || item1.next != item3 || item3.prev != item1 || item3.next != item2 || item2.prev != item3 || item2.next != nil {
			t.Fatal("Invalid next or prev")
		}
	})
}
