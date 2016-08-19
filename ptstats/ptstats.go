/*
Package ptstats implements an interceptor to update patient statistics
for a given county or county subdivision (town).

Carlton Duffett
*/
package ptstats

import (
	"errors"
	"fmt"
	"github.com/intervention-engine/fhir/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
)

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

func (da *PgCountySubdivisionDataAccess) GetCountySubdivisionFp(city string) string {
	var cousub CountySubdivision
	da.DB.Where(&CountySubdivision{Name: city}).First(&cousub)
	return cousub.CousubFp
}

func (da *PgCountySubdivisionDataAccess) GetCountyFp(cousubFp string) string {
	var cousub CountySubdivision
	da.DB.Where(&CountySubdivision{CousubFp: cousubFp}).First(&cousub)
	return cousub.CountyFp
}

func (da *PgCountySubdivisionDataAccess) GetStateFp(countyFp string) string {
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

func (da *PgSyntheticCountyStatsDataAccess) GetPopulation(countyFp string) int64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.Population
}

func (da *PgSyntheticCountyStatsDataAccess) GetMalePopulation(countyFp string) int64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.PopulationMale
}

func (da *PgSyntheticCountyStatsDataAccess) GetFemalePopulation(countyFp string) int64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.PopulationFemale
}

func (da *PgSyntheticCountyStatsDataAccess) GetPopulationPerSquareMile(countyFp string) float64 {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	return county.PopulationPerSquareMile
}

func (da *PgSyntheticCountyStatsDataAccess) AddMale(countyFp string) {
	da.modifyPopulationCount(countyFp, 1, 0)
}

func (da *PgSyntheticCountyStatsDataAccess) AddFemale(countyFp string) {
	da.modifyPopulationCount(countyFp, 0, 1)
}

func (da *PgSyntheticCountyStatsDataAccess) RemoveMale(countyFp string) {
	da.modifyPopulationCount(countyFp, -1, 0)
}

func (da *PgSyntheticCountyStatsDataAccess) RemoveFemale(countyFp string) {
	da.modifyPopulationCount(countyFp, 0, -1)
}

