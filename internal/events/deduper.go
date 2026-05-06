package events

const DefaultDeduperLimit = 1000

type Deduper struct {
	seen  map[string]struct{}
	order []string
	limit int
}

func NewDeduper(limit int) *Deduper {
	if limit <= 0 {
		limit = DefaultDeduperLimit
	}

	return &Deduper{
		seen:  make(map[string]struct{}, limit),
		order: make([]string, 0, limit),
		limit: limit,
	}
}

func (d *Deduper) IsNew(id string) bool {
	if id == "" {
		return false
	}
	if _, ok := d.seen[id]; ok {
		return false
	}

	d.seen[id] = struct{}{}
	d.order = append(d.order, id)
	d.evictOldest()
	return true
}

func (d *Deduper) evictOldest() {
	for len(d.order) > d.limit {
		oldest := d.order[0]
		d.order = d.order[1:]
		delete(d.seen, oldest)
	}
}
