/*
Package ptstats implements an interceptor to update patient statistics
for a given county or county subdivision (town).

Carlton Duffett
*/
package ptstats

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Generic address that will be automatically populated by json.Unmarshall() when
// decoding a JSON Patient object in a POST/PUT body.
type Address struct {
	Line       []string
	City       string
	State      string
	PostalCode string
}

// Generic patient that will be automatically populated by json.Unmarshall() when
// decoding a JSON Patient object in a PUT/POST body.
type Patient struct {
	Id      string
	Gender  string
	Address []Address
}

// CountySubdivision is a GORM model that maps to the "tiger.cousub" table.
type CountySubdivision struct {
	StateFp  string `gorm:"column:statefp"`
	CountyFp string `gorm:"column:countyfp"`
	CousubFp string `gorm:"column:cousubfp"`
	Name     string `gorm:"column:name"`
}

func (CountySubdivision) TableName() string {
	return "tiger.cousub"
}

// SyntheticCountyStatistics is a GORM model that maps to the "synth_ma.synth_county_stats" table.
type SyntheticCountyStatistics struct {
	CountyName              string  `gorm:"column:ct_name"`
	CountyFp                string  `gorm:"column:ct_fips"`
	SquareMiles             float64 `gorm:"column:sq_mi"`
	Population              int64   `gorm:"column:pop"`
	PopulationMale          int64   `gorm:"column:pop_male"`
	PopulationFemale        int64   `gorm:"column:pop_female"`
	PopulationPerSquareMile float64 `gorm:"column:pop_sm"`
}

func (SyntheticCountyStatistics) TableName() string {
	return "synth_ma.synth_county_stats"
}

// SyntheticCountySubdivisionStatistics is a GORM model that maps to the
// "synth_ma.synth_cousub_stats" table.
type SyntheticCountySubdivisionStatistics struct {
	CountyName              string  `gorm:"column:ct_name"`
	CountyFp                string  `gorm:"column:ct_fips"`
	CousubFp                string  `gorm:"column:cs_fips"`
	CountySubdivisionName   string  `gorm:"column:"cs_name"`
	SquareMiles             float64 `gorm:"column:sq_mi"`
	Population              int64   `gorm:"column:pop"`
	PopulationMale          int64   `gorm:"column:pop_male"`
	PopulationFemale        int64   `gorm:"column:pop_female"`
	PopulationPerSquareMile float64 `gorm:"column:pop_sm"`
}

func (SyntheticCountySubdivisionStatistics) TableName() string {
	return "synth_ma.synth_cousub_stats"
}

// CountySubdivsionDataAccess provides an interface for querying county subdivision information.
type CountySubdivisionDataAccess interface {
	GetCountySubdivisionFp(city string) string
	GetCountyFp(cousubFp string) string
	GetStateFp(countyFp string) string
}

// PgSyntheticCountyStatsDataAccess implements the CountySubdivisionDataAccess interface
// using a Postgres database connection and a GORM model for CountySubdivision.
type PgCountySubdivisionDataAccess struct {
	DB *gorm.DB
}

func (da PgCountySubdivisionDataAccess) GetCountySubdivisionFp(city string) string {
	var cousub CountySubdivision
	da.DB.Where(&CountySubdivision{Name: city}).First(&cousub)
	return cousub.CousubFp
}

func (da PgCountySubdivisionDataAccess) GetCountyFp(cousubFp string) string {
	var cousub CountySubdivision
	da.DB.Where(&CountySubdivision{CousubFp: cousubFp}).First(&cousub)
	return cousub.CountyFp
}

func (da PgCountySubdivisionDataAccess) GetStateFp(countyFp string) string {
	var cousub CountySubdivision
	da.DB.Where(&CountySubdivision{CountyFp: countyFp}).First(&cousub)
	return cousub.StateFp
}

// SyntheticCountyStatsDataAccess provides an interface for querying and updating
// statistics for a given county.
type SyntheticCountyStatsDataAccess interface {
	GetPopulation(countyFp string) int64
	GetMalePopulation(countyFp string) int64
	GetFemalePopulation(countyFp string) int64
	GetPopulationPerSquareMile(countyFp string) float64

	AddMale(countyFp string)
	AddFemale(countyFp string)
	RemoveMale(countyFp string)
	RemoveFemale(countyFp string)
}

