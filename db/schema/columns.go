package schema

import "fmt"

func (bp *Blueprint) BigInteger(name string) {
	bp.columns = append(bp.columns, Column{Name: name, Type: "BIGINT"})
}

func (bp *Blueprint) UnsignedBigInteger(name string) {
	bp.columns = append(bp.columns, Column{Name: name, Type: "BIGINT UNSIGNED"})
}

func (bp *Blueprint) Text(name string) {
	bp.columns = append(bp.columns, Column{Name: name, Type: "TEXT"})
}

func (bp *Blueprint) JSON(name string) {
	if bp.driver == "postgres" || bp.driver == "postgresql" {
		bp.columns = append(bp.columns, Column{Name: name, Type: "JSONB"})
		return
	}
	bp.columns = append(bp.columns, Column{Name: name, Type: "JSON"})
}

func (bp *Blueprint) Decimal(name string, precision, scale int) {
	if precision <= 0 {
		precision = 10
	}
	if scale < 0 {
		scale = 0
	}
	bp.columns = append(bp.columns, Column{Name: name, Type: fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)})
}

func (bp *Blueprint) Enum(name string, values ...string) {
	if len(values) == 0 {
		return
	}
	vals := "'" + values[0] + "'"
	for i := 1; i < len(values); i++ {
		vals += ",'" + values[i] + "'"
	}
	bp.columns = append(bp.columns, Column{Name: name, Type: "ENUM(" + vals + ")"})
}

func (bp *Blueprint) CompositeIndex(name string, columns ...string) {
	if len(columns) == 0 {
		return
	}
	bp.indexes = append(bp.indexes, fmt.Sprintf("CREATE INDEX %s ON %s (%s)", name, bp.table, join(columns)))
}

func (bp *Blueprint) CompositeUnique(name string, columns ...string) {
	if len(columns) == 0 {
		return
	}
	bp.uniques = append(bp.uniques, fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)", name, join(columns)))
}

func join(cols []string) string {
	out := cols[0]
	for i := 1; i < len(cols); i++ {
		out += "," + cols[i]
	}
	return out
}
