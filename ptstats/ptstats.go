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

// Generic address that will be automatically populated by json.Unmarshal() when
// decoding a JSON Patient object in a POST/PUT body.
type Address struct {
	Line       []string
	City       string
	State      string
	PostalCode string
}

// Generic patient that will be automatically populated by json.Unmarshal() when
// decoding a JSON Patient object in a PUT/POST body.
type Patient struct {
	Id      string
	Gender  string
	Address []Address
}

// CountySubdivision is a GORM model that maps to the "tiger.cousub" table.
type CountySubdivision struct {
	CosbidFp string `gorm:"column:cosbidfp;primary_key"`
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
	CountyFp                string  `gorm:"column:ct_fips;primary_key"`
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
	CountyFp                string  `gorm:"column:ct_fips;primary_key"`
	CousubFp                string  `gorm:"column:cs_fips;primary_key"`
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
	da.DB.Model(&county).Update(SyntheticCountyStatistics{
		Population:              county.Population,
		PopulationMale:          county.PopulationMale,
		PopulationPerSquareMile: county.PopulationPerSquareMile,
	})
}

func (da PgSyntheticCountyStatsDataAccess) AddFemale(countyFp string) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population += 1
	county.PopulationFemale += 1
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles
	da.DB.Model(&county).Update(SyntheticCountyStatistics{
		Population:              county.Population,
		PopulationFemale:        county.PopulationFemale,
		PopulationPerSquareMile: county.PopulationPerSquareMile,
	})
}

func (da PgSyntheticCountyStatsDataAccess) RemoveMale(countyFp string) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population -= 1
	county.PopulationMale -= 1
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles
	da.DB.Model(&county).Update(SyntheticCountyStatistics{
		Population:              county.Population,
		PopulationMale:          county.PopulationMale,
		PopulationPerSquareMile: county.PopulationPerSquareMile,
	})
}

func (da PgSyntheticCountyStatsDataAccess) RemoveFemale(countyFp string) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population -= 1
	county.PopulationFemale -= 1
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles
	da.DB.Model(&county).Update(SyntheticCountyStatistics{
		Population:              county.Population,
		PopulationFemale:        county.PopulationFemale,
		PopulationPerSquareMile: county.PopulationPerSquareMile,
	})
}

// SyntheticCountyStatsDataAccess provides an interface for querying and updating
// statistics for a given county subdivision (town). Note: this table has a composite
// primary key so both countyFp and cousubFp are required for queries.
type SyntheticCountySubdivisionStatsDataAccess interface {
	GetPopulation(countyFp string, cousubFp string) int64
	GetMalePopulation(countyFp string, cousubFp string) int64
	GetFemalePopulation(countyFp string, cousubFp string) int64
	GetPopulationPerSquareMile(countyFp string, cousubFp string) float64

	AddMale(countyFp string, cousubFp string)
	AddFemale(countyFp string, cousubFp string)
	RemoveMale(countyFp string, cousubFp string)
	RemoveFemale(countyFp string, cousubFp string)
}

// PgSyntheticCountySubdivisionStatsDataAccess implements the SyntheticCountySubdivisionStatsDataAccess
// using a Postgres dataSbase connection and a GORM model for SyntheticCountySubdivisionStatistics.
type PgSyntheticCountySubdivisionStatsDataAccess struct {
	DB *gorm.DB
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetPopulation(countyFp string, cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.Population
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetMalePopulation(countyFp string, cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationMale
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetFemalePopulation(countyFp string, cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationFemale
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) GetPopulationPerSquareMile(countyFp string, cousubFp string) float64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationPerSquareMile
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) AddMale(countyFp string, cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	cousub.Population += 1
	cousub.PopulationMale += 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Model(&cousub).Update(SyntheticCountySubdivisionStatistics{
		Population:              cousub.Population,
		PopulationMale:          cousub.PopulationMale,
		PopulationPerSquareMile: cousub.PopulationPerSquareMile,
	})
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) AddFemale(countyFp string, cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	cousub.Population += 1
	cousub.PopulationFemale += 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Model(&cousub).Update(SyntheticCountySubdivisionStatistics{
		Population:              cousub.Population,
		PopulationFemale:        cousub.PopulationFemale,
		PopulationPerSquareMile: cousub.PopulationPerSquareMile,
	})
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) RemoveMale(countyFp string, cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	cousub.Population -= 1
	cousub.PopulationMale -= 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Model(&cousub).Update(SyntheticCountySubdivisionStatistics{
		Population:              cousub.Population,
		PopulationMale:          cousub.PopulationMale,
		PopulationPerSquareMile: cousub.PopulationPerSquareMile,
	})
}

func (da PgSyntheticCountySubdivisionStatsDataAccess) RemoveFemale(countyFp string, cousubFp string) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	cousub.Population -= 1
	cousub.PopulationFemale -= 1
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles
	da.DB.Model(&cousub).Update(SyntheticCountySubdivisionStatistics{
		Population:              cousub.Population,
		PopulationFemale:        cousub.PopulationFemale,
		PopulationPerSquareMile: cousub.PopulationPerSquareMile,
	})
}

// Middleware that handles the interceptor
type PtStatsInterceptor struct {
	CousubDA           CountySubdivisionDataAccess
	SynthCountyStatsDA SyntheticCountyStatsDataAccess
	SynthCousubStatsDA SyntheticCountySubdivisionStatsDataAccess
}

func UpdatePatientStats(s *PtStatsInterceptor, c *gin.Context) {

	// Read the body and close the stream
	body, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body.Close()

	// Parse the patient from the request body
	var patient Patient
	err := json.Unmarshal(body, &patient)

	if err != nil {
		log.Printf("ptstats: %s", err.Error())
		return
	}

	// We need to replenish the body since we drained the stream
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

	city := patient.Address[0].City
	gender := patient.Gender

	switch {
	case city == "":
		log.Printf("ptstats: No patient city in request body")
		return

	case gender == "":
		log.Printf("ptstats: No patient gender in request body")
		return
	}

	// Update Subdivision and County statistics
	cousubFp := s.CousubDA.GetCountySubdivisionFp(city)
	if cousubFp == "00000" || cousubFp == "" {
		log.Printf("ptstats: City %s does not exist", city)
		return
	}
	countyFp := s.CousubDA.GetCountyFp(cousubFp)

	switch c.Request.Method {
	case "POST":
		switch gender {
		case "male":
			s.SynthCousubStatsDA.AddMale(countyFp, cousubFp)
			s.SynthCountyStatsDA.AddMale(countyFp)
		case "female":
			s.SynthCousubStatsDA.AddFemale(countyFp, cousubFp)
			s.SynthCountyStatsDA.AddFemale(countyFp)
		}

	case "DELETE":
		switch gender {
		case "male":
			s.SynthCousubStatsDA.RemoveMale(countyFp, cousubFp)
			s.SynthCountyStatsDA.RemoveMale(countyFp)
		case "female":
			s.SynthCousubStatsDA.RemoveFemale(countyFp, cousubFp)
			s.SynthCountyStatsDA.RemoveFemale(countyFp)
		}
	}
}

// Handler is registered with the GoFHIR server and invoked on every request
func (s *PtStatsInterceptor) Handler(c *gin.Context) {

	// This intereptor is only needed for Create or Delete operations on a Patient
	if c.Request != nil &&
		c.Request.URL.Path == "/Patient" &&
		(c.Request.Method == "POST" || c.Request.Method == "DELETE") {

		UpdatePatientStats(s, c)
	}

	// Go to the next handler
	c.Next()
}
