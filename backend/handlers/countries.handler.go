package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

var (
	countries_list = []models.Country{
		{ID: 1, Name: "United States", Code: "US", Description: "United States of America", DialCode: "+1", FlagURL: "https://flagcdn.com/us.svg"},
		{ID: 2, Name: "Canada", Code: "CA", Description: "Canada", DialCode: "+1", FlagURL: "https://flagcdn.com/ca.svg"},
		{ID: 3, Name: "United Kingdom", Code: "GB", Description: "United Kingdom", DialCode: "+44", FlagURL: "https://flagcdn.com/gb.svg"},
		{ID: 4, Name: "Australia", Code: "AU", Description: "Australia", DialCode: "+61", FlagURL: "https://flagcdn.com/au.svg"},
		{ID: 5, Name: "Germany", Code: "DE", Description: "Germany", DialCode: "+49", FlagURL: "https://flagcdn.com/de.svg"},
		{ID: 6, Name: "France", Code: "FR", Description: "France", DialCode: "+33", FlagURL: "https://flagcdn.com/fr.svg"},
		{ID: 7, Name: "Japan", Code: "JP", Description: "Japan", DialCode: "+81", FlagURL: "https://flagcdn.com/jp.svg"},
		{ID: 8, Name: "India", Code: "IN", Description: "India", DialCode: "+91", FlagURL: "https://flagcdn.com/in.svg"},
		{ID: 9, Name: "Brazil", Code: "BR", Description: "Brazil", DialCode: "+55", FlagURL: "https://flagcdn.com/br.svg"},
		{ID: 10, Name: "South Africa", Code: "ZA", Description: "South Africa", DialCode: "+27", FlagURL: "https://flagcdn.com/za.svg"},
	}
)

// views
func CountriesView(c *fiber.Ctx) error {

	// userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	countries_component := components.Countries(countries_list)

	return adaptor.HTTPHandler(templ.Handler(countries_component))(c)
}

