package page

type PageID string

// PageChangeMsg is used to change the current page
type PageChangeMsg struct {
	ID   PageID
	Data interface{} // Optional data to pass to the new page
}
