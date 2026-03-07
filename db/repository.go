package db

type Hooks struct {
	BeforeCreate func(map[string]any) error
	AfterCreate  func(map[string]any) error
	BeforeUpdate func(map[string]any) error
	AfterUpdate  func(map[string]any) error
	BeforeDelete func() error
	AfterDelete  func() error
}

type Repository struct {
	DB    *DB
	Table string
	Hooks Hooks
}

func (r *Repository) Insert(data map[string]any) error {
	if r.Hooks.BeforeCreate != nil {
		if err := r.Hooks.BeforeCreate(data); err != nil {
			return err
		}
	}
	_, err := r.DB.Table(r.Table).Insert(data)
	if err != nil {
		return err
	}
	if r.Hooks.AfterCreate != nil {
		return r.Hooks.AfterCreate(data)
	}
	return nil
}

func (r *Repository) Update(whereCol string, whereVal any, data map[string]any) error {
	if r.Hooks.BeforeUpdate != nil {
		if err := r.Hooks.BeforeUpdate(data); err != nil {
			return err
		}
	}
	_, err := r.DB.Table(r.Table).Where(whereCol, "=", whereVal).Update(data)
	if err != nil {
		return err
	}
	if r.Hooks.AfterUpdate != nil {
		return r.Hooks.AfterUpdate(data)
	}
	return nil
}

func (r *Repository) Delete(whereCol string, whereVal any) error {
	if r.Hooks.BeforeDelete != nil {
		if err := r.Hooks.BeforeDelete(); err != nil {
			return err
		}
	}
	_, err := r.DB.Table(r.Table).Where(whereCol, "=", whereVal).Delete()
	if err != nil {
		return err
	}
	if r.Hooks.AfterDelete != nil {
		return r.Hooks.AfterDelete()
	}
	return nil
}
