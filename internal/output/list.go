package output

// ListColumn pairs a TableColumn definition with a value extractor for type T.
// Define one per column; the extractor receives a single list item and returns
// the string to display in that cell.
type ListColumn[T any] struct {
	TableColumn
	Value func(T) string
}

// RenderList builds a table from items using the column definitions in cols,
// then calls Print. Items for which any filter function returns false are
// excluded from the rendered output.
func RenderList[T any](format string, obj any, cols []ListColumn[T], items []T, filters ...func(T) bool) {
	headers := make([]TableColumn, len(cols))
	for i, c := range cols {
		headers[i] = c.TableColumn
	}
	rows := make([][]string, 0, len(items))
outer:
	for _, item := range items {
		for _, f := range filters {
			if !f(item) {
				continue outer
			}
		}
		row := make([]string, len(cols))
		for i, c := range cols {
			row[i] = c.Value(item)
		}
		rows = append(rows, row)
	}
	Print(format, obj, headers, rows)
}
