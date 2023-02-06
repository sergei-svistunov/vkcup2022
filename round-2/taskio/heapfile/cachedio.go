package heapfile

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sort"
)

const pageSize = 4096

type ioCache struct {
	f     *os.File
	size  int
	list  *list
	stat  map[string]int
	cache map[int]*ioCachePage
}

type ioCachePage struct {
	id      int
	data    [pageSize]byte
	dataLen int
	dirty   bool
	next    *ioCachePage
	prev    *ioCachePage
}

func (c *ioCache) read(pos int) (int64, error) {
	item, err := c.getPage(pos * 8 / pageSize)
	if err != nil {
		return 0, err
	}

	n := int64(binary.LittleEndian.Uint64(item.data[pos*8%pageSize:]))

	return n, nil
}

func (c *ioCache) write(pos int, n int64) error {
	item, err := c.getPage(pos * 8 / pageSize)
	if err != nil {
		return err
	}

	item.dirty = true
	binary.LittleEndian.PutUint64(item.data[pos*8%pageSize:], uint64(n))

	return nil
}

func (c *ioCache) flush() error {
	ids := make([]int, 0, len(c.cache))

	for id := range c.cache {
		ids = append(ids, id)
	}

	sort.Ints(ids)

	for _, pos := range ids {
		if err := c.flushPage(c.cache[pos]); err != nil {
			return err
		}
	}

	return nil
}

func (c *ioCache) getPage(pageId int) (*ioCachePage, error) {
	page := c.cache[pageId]
	if page != nil {
		c.list.moveLeft(page)
		return page, nil
	}

	if len(c.cache) >= c.size {
		page = c.list.tail

		c.list.remove(page)
		delete(c.cache, page.id)

		if err := c.flushPage(page); err != nil {
			return nil, err
		}
	} else {
		c.stat["allocate"]++
		page = &ioCachePage{}
	}

	c.stat["read_page"]++
	page.id = pageId
	page.dirty = false

	if n, err := c.f.ReadAt(page.data[:], int64(pageId*pageSize)); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		} else {
			page.dataLen = n
		}
	} else {
		page.dataLen = pageSize
	}

	c.list.pushBack(page)
	c.cache[pageId] = page

	return page, nil
}

func (c *ioCache) flushPage(page *ioCachePage) error {
	if page.dirty {
		c.stat["write_page"]++
		//c.stat[fmt.Sprintf("page_%d", page.id)]++

		if _, err := c.f.WriteAt(page.data[:page.dataLen], int64(page.id*pageSize)); err != nil {
			return err
		}
	}

	return nil
}

type list struct {
	head *ioCachePage
	tail *ioCachePage
}

func (l *list) pushBack(item *ioCachePage) {
	if l.tail == nil {
		l.head = item
		l.tail = item
		return
	}

	item.prev = l.tail
	item.next = nil
	l.tail.next = item
	l.tail = item
}

func (l *list) remove(item *ioCachePage) {
	if item == l.head {
		l.head = item.next
	}

	if item == l.tail {
		l.tail = item.prev
	}

	if item.prev != nil {
		item.prev.next = item.next
	}

	if item.next != nil {
		item.next.prev = item.prev
	}

}

func (l *list) moveLeft(item *ioCachePage) {
	if item.prev == nil { // already in head
		return
	}

	after := item.prev
	before := item.prev.prev
	l.remove(item)

	if before != nil {
		before.next = item
	} else {
		l.head = item
	}
	after.prev = item
	item.next = after
	item.prev = before
}
