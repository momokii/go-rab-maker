package utils

import (
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/momokii/go-rab-maker/backend/models"
)

func SetRefreshTableTriggerHeader(c *fiber.Ctx) {
	c.Set("HX-Trigger", "refreshTable")
}

func GetPaginationData(c *fiber.Ctx) (models.TablePaginationDataInput, error) {
	var input models.TablePaginationDataInput

	// get param data
	search := c.Query("search", "")
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		page = 1
	}
	perPage, err := strconv.Atoi(c.Query("per_page", "10"))
	if err != nil {
		perPage = 10
	}

	input.Search = search
	input.Page = page
	input.PerPage = perPage

	return input, nil
}

func GeneratePaginationLinks(currentPage, totalPages, padding int) []int {
	if totalPages <= 1 {
		return []int{}
	}

	// Gunakan map untuk menghindari duplikasi nomor halaman
	pageMap := make(map[int]bool)

	// 1. Selalu tambahkan halaman pertama dan terakhir
	pageMap[1] = true
	pageMap[totalPages] = true

	// 2. Tambahkan halaman di sekitar halaman saat ini (padding)
	for i := -padding; i <= padding; i++ {
		page := currentPage + i
		if page > 1 && page < totalPages {
			pageMap[page] = true
		}
	}

	// 3. Tambahkan halaman saat ini untuk memastikan
	if currentPage > 1 && currentPage < totalPages {
		pageMap[currentPage] = true
	}

	// Konversi map ke slice untuk diurutkan
	pages := make([]int, 0, len(pageMap))
	for page := range pageMap {
		pages = append(pages, page)
	}
	sort.Ints(pages)

	// 4. Masukkan elipsis (0) jika ada celah
	var result []int
	var lastPage int
	for _, page := range pages {
		if lastPage != 0 && page > lastPage+1 {
			result = append(result, 0) // 0 sebagai penanda elipsis
		}
		result = append(result, page)
		lastPage = page
	}

	return result
}
