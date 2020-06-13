package book

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Sort for search
type Sort string

const (
	// STotalRecommend search sort
	STotalRecommend Sort = "2"
	// SCharCount search sort
	SCharCount = "3"
	// SLastUpdated search sort
	SLastUpdated = "5"
	// SRecentFinished search sort
	SRecentFinished = "6"
	// SWeekRecommend search sort
	SWeekRecommend = "9"
	// SMonthRecommend search sort
	SMonthRecommend = "10"
	// STotalBookmark search sort
	STotalBookmark = "11"
)

// Search options
type Search struct {
	// Keyword     *string
	Sort        Sort
	Page        int
	Category    Category
	SubCategory SubCategory
}

// NewSearch create a new search for function chaining.
func NewSearch() *Search {
	return &Search{}
}

// SetPage then returns self.
func (s *Search) SetPage(v int) *Search {
	s.Page = v
	return s
}

// SetSort then returns self.
func (s *Search) SetSort(v Sort) *Search {
	s.Sort = v
	return s
}

// SetCategory then returns self.
func (s *Search) SetCategory(v Category) *Search {
	s.Category = v
	return s
}

// SetSubCategory and category then returns self.
func (s *Search) SetSubCategory(v SubCategory) *Search {
	s.SubCategory = v
	s.Category = v.Parent()
	return s
}

func (s Search) excuteByAllPage(ctx context.Context) (ret []Book, err error) {
	u, err := url.Parse("https://www.qidian.com/all")
	if err != nil {
		return
	}
	q := u.Query()
	q.Set("style", "2")
	if s.Page > 1 {
		q.Set("page", strconv.Itoa(s.Page))
	}
	if s.Sort != "" {
		q.Set("orderId", string(s.Sort))
	}
	if s.Category != "" {
		q.Set("chanId", string(s.Category))
	}
	if s.SubCategory != "" {
		q.Set("subCateId", string(s.SubCategory))
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
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
		h, _ := doc.Html()
		fmt.Println(h)
		return nil, errors.New("can not found result table")
	}
	var columns = make([]string, 0)
	table.Find("thead > tr > th").Each(func(i int, s *goquery.Selection) {
		columns = append(columns, s.Text())
	})
	ret = make([]Book, 0, 50)
	table.
		Find("tbody > tr").
		EachWithBreak(func(i int, s *goquery.Selection) bool {
			var book = Book{}
			s.
				ChildrenFiltered("td").
				EachWithBreak(func(i int, s *goquery.Selection) bool {
					if i >= len(columns) {
						return false
					}
					switch columns[i] {
					case "类别":
						parts := strings.SplitN(strings.Trim(s.Text(), "「」"), "·", 2)
						if len(parts) != 2 {
							err = errors.New("unexpected category format")
							return false
						}
						book.Category = CategoryByName(parts[0])
						book.SubCategory = SubCategoryByName(parts[1])
					case "小说书名":
						book.Title = s.Text()
						book.ID, _ = s.Find("a").Attr("data-bid")
					case "小说作者":
						book.Author = s.Text()
					case "字数":
						book.CharCount, err = parseCountSelection(s)
						if err != nil {
							return false
						}
					case "总收藏":
						book.BookmarkCount, err = parseCountSelection(s)
						if err != nil {
							return false
						}
					case "周推荐":
						book.WeekRecommendCount, err = parseCountSelection(s)
						if err != nil {
							return false
						}
					case "月推荐":
						book.MonthRecommendCount, err = parseCountSelection(s)
						if err != nil {
							return false
						}
					case "总推荐":
						book.TotalRecommendCount, err = parseCountSelection(s)
						if err != nil {
							return false
						}
					case "更新时间":
						book.LastUpdated, err = parseTime(s.Text())
						if err != nil {
							return false
						}
					case "完本时间":
						book.Finished, err = parseTime(s.Text())
						if err != nil {
							return false
						}
					}
					return true
				})
			if err != nil {
				return false
			}
			ret = append(ret, book)
			return true
		})
	return

}

// Execute search
func (s Search) Execute(ctx context.Context) ([]Book, error) {
	return s.excuteByAllPage(ctx)
}