func CreateCountriesModal(c *fiber.Ctx) error {
	modal := components.CountryFormModal(
		"Add New Country",
		"/countries/new",
		"new-country-form",
		"Add Country",
		models.Country{}, // empty country for new form
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func EditCountriesModal(c *fiber.Ctx) error {
	countryId := c.Params("id")
	countryIdInt, _ := strconv.Atoi(countryId)
	var countryData models.Country
	for _, country := range countries_list {
		if country.ID == countryIdInt {
			countryData = country
			break
		}
	}
	modal := components.CountryFormModal(
		"Edit Country",
		"/countries/"+countryId+"/edit",
		"edit-country-form",
		"Update Country",
		countryData,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func DeleteCountriesModal(c *fiber.Ctx) error {
	countryId := c.Params("id")
	countryIdInt, _ := strconv.Atoi(countryId)
	var countryData models.Country
	for _, country := range countries_list {
		if country.ID == countryIdInt {
			countryData = country
			break
		}
	}

	modal := components.ConfirmationDeleteModal(
		"Delete Country",
		"Are you sure you want to delete the country "+countryData.Name+"?",
		"/countries/"+countryId+"/delete",
		"Delete Countries",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func AddNewCountries(c *fiber.Ctx) error {
	time.Sleep(3 * time.Second)

	errorTest := false
	if errorTest {
		return utils.ResponseErrorModal(c, "Error", "There was an error adding the new country.")
	}

	// dummy add new countries
	countries_list = append(countries_list, models.Country{
		ID:          len(countries_list) + 1,
		Name:        c.FormValue("name"),
		Code:        c.FormValue("code"),
		Description: c.FormValue("description"),
		DialCode:    c.FormValue("dial_code"),
		FlagURL:     "https://flagcdn.com/us.svg", // default flag
	})

	// trigger the refresh table
	utils.SetRefreshTableTriggerHeader(c)

	return utils.ResponseSuccessModal(c, "Success", "New country has been added successfully.", true)
}

func EditCountries(c *fiber.Ctx) error {
	time.Sleep(3 * time.Second)

	errorTest := false
	if errorTest {
		return utils.ResponseErrorModal(c, "Error", "There was an error adding the edit country.")
	}

	countryId := c.Params("id")
	countryIdInt, _ := strconv.Atoi(countryId)
	for idx, countryData := range countries_list {
		if countryData.ID == countryIdInt {

			country := &countries_list[idx] // get the pointer to the country to directly

			// edit the data
			country.Name = c.FormValue("name")
			country.Code = c.FormValue("code")
			country.Description = c.FormValue("description")
			country.DialCode = c.FormValue("dial_code")
			break
		}
	}

	// trigger the refresh table
	utils.SetRefreshTableTriggerHeader(c)

	return utils.ResponseSuccessModal(c, "Success", "Edit Country Success.", true)
}

func DeleteCountries(c *fiber.Ctx) error {
	time.Sleep(3 * time.Second)

	errorTest := false
	if errorTest {
		return utils.ResponseErrorModal(c, "Error", "There was an error deleting the country.")
	}

	countryId := c.Params("id")
	countryIdInt, _ := strconv.Atoi(countryId)
	for idx, countryData := range countries_list {
		if countryData.ID == countryIdInt {
			// delete the country from the list
			countries_list = append(countries_list[:idx], countries_list[idx+1:]...)
			break
		}
	}

	// trigger the refresh table
	utils.SetRefreshTableTriggerHeader(c)

	return utils.ResponseSuccessModal(c, "Success", "Delete country has successs.", true)
}

func CountriesTableView(c *fiber.Ctx) error {

	// get the pagination input data for table
	paginationData, err := utils.GetPaginationData(c)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to get pagination data")
	}

	// Get pagination and filtering parameters
	currentPage := paginationData.Page
	totalPages := (len(countries_list) + paginationData.PerPage - 1) / paginationData.PerPage
	if currentPage < 1 {
		currentPage = 1
	} else if currentPage > totalPages {
		currentPage = totalPages
	}
	totalItems := len(countries_list)
	itemsPerPage := paginationData.PerPage

	countries_list_filtered := countries_list

	// filter country list based on the search user data
	if paginationData.Search != "" {
		var filteredCountries []models.Country
		for _, country := range countries_list_filtered {
			if strings.Contains(strings.ToLower(country.Name), strings.ToLower(paginationData.Search)) ||
				strings.Contains(strings.ToLower(country.Code), strings.ToLower(paginationData.Search)) ||
				strings.Contains(strings.ToLower(country.Description), strings.ToLower(paginationData.Search)) ||
				strings.Contains(strings.ToLower(country.DialCode), strings.ToLower(paginationData.Search)) {
				filteredCountries = append(filteredCountries, country)
			}
		}
		countries_list_filtered = filteredCountries
		// recalculate total items and pages after filtering
		totalItems = len(countries_list_filtered)
		totalPages = (totalItems + itemsPerPage - 1) / itemsPerPage
		if currentPage > totalPages {
			currentPage = totalPages
		}
		if currentPage < 1 {
			currentPage = 1
		}
	}

	// setup country list based on the input per page and page data
	countries_list_filtered = countries_list[(currentPage-1)*itemsPerPage : currentPage*itemsPerPage]
	if currentPage*itemsPerPage > totalItems {
		countries_list_filtered = countries_list[(currentPage-1)*itemsPerPage : totalItems]
	}

	// base data for table
	tableConfig := models.TableConfig{
		BaseURL:           "/",
		Title:             "Countries",
		SearchEnabled:     true,
		PaginationEnabled: true,
		PerPageEnabled:    true,
	}

	paginationInfo := models.PaginationInfo{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: itemsPerPage,
	}

	if c.Get("HX-Request") == "true" {
		tableComponents := components.CountriesTablePage(countries_list_filtered, paginationInfo, tableConfig)
		return adaptor.HTTPHandler(templ.Handler(tableComponents))(c)
	}

	countries_component := components.CountriesPage(countries_list_filtered, paginationInfo, tableConfig)
	return adaptor.HTTPHandler(templ.Handler(countries_component))(c)
}

// handler

func GetCountryDetails(c *fiber.Ctx) error {

	countryId := c.Params("id")
	countryIdInt, _ := strconv.Atoi(countryId)

	countryData := models.Country{}
	for _, country := range countries_list {
		if country.ID == countryIdInt {
			countryData = country
			break
		}
	}

	country_details_component := components.CountryDetails(countryData)

	return adaptor.HTTPHandler(templ.Handler(country_details_component))(c)
}

func SearchCountries(c *fiber.Ctx) error {

	var result []models.Country
	query := c.Query("search")

	for _, country := range countries_list {
		if strings.Contains(strings.ToLower(country.Name), strings.ToLower(query)) {
			result = append(result, country)
		}
	}

	countries_component := components.CountryList(result)

	return adaptor.HTTPHandler(templ.Handler(countries_component))(c)
}

func DemoLoadingTest(c *fiber.Ctx) error {
	time.Sleep(5 * time.Second)

	errorTest := false
	if errorTest {
		return utils.ResponseErrorModal(c, "Loading Failed", "There was an error during the loading test.")
	}

	isAuthTest := false
	if isAuthTest {
		// return utils.ResponseErrorModalNotAuthDirect(c, "Unauthorized", "You are not authorized to perform this action.")
		return utils.ResponseRedirectHTMX(c, middlewares.LOGIN_PAGE_URL, fiber.StatusUnauthorized)
	}

	// componentsReturn := components.CountryFormModal("asd", "/api/countries", models.Country{})
	// return adaptor.HTTPHandler(templ.Handler(componentsReturn))(c)

	return utils.ResponseSuccessModalDirect(c, "Loading Complete", "The loading test has completed successfully.")
}

func SucceessModalTest(c *fiber.Ctx) error {

	time.Sleep(2 * time.Second)

	errorData := false
	// if error happen here, so the swap will be the error modal
	if errorData {
		return utils.ResponseErrorModal(c, "Operation Failed", "There was an error completing the operation.")
	}

	// return adaptor.HTTPHandler(templ.Handler(successModal))(c)
	componentsReturn := components.SuccessAddCountries()
	// componentsReturn := components.CountryFormModal("asd", "/api/countries", models.Country{})

	return adaptor.HTTPHandler(templ.Handler(componentsReturn))(c)
}

func ErrorModalTest(c *fiber.Ctx) error {
	time.Sleep(5 * time.Second)

	return utils.ResponseErrorModal(c, "Operation Failed", "There was an error completing the operation.")
}