// PgSyntheticCountyStatsDataAccess implements the SyntheticCountyStatsDataAccess
// using a Postgres database connection and a GORM model for SyntheticCountyStatistics.
type PgSyntheticCountyStatsDataAccess struct {
	DB *gorm.DB
}

func (da PgSyntheticCountyStatsDataAccess) GetPopulation(countyFp string) int64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.Population
}

func (da PgSyntheticCountyStatsDataAccess) GetMalePopulation(countyFp string) int64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.PopulationMale
}

func (da PgSyntheticCountyStatsDataAccess) GetFemalePopulation(countyFp string) int64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.PopulationFemale
}

func (da PgSyntheticCountyStatsDataAccess) GetPopulationPerSquareMile(countyFp string) float64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.PopulationPerSquareMile
}

func (da PgSyntheticCountyStatsDataAccess) AddMale(countyFp string) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population += 1
	county.PopulationMale += 1
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles
	da.DB.Save(&county)
}

func (da PgSyntheticCountyStatsDataAccess) AddFemale(countyFp string) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population += 1
	county.PopulationFemale += 1
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles
	da.DB.Save(&county)
}

func (da PgSyntheticCountyStatsDataAccess) RemoveMale(countyFp string) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population -= 1
	county.PopulationMale -= 1
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles
	da.DB.Save(&county)
}

func (da PgSyntheticCountyStatsDataAccess) RemoveFemale(countyFp string) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population -= 1
	county.PopulationFemale -= 1
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles
	da.DB.Save(&county)
}

// SyntheticCountyStatsDataAccess provides an interface for querying and updating
// statistics for a given county subdivision (town).
type SyntheticCountySubdivisionStatsDataAccess interface {
	GetPopulation(cousubFp string) int64
	GetMalePopulation(cousubFp string) int64
	GetFemalePopulation(cousubFp string) int64
	GetPopulationPerSquareMile(cousubFp string) float64

	AddMale(cousubFp string)
	AddFemale(cousubFp string)
	RemoveMale(cousubFp string)
	RemoveFemale(cousubFp string)
}

// PgSyntheticCountySubdivisionStatsDataAccess implements the SyntheticCountySubdivisionStatsDataAccess
// using a Postgres dataSbase connection and a GORM model for SyntheticCountySubdivisionStatistics.
type PgSyntheticCountySubdivisionStatsDataAccess struct {
	DB *gorm.DB
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetPopulation(cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	return cousub.Population
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetMalePopulation(cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationMale
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetFemalePopulation(cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationFemale
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetPopulationPerSquareMile(cousubFp string) float64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationPerSquareMile
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) AddMale(cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	cousub.Population += 1
	cousub.PopulationMale += 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Save(&cousub)
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) AddFemale(cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	cousub.Population += 1
	cousub.PopulationFemale += 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Save(&cousub)
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) RemoveMale(cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	cousub.Population -= 1
	cousub.PopulationMale -= 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Save(&cousub)
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) RemoveFemale(cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CousubFp: cousubFp}).First(&cousub)
	cousub.Population -= 1
	cousub.PopulationFemale -= 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Save(&cousub)
}

// Middleware that handles the interceptor
type PtStatsInterceptor struct {
	CousubDA           CountySubdivisionDataAccess
	SynthCountyStatsDA SyntheticCountyStatsDataAccess
	SynthCousubStatsDA SyntheticCountySubdivisionStatsDataAccess
}

// Handler is registered with the GoFHIR server and invoked on every request
func (s *PtStatsInterceptor) Handler(c *gin.Context) {

	// Only handle Create, Update, and Delete operations for a Patient
	if c.Request != nil &&
		c.Request.URL.Path == "/Patient" &&
		(c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE") {

		// Read the body and close the stream
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()

		// Parse the patient from the request body
		var patient Patient
		err := json.Unmarshal(body, &patient)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("city: %s", patient.Address[0].City)

		// We need to replenish the body since we drained the stream
		c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	// Go to the next handler
	c.Next()
}
