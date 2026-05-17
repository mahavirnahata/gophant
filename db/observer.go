package db


// Observer receives lifecycle events from a Model. Any method can return an
// error to cancel the operation (Creating, Updating, Deleting only).
//
//	type AuditObserver struct{}
//	func (o *AuditObserver) Creating(data map[string]any) error { return nil }
//	func (o *AuditObserver) Created(data map[string]any)       {}
//	// ... etc.
//
//	UserModel.Observe(&AuditObserver{})
type Observer interface {
	Creating(data map[string]any) error
	Created(data map[string]any)
	Updating(data map[string]any) error
	Updated(data map[string]any)
	Deleting(id any) error
	Deleted(id any)
}

// NoopObserver embeds a do-nothing base so you only override what you need.
//
//	type MyObserver struct { db.NoopObserver }
//	func (o *MyObserver) Created(data map[string]any) { log.Println("created", data) }
type NoopObserver struct{}

func (NoopObserver) Creating(map[string]any) error { return nil }
func (NoopObserver) Created(map[string]any)        {}
func (NoopObserver) Updating(map[string]any) error { return nil }
func (NoopObserver) Updated(map[string]any)        {}
func (NoopObserver) Deleting(any) error            { return nil }
func (NoopObserver) Deleted(any)                   {}

// Observe registers an observer on the model. Multiple observers are supported;
// they run in registration order. Returning an error from Creating/Updating/Deleting
// aborts the operation.
func (m *Model) Observe(o Observer) {
	m.observers = append(m.observers, o)
}

// ── observed mutations ────────────────────────────────────────────────────────

// ObservedCreate runs Creating hooks, performs the insert, then runs Created hooks.
func (m *Model) ObservedCreate(data map[string]any) (int64, error) {
	for _, o := range m.observers {
		if err := o.Creating(data); err != nil {
			return 0, err
		}
	}
	id, err := m.Create(data)
	if err != nil {
		return 0, err
	}
	for _, o := range m.observers {
		o.Created(data)
	}
	return id, nil
}

// ObservedSave runs Updating hooks, performs the update, then runs Updated hooks.
func (m *Model) ObservedSave(id any, data map[string]any) error {
	for _, o := range m.observers {
		if err := o.Updating(data); err != nil {
			return err
		}
	}
	if err := m.Save(id, data); err != nil {
		return err
	}
	for _, o := range m.observers {
		o.Updated(data)
	}
	return nil
}

// ObservedDestroy runs Deleting hooks, performs the destroy, then runs Deleted hooks.
func (m *Model) ObservedDestroy(id any) error {
	for _, o := range m.observers {
		if err := o.Deleting(id); err != nil {
			return err
		}
	}
	if err := m.Destroy(id); err != nil {
		return err
	}
	for _, o := range m.observers {
		o.Deleted(id)
	}
	return nil
}

