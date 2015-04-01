package pagination

import (
	"math"
	"net/http"
	"reflect"
	"strconv"
)

// Paginator is a general purpose pagination type, it knows how to calculate
// offset and number of pages. It also contains some utility functions
// that helps common tasks. One special utility is the PagesStream method
// that returns a channel to range over for presenting a list of all pages
// without adding them all to a slice.
type Paginator struct {
	itemsPerPage  int
	numberOfItems int
	currentPage   int
}

// Pagination is a public version of the paginator. It does not have any logic
// attached and can be easily serialized to JSON.
type Pagination struct {
	ItemsPerPage  int           `json:"per_page"`
	NumberOfItems int           `json:"total_entries"`
	CurrentPage   int           `json:"page"`
	Offset        int           `json:"offset"`
	NextPage      int           `json:"next_page"`
	PreviousPage  int           `json:"previous_page"`
	TotalPages    int           `json:"total_pages"`
	Data          []interface{} `json:"data"`
}

// New returns a new Pagination with the provided values.
// The current page is normalized to be inside the bounds
// of the available pages. So if the current page supplied
// is less than 1 the current page is normalized as 1, and if
// it is larger than the number of pages needed its normalized
// as the last available page.
func New(numberOfItems, itemsPerPage, currentPage int) *Paginator {
	if currentPage == 0 {
		currentPage = 1
	}

	n := int(math.Ceil(float64(numberOfItems) / float64(itemsPerPage)))
	if currentPage > n {
		currentPage = n
	}

	return &Paginator{
		itemsPerPage:  itemsPerPage,
		numberOfItems: numberOfItems,
		currentPage:   currentPage,
	}
}

// NewFromRequest retusn a new Pagination with the provided values. However,
// unlike New, it uses an HTTP request to parse the page number to use.
func NewFromRequest(numberOfItems int, itemsPerPage int, req *http.Request) *Paginator {
	currentPageString := req.URL.Query().Get("page")
	currentPage, _ := strconv.Atoi(currentPageString)

	return New(numberOfItems, itemsPerPage, currentPage)
}

// ToPagination returns a Pagination instance which can be serialized and
// returned in API responses
func (p *Paginator) ToPagination() Pagination {
	return Pagination{
		ItemsPerPage:  p.ItemsPerPage(),
		NumberOfItems: p.NumberOfItems(),
		CurrentPage:   p.CurrentPage(),
		Offset:        p.Offset(),
		NextPage:      p.NextPage(),
		PreviousPage:  p.PreviousPage(),
		TotalPages:    p.NumberOfPages(),
		Data:          make([]interface{}, 0),
	}
}

// ToPaginationWithData is like ToPagination but it also includes some arbitrary
// data, which usually ends up being the databa being paginated.
func (p *Paginator) ToPaginationWithData(slice interface{}) Pagination {
	pagination := p.ToPagination()

	pagination.Data = interfaceSlice(slice)

	return pagination
}

// interfaceSlice converts a type slice into a more relaxed []interface{}
// which is used specifically for JSON encoding.
func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("interfaceSlice given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

// PagesStream returns a channel that will be incremented to
// the available number of pages. Useful to range over when
// building a list of pages.
func (p *Paginator) PagesStream() chan int {
	stream := make(chan int)
	go func() {
		for i := 1; i <= p.NumberOfPages(); i++ {
			stream <- i
		}
		close(stream)
	}()
	return stream
}

// Offset calculates the offset into the collection the current page represents.
func (p *Paginator) Offset() int {
	return (p.CurrentPage() - 1) * p.ItemsPerPage()
}

// NumberOfPages calculates the number of pages needed
// based on number of items and items per page.
func (p *Paginator) NumberOfPages() int {
	return int(math.Ceil(float64(p.NumberOfItems()) / float64(p.ItemsPerPage())))
}

// PreviousPage returns the page number of the page before current page.
// If current page is the first in the list of pages, 1 is returned.
func (p *Paginator) PreviousPage() int {
	if p.CurrentPage() <= 1 {
		return 1
	}

	return p.CurrentPage() - 1
}

// NextPage returns the page number of the page after current page.
// If current page is the last in the list of pages, the last page number is returned.
func (p *Paginator) NextPage() int {
	if p.CurrentPage() >= p.NumberOfPages() {
		return p.NumberOfPages()
	}

	return p.CurrentPage() + 1
}

// IsCurrentPage checks a number to see if it matches the current page.
func (p *Paginator) IsCurrentPage(page int) bool {
	return p.CurrentPage() == page
}

// Pages returns a list with all page numbers.
// Eg. [1 2 3 4 5]
func (p *Paginator) Pages() []int {
	s := make([]int, 0, p.NumberOfPages())

	for i := 1; i <= p.NumberOfPages(); i++ {
		s = append(s, i)
	}

	return s
}

// Show returns true if the pagination should be used.
// Ie. if there is more than one page.
func (p *Paginator) Show() bool {
	return p.NumberOfPages() > 1
}

// CurrentPage returns the current page.
func (p *Paginator) CurrentPage() int {
	return p.currentPage
}

// NumberOfItems returns the number of items.
func (p *Paginator) NumberOfItems() int {
	return p.numberOfItems
}

// ItemsPerPage returns the number of items to show per page.
func (p *Paginator) ItemsPerPage() int {
	return p.itemsPerPage
}
