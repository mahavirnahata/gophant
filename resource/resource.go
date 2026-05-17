// Package resource provides generic helpers for transforming model data into
// JSON-serializable maps before sending API responses.
//
// Usage:
//
//	// Define a transformer function:
//	func UserResource(u map[string]any) map[string]any {
//	    return map[string]any{
//	        "id":    u["id"],
//	        "name":  u["name"],
//	        "email": u["email"],
//	        // deliberately omit "password_hash", "remember_token", etc.
//	    }
//	}
//
//	// In a controller:
//	row, _ := UserModel.Find(id)
//	c.JSON(200, resource.One(row, UserResource))
//
//	rows, _ := UserModel.Get()
//	c.JSON(200, resource.Collection(rows, UserResource, nil))
package resource

import "github.com/mahavirnahata/gophant/db"

// Transformer converts an item of type T into a JSON-safe map.
type Transformer[T any] func(item T) map[string]any

// One transforms a single item.
func One[T any](item T, fn Transformer[T]) map[string]any {
	return fn(item)
}

// Many transforms a slice of T into a slice of maps.
func Many[T any](items []T, fn Transformer[T]) []map[string]any {
	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

// Collection wraps transformed items in a standard {data, meta} envelope.
// Pass nil for meta to omit it.
//
//	c.JSON(200, resource.Collection(rows, UserResource, map[string]any{
//	    "total": 100, "page": 1, "per_page": 15,
//	}))
func Collection[T any](items []T, fn Transformer[T], meta map[string]any) map[string]any {
	resp := map[string]any{"data": Many(items, fn)}
	if meta != nil {
		resp["meta"] = meta
	}
	return resp
}

// FromPage builds a Collection response from a db.Page result.
//
//	page, _ := UserModel.Query().Paginate(pageNum, 15)
//	c.JSON(200, resource.FromPage(page, UserResource))
func FromPage(page db.Page, fn Transformer[map[string]any]) map[string]any {
	pages := 0
	if page.PerPage > 0 {
		pages = (page.Total + page.PerPage - 1) / page.PerPage
	}
	return Collection(page.Data, fn, map[string]any{
		"total":    page.Total,
		"page":     page.Page,
		"per_page": page.PerPage,
		"pages":    pages,
	})
}
