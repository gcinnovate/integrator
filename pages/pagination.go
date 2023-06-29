package pages

// pagination.go

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"
	"math"
	"strconv"
)

// Paginator is the structure representing the paginator object
type Paginator struct {
	PageCount    int64
	PageSize     int64 // the limit
	Total        int64
	CurrentPage  int64
	NextPage     int64
	PreviousPage int64
	Offset       int64
	PageExists   bool
	Paging       bool
}

// HasNext returns true if there is a next page
func (p *Paginator) HasNext() bool {
	if p.HasPage(p.CurrentPage + 1) {
		return true
	}
	return false
}

// HasPrev returns true if there is a previous page
func (p *Paginator) HasPrev() bool {
	if p.HasPage(p.CurrentPage-1) && (p.CurrentPage-1) != 0 {
		return true
	}
	return false
}

// Pages returns the number of pages
func (p *Paginator) Pages() int64 {
	return int64(math.Ceil(
		float64(p.Total) / float64(p.PageSize)))
}

// HasPages returns true if there are pages
func (p *Paginator) HasPages() bool {
	if p.Total < 1 {
		return false
	}
	return true
}

// HasPage returns true if @param page is available
func (p *Paginator) HasPage(page int64) bool {
	if p.HasPages() && page <= p.Pages() {
		return true
	}
	return false
}

// FirstItem returns the number for first item in the page - used as OFFSET
func (p *Paginator) FirstItem() int64 {
	if p.PageCount < p.CurrentPage {
		return p.Total + 1
	}
	return int64(math.Min(float64((p.CurrentPage-1)*p.PageSize+1), float64(p.Total)))
}

// LastItem returns the number for the last item in the page
func (p *Paginator) LastItem() int64 {
	return int64(math.Min(float64(p.FirstItem()+p.PageSize-1), float64(p.Total)))
}

// GetPaginator returns a pointer to the Paginator structure
func GetPaginator(totalRecords, pageSize, page int64, paging bool) Paginator {
	p := Paginator{}
	p.Paging = true
	p.Total = totalRecords

	p.PageSize = pageSize

	p.CurrentPage = page
	p.PageCount = p.Pages()
	p.Offset = p.FirstItem() - 1
	if p.HasPrev() {
		p.PreviousPage = p.CurrentPage - 1
	}
	if p.HasNext() {
		p.NextPage = p.CurrentPage + 1
	}
	p.PageExists = p.HasPage(page)
	return p
}

func PaginationWidget(p *Paginator, currentPage *int, table *widget.Table) fyne.CanvasObject {
	showingLabel := widget.NewLabel("Show")
	itemsPerPageSelect := widget.NewSelectEntry([]string{"5", "10", "20", "30", "40", "50", "100"})
	itemsPerPageSelect.SetText("5")
	itemsPerPageSelect.OnChanged = func(s string) {
		fmt.Printf("Show %v per page\n", s)
	}
	perPageLabel := widget.NewLabel("per page")

	prevButton := widget.NewButton("Previous", func() {
		if *currentPage > 0 {
			*currentPage--
			table.Refresh()
		}
	})
	viewStr := fmt.Sprintf("Viewing %d-%d of %d", p.FirstItem(), p.LastItem(), p.Total)
	viewingData := binding.BindString(&viewStr)
	viewingLabel := widget.NewLabelWithData(viewingData)

	pageLabel := widget.NewLabel("Page")
	// pages := []string{"1", "2"}
	pages := lo.Times(int(p.PageCount), func(i int) string {
		return strconv.FormatInt(int64(i+1), 10)
	})
	pageSelect := widget.NewSelect(pages, func(page string) {
		fmt.Printf("Go to Page: %v\n", page)
	})
	if p.PageCount > 0 {
		pageSelect.Selected = pages[0]
	}
	totalPagesStr := fmt.Sprintf("of %d", p.PageCount)
	totalPagesBinding := binding.BindString(&totalPagesStr)
	totalPagesLabel := widget.NewLabelWithData(totalPagesBinding)

	nextButton := widget.NewButton("Next", func() {
		fmt.Println("Yes we pressed next")
		if (*currentPage < int(p.PageCount)-1) && p.HasPage(int64(*currentPage+1)) {
			*currentPage++
			fmt.Printf("Total Pages: %d, Current Page: %d, Next Page: %d, Has next Page: %v\n",
				p.PageCount, *currentPage, *currentPage+1, p.HasPage(int64(*currentPage+1)))
			table.Refresh()
		}
	})
	//if !p.HasPage(int64(*currentPage + 1)) {
	//	nextButton.Disabled()
	//} else {
	//	nextButton.Enable()
	//}

	container := container.NewHBox(
		showingLabel, itemsPerPageSelect, perPageLabel,
		prevButton,
		viewingLabel, pageLabel, pageSelect, totalPagesLabel,
		nextButton)
	return container
}
