package database

// Pagination limits.
const (
	DefaultLimit = 50
	MaxLimit     = 200
)

// Query describes a List call.
//
// Pagination is by cursor by default: pass the NextCursor of the previous page
// in After. Offset pagination is available for small administrative lists;
// After and Offset cannot be combined.
type Query struct {
	Filters []Filter
	// OrderBy is the column to sort by (default "id", which with UUIDv7 means
	// creation order). It must be a non-nullable column. Ties are broken by id.
	OrderBy string
	Desc    bool
	// Limit is the page size: DefaultLimit when 0, capped at MaxLimit.
	Limit  int
	After  string
	Offset int
}

// Page is one page of results.
type Page[T any] struct {
	Items []T
	// NextCursor fetches the next page (empty when there is none or with Offset).
	NextCursor string
	HasMore    bool
}

// Op is a filter operator.
type Op int

const (
	OpEq Op = iota
	OpIn
	OpIsNull
	OpNotNull
)

// Filter restricts a List or Count. Column must be one of the entity's columns.
type Filter struct {
	Column string
	Op     Op
	Value  any
}

// Eq matches rows where column equals value.
func Eq(column string, value any) Filter { return Filter{Column: column, Op: OpEq, Value: value} }

// In matches rows where column is one of values (a slice).
func In(column string, values any) Filter { return Filter{Column: column, Op: OpIn, Value: values} }

// IsNull matches rows where column is NULL.
func IsNull(column string) Filter { return Filter{Column: column, Op: OpIsNull} }

// NotNull matches rows where column is not NULL.
func NotNull(column string) Filter { return Filter{Column: column, Op: OpNotNull} }
