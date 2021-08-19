package book

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/NateScarlet/qidian/pkg/client"
	"github.com/PuerkitoBio/goquery"
)

// Sort for search
type Sort string

// Sort for search
const (
	SortTotalRecommend Sort = "2"
	SortCharCount      Sort = "3"
	SortLastUpdated    Sort = "5"
	SortRecentFinished Sort = "6"
	SortWeekRecommend  Sort = "9"
	SortMonthRecommend Sort = "10"
	SortTotalBookmark  Sort = "11"
)

// State for book
type State string

// State for book
const (
	StateAll      State = ""
	StateOnGoing  State = "1"
	StateFinished State = "2"
)

// Sign for book
type Sign string

// Sign for book
const (
	// 全部作品
	SignAll Sign = ""
	// 签约作品
	SignSigned Sign = "1"
	// 精品小说
	SignChoicest Sign = "2"
)

// VIP state for book
type VIP string

// VIP state for book
const (
	VIPAll   VIP = ""
	VIPFalse VIP = "1"
	VIPTrue  VIP = "2"
)

// Update for book
type Update string

// Update for book
const (
	UpdateAll         Update = ""
	UpdateIn3Day      Update = "1"
	UpdateIn7Day      Update = "2"
	UpdateInHalfMonth Update = "3"
	UpdateInMonth     Update = "4"
)

// Size for book
type Size string

// Size for book
const (
	SizeAll          Size = ""
	SizeLt300k       Size = "1"
	SizeGt300kLt500k Size = "2"
	SizeGt500kLt1m   Size = "3"
	SizeGt1mLt2m     Size = "4"
	SizeGt2m         Size = "5"
)

// CategorySearch use https://www.qidian.com/all page
type CategorySearch struct {
	Site        string
	Sort        Sort
	Page        int
	Category    Category
	SubCategory SubCategory
	State       State
	Tag         string
	Sign        Sign
	Update      Update
	VIP         VIP
	Size        Size
}

// NewCategorySearch create a new search for function chaining.
func NewCategorySearch() *CategorySearch {
	return &CategorySearch{}
}

// SetPage then returns self.
func (s *CategorySearch) SetPage(v int) *CategorySearch {
	s.Page = v
	return s
}

// SetSort then returns self.
func (s *CategorySearch) SetSort(v Sort) *CategorySearch {
	s.Sort = v
	return s
}

// SetCategory then returns self.
func (s *CategorySearch) SetCategory(v Category) *CategorySearch {
	s.Category = v
	s.Site = v.Site()
	return s
}

// SetSubCategory and category then returns self.
func (s *CategorySearch) SetSubCategory(v SubCategory) *CategorySearch {
	s.SubCategory = v
	s.SetCategory(v.Parent())
	return s
}

// SetState then returns self.
func (s *CategorySearch) SetState(v State) *CategorySearch {
	s.State = v
	return s
}

// SetSign then returns self.
func (s *CategorySearch) SetSign(v Sign) *CategorySearch {
	s.Sign = v
	return s
}

// SetUpdate then returns self.
func (s *CategorySearch) SetUpdate(v Update) *CategorySearch {
	s.Update = v
	return s
}

// SetVIP then returns self.
func (s *CategorySearch) SetVIP(v VIP) *CategorySearch {
	s.VIP = v
	return s
}

// SetSize then returns self.
func (s *CategorySearch) SetSize(v Size) *CategorySearch {
	s.Size = v
	return s
}

// SetTag then returns self.
func (s *CategorySearch) SetTag(v string) *CategorySearch {
	s.Tag = v
	return s
}

// URL of search result page.
func (s CategorySearch) URL() string {
	u := url.URL{
		Scheme: "https",
		Host:   "www.qidian.com",
		Path:   "all",
	}
	if s.Site != "" {
		u.Path = s.Site + "/" + u.Path
	}
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	var filters = []string{}
	if s.Category != "" {
		filters = append(filters, fmt.Sprintf("chanId%s", string(s.Category)))
	}
	if s.SubCategory != "" {
		filters = append(filters, fmt.Sprintf("subCateId%s", string(s.SubCategory)))
	}
	if s.State != "" {
		filters = append(filters, fmt.Sprintf("action%s", string(s.State)))
	}
	if s.VIP != "" {
		filters = append(filters, fmt.Sprintf("vip%s", string(s.VIP)))
	}
	if s.Size != "" {
		filters = append(filters, fmt.Sprintf("size%s", string(s.Size)))
	}
	if s.Sign != "" {
		filters = append(filters, fmt.Sprintf("sign%s", string(s.Sign)))
	}
	if s.Update != "" {
		filters = append(filters, fmt.Sprintf("update%s", string(s.Update)))
	}
	if s.Sort != "" {
		filters = append(filters, fmt.Sprintf("orderId%s", string(s.Sort)))
	}
	if s.Tag != "" {
		filters = append(filters, fmt.Sprintf("tag%s", string(s.Tag)))
	}
	if s.Page > 1 {
		filters = append(filters, fmt.Sprintf("page%d", s.Page))
	}
	if len(filters) > 0 {
		u.Path += strings.Join(filters, "-") + "/"
	}
	return u.String()
}

// Execute search
func (s CategorySearch) Execute(ctx context.Context) (ret []Book, err error) {
	u := s.URL()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return
	}
	req.AddCookie(&http.Cookie{
		Name:  "listStyle",
		Value: "2",
	})
	res, err := client.For(ctx).Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	table := doc.
		Find("table.rank-table-list")
	if table.Length() == 0 {
		return nil, fmt.Errorf("qidian: can not found result table: %s", u)
	}
	return parseTable(table, nil, s.Site)
}
