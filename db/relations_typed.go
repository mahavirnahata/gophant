package db

// EagerLoadTyped loads related rows and attaches them to each parent as []R.
// It is the typed equivalent of EagerLoadHasMany.
//
//	type Post struct {
//	    ID     int    `json:"id"`
//	    UserID int    `json:"user_id"`
//	    Title  string `json:"title"`
//	}
//
//	users, _ := Users.Get()
//	err := db.EagerLoad[User, Post](users, "posts", "user_id", conn, "posts", "id")
func EagerLoad[Parent any, Child any](
	parents []Parent,
	key string,
	foreignKey string,
	conn *DB,
	childTable string,
	localKey string,
	setter func(*Parent, []Child),
) error {
	if len(parents) == 0 {
		return nil
	}

	// collect local IDs
	ids := make([]any, 0, len(parents))
	parentMaps := make([]map[string]any, len(parents))
	for i, p := range parents {
		m := StructToMap(p)
		parentMaps[i] = m
		if v, ok := m[localKey]; ok {
			ids = append(ids, v)
		}
	}

	// fetch children grouped by foreign key
	rows, err := conn.Table(childTable).WhereIn(foreignKey, ids).Get()
	if err != nil {
		return err
	}

	// convert children to typed
	children, err := mapsToTyped[Child](rows)
	if err != nil {
		return err
	}

	// group by foreign key value
	childMaps := make([]map[string]any, len(children))
	for i, c := range children {
		childMaps[i] = StructToMap(c)
	}

	grouped := make(map[string][]Child, len(ids))
	for i, cm := range childMaps {
		fk := normalizeKey(cm[foreignKey])
		grouped[fk] = append(grouped[fk], children[i])
	}

	// attach to parents
	for i := range parents {
		lk := normalizeKey(parentMaps[i][localKey])
		setter(&parents[i], grouped[lk])
	}
	return nil
}

// BelongsToMany loads a many-to-many relationship through a pivot table.
// It returns a map from local ID to slice of related rows.
//
//	// users → roles via user_roles(user_id, role_id)
//	roleMap, err := db.BelongsToMany(conn, "roles", "user_roles", "user_id", "role_id", "id", userIDs)
func BelongsToMany(
	conn *DB,
	relatedTable string,
	pivotTable string,
	foreignKey string,
	relatedKey string,
	relatedPK string,
	localIDs []any,
) (map[any][]map[string]any, error) {
	if len(localIDs) == 0 {
		return map[any][]map[string]any{}, nil
	}

	// SELECT pivot.foreign_key, related.* FROM related
	// JOIN pivot ON pivot.related_key = related.related_pk
	// WHERE pivot.foreign_key IN (...)
	q := conn.Table(relatedTable).
		Select(pivotTable+"."+foreignKey, relatedTable+".*").
		WhereIn(pivotTable+"."+foreignKey, localIDs)
	q = q.Join(pivotTable, relatedTable+"."+relatedPK, "=", pivotTable+"."+relatedKey)

	rows, err := q.Get()
	if err != nil {
		return nil, err
	}

	out := map[any][]map[string]any{}
	for _, row := range rows {
		fk := row[foreignKey]
		out[fk] = append(out[fk], row)
	}
	return out, nil
}

// HasManyThrough loads a has-many-through relationship.
// Example: Country → has many User → has many Post (country has many posts through users).
//
//	// posts through users: country_id on users, user_id on posts
//	postMap, err := db.HasManyThrough(conn,
//	    "posts",  "users",
//	    "user_id", "country_id",
//	    "id",     countryIDs,
//	)
func HasManyThrough(
	conn *DB,
	finalTable string,
	throughTable string,
	finalForeignKey string,
	throughForeignKey string,
	throughPK string,
	localIDs []any,
) (map[any][]map[string]any, error) {
	if len(localIDs) == 0 {
		return map[any][]map[string]any{}, nil
	}

	// SELECT through.through_fk, final.* FROM final
	// JOIN through ON through.pk = final.final_fk
	// WHERE through.through_fk IN (...)
	q := conn.Table(finalTable).
		Select(throughTable+"."+throughForeignKey, finalTable+".*").
		WhereIn(throughTable+"."+throughForeignKey, localIDs)
	q = q.Join(throughTable, throughTable+"."+throughPK, "=", finalTable+"."+finalForeignKey)

	rows, err := q.Get()
	if err != nil {
		return nil, err
	}

	out := map[any][]map[string]any{}
	for _, row := range rows {
		fk := row[throughForeignKey]
		out[fk] = append(out[fk], row)
	}
	return out, nil
}

// MorphMany loads a polymorphic has-many relationship.
// morphType and morphID are the type/id columns on the related table.
// ownerType is the string that identifies the parent model (e.g. "App\\User").
//
//	comments, err := db.MorphMany(conn, "comments", "commentable_type", "commentable_id", "Post", postIDs)
func MorphMany(
	conn *DB,
	relatedTable string,
	morphTypeCol string,
	morphIDCol string,
	ownerType string,
	ownerIDs []any,
) (map[any][]map[string]any, error) {
	if len(ownerIDs) == 0 {
		return map[any][]map[string]any{}, nil
	}
	rows, err := conn.Table(relatedTable).
		Where(morphTypeCol, "=", ownerType).
		WhereIn(morphIDCol, ownerIDs).
		Get()
	if err != nil {
		return nil, err
	}
	out := map[any][]map[string]any{}
	for _, row := range rows {
		id := row[morphIDCol]
		out[id] = append(out[id], row)
	}
	return out, nil
}