func (da *PgSyntheticCountyStatsDataAccess) modifyPopulationCount(countyFp string, maleDelta, femaleDelta int64) {
	var county SyntheticCountyStatistics
	da.DB.Where(&SyntheticCountyStatistics{CountyFp: countyFp}).First(&county)
	county.Population += (maleDelta + femaleDelta)
	county.PopulationMale += maleDelta
	county.PopulationFemale += femaleDelta
	county.PopulationPerSquareMile = float64(county.Population) / county.SquareMiles

	// See http://jinzhu.me/gorm/curd.html#update-changed-fields for potential issues with Struct-based updates
	da.DB.Model(&county).Update(map[string]interface{}{
		"pop":        county.Population,
		"pop_male":   county.PopulationMale,
		"pop_female": county.PopulationFemale,
		"pop_sm":     county.PopulationPerSquareMile,
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

func (da *PgSyntheticCountySubdivisionStatsDataAccess) GetPopulation(countyFp string, cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.Population
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) GetMalePopulation(countyFp string, cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationMale
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) GetFemalePopulation(countyFp string, cousubFp string) int64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationFemale
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) GetPopulationPerSquareMile(countyFp string, cousubFp string) float64 {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	return cousub.PopulationPerSquareMile
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) AddMale(countyFp string, cousubFp string) {
	da.modifyPopulationCount(countyFp, cousubFp, 1, 0)
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) AddFemale(countyFp string, cousubFp string) {
	da.modifyPopulationCount(countyFp, cousubFp, 0, 1)
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) RemoveMale(countyFp string, cousubFp string) {
	da.modifyPopulationCount(countyFp, cousubFp, -1, 0)
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) RemoveFemale(countyFp string, cousubFp string) {
	da.modifyPopulationCount(countyFp, cousubFp, 0, -1)
}

func (da *PgSyntheticCountySubdivisionStatsDataAccess) modifyPopulationCount(countyFp, cousubFp string, maleDelta, femaleDelta int64) {
	var cousub SyntheticCountySubdivisionStatistics
	da.DB.Where(&SyntheticCountySubdivisionStatistics{CountyFp: countyFp, CousubFp: cousubFp}).First(&cousub)
	cousub.Population += (maleDelta + femaleDelta)
	cousub.PopulationMale += maleDelta
	cousub.PopulationFemale += femaleDelta
	cousub.PopulationPerSquareMile = float64(cousub.Population) / cousub.SquareMiles

	// See http://jinzhu.me/gorm/curd.html#update-changed-fields for potential issues with Struct-based updates
	da.DB.Model(&cousub).Update(map[string]interface{}{
		"pop":        cousub.Population,
		"pop_male":   cousub.PopulationMale,
		"pop_female": cousub.PopulationFemale,
		"pop_sm":     cousub.PopulationPerSquareMile,
	})
}

// PatientStatsDataAccess provides a common high-level interface to each of the
// interceptor handlers for modifying patient statistics in the Postgres database
type PatientStatsDataAccess struct {
	CountyStats SyntheticCountyStatsDataAccess
	CousubStats SyntheticCountySubdivisionStatsDataAccess
	Cousub      CountySubdivisionDataAccess
}

// NewPatientStatsDataAccess creates a new patient stats data access interface
// for use by one or more patient statistics interceptors
func NewPatientStatsDataAccess(db *gorm.DB) *PatientStatsDataAccess {
	return &PatientStatsDataAccess{
		CountyStats: &PgSyntheticCountyStatsDataAccess{DB: db},
		CousubStats: &PgSyntheticCountySubdivisionStatsDataAccess{DB: db},
		Cousub:      &PgCountySubdivisionDataAccess{DB: db},
	}
}

func (da *PatientStatsDataAccess) IdentifyCountyAndSubdivision(patient *models.Patient) (countyFp, cousubFp string, err error) {
	city := patient.Address[0].City
	if city == "" {
		return "", "", errors.New("IdentifyCountyAndSubdivision: No city found in patient's address")
	}

	cousubFp = da.Cousub.GetCountySubdivisionFp(city)
	if cousubFp == "00000" || cousubFp == "" {
		return "", "", errors.New(fmt.Sprintf("IdentifyCountyAndSubdivision: City %s does not exist", city))
	}

	countyFp = da.Cousub.GetCountyFp(cousubFp)
	return countyFp, cousubFp, nil
}

func (da *PatientStatsDataAccess) AddMale(patient *models.Patient) error {

	countyFp, cousubFp, err := da.IdentifyCountyAndSubdivision(patient)

	if err != nil {
		return err
	}

	if patient.Gender == "male" {
		da.CousubStats.AddMale(countyFp, cousubFp)
		da.CountyStats.AddMale(countyFp)
		return nil
	} else {
		return errors.New("AddMale: Patient is not a male")
	}
}

func (da *PatientStatsDataAccess) AddFemale(patient *models.Patient) error {

	countyFp, cousubFp, err := da.IdentifyCountyAndSubdivision(patient)

	if err != nil {
		return err
	}

	if patient.Gender == "female" {

		da.CousubStats.AddFemale(countyFp, cousubFp)
		da.CountyStats.AddFemale(countyFp)
		return nil
	} else {
		return errors.New("AddFemale: Patient is not a female")
	}
}

func (da *PatientStatsDataAccess) RemoveMale(patient *models.Patient) error {

	countyFp, cousubFp, err := da.IdentifyCountyAndSubdivision(patient)

	if err != nil {
		return err
	}

	if patient.Gender == "male" {
		da.CousubStats.RemoveMale(countyFp, cousubFp)
		da.CountyStats.RemoveMale(countyFp)
		return nil
	} else {
		return errors.New("RemoveMale: Patient is not a male")
	}
}

func (da *PatientStatsDataAccess) RemoveFemale(patient *models.Patient) error {

	countyFp, cousubFp, err := da.IdentifyCountyAndSubdivision(patient)

	if err != nil {
		return err
	}

	if patient.Gender == "female" {

		da.CousubStats.RemoveFemale(countyFp, cousubFp)
		da.CountyStats.RemoveFemale(countyFp)
		return nil
	} else {
		return errors.New("RemoveFemale: Patient is not a female")
	}
}

// PatientStatsCreateInterceptor intercepts any new patient resources added to the database
// and adds that patient's statistics to the Synthetic Mass stats
type PatientStatsCreateInterceptor struct {
	DataAccess *PatientStatsDataAccess
}

func (s *PatientStatsCreateInterceptor) After(resource interface{}) {
	patient, ok := resource.(*models.Patient)
	var err error

	if ok {
		gender := patient.Gender
		switch gender {
		case "male":
			err = s.DataAccess.AddMale(patient)
		case "female":
			err = s.DataAccess.AddFemale(patient)
		default:
			log.Printf("PatientStatsCreateInterceptor: Invalid gender for patient %s\n", patient.Id)
		}

		if err != nil {
			log.Printf("PatientStatsCreateInterceptor: Failed to add statistics for patient %s\n", patient.Id)
		}
	}
}

// unused interceptor handlers:
func (s *PatientStatsCreateInterceptor) Before(resource interface{})             {}
func (s *PatientStatsCreateInterceptor) OnError(err error, resource interface{}) {}

// PatientStatsUpdateInterceptor intercepts any updated patient resources
// and updates that patient's statistics in the Synthetic Mass stats
type PatientStatsUpdateInterceptor struct {
	DataAccess *PatientStatsDataAccess
	// The state of the patient before the database update, for comparison after the database update
	patientBefore *models.Patient
	// Tracks if the interceptor failed to cache the patient model before the update
	cacheError error
}

func (s *PatientStatsUpdateInterceptor) Before(resource interface{}) {
	patient, ok := resource.(*models.Patient)

	if ok {
		s.patientBefore = patient
	} else {
		errmsg := "PatientStatsUpdateInterceptor:Before: Failed to cache patient before update\n"
		s.cacheError = errors.New(errmsg)
		log.Printf(errmsg)
	}
}

func (s *PatientStatsUpdateInterceptor) After(resource interface{}) {
	patientAfter, ok := resource.(*models.Patient)
	var removeErr, addErr error

	if ok && s.cacheError == nil {
		// see if the patient's address (or at least, his/her city) changed, and update stats
		if s.patientBefore.Address[0].City != "" && s.patientBefore.Address[0].City != patientAfter.Address[0].City {

			switch patientAfter.Gender {
			case "male":
				removeErr = s.DataAccess.RemoveMale(s.patientBefore)
				addErr = s.DataAccess.AddMale(patientAfter)
			case "female":
				removeErr = s.DataAccess.RemoveFemale(s.patientBefore)
				addErr = s.DataAccess.AddFemale(patientAfter)
			default:
				log.Printf("PatientStatsUpdateInterceptor: Invalid gender for patient %s\n", patientAfter.Id)
			}
		}

		if removeErr != nil || addErr != nil {
			log.Printf("PatientStatsUpdateInterceptor: Failed to update statistics for patient %s\n", patientAfter.Id)
		}
	}
}

// unused interceptor handler:
func (s *PatientStatsUpdateInterceptor) OnError(err error, resource interface{}) {}

// PatientStatsDeleteInterceptor intercepts any deleted patient resources
// and removes that patient's statistics from the Synthetic Mass stats
type PatientStatsDeleteInterceptor struct {
	DataAccess *PatientStatsDataAccess
}

func (s *PatientStatsDeleteInterceptor) After(resource interface{}) {
	patient, ok := resource.(*models.Patient)
	var err error

	if ok {
		switch patient.Gender {
		case "male":
			err = s.DataAccess.RemoveMale(patient)
		case "female":
			err = s.DataAccess.RemoveFemale(patient)
		default:
			log.Printf("PatientStatsDeleteInterceptor: Invalid gender for patient %s\n", patient.Id)
		}

		if err != nil {
			log.Printf("PatientStatsDeleteInterceptor: Failed to remove statistics for patient %s\n", patient.Id)
		}
	}
}

// unused interceptor handlers:
func (s *PatientStatsDeleteInterceptor) Before(resource interface{})             {}
func (s *PatientStatsDeleteInterceptor) OnError(err error, resource interface{}) {}
