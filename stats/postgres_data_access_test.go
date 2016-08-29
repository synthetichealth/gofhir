package stats

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/intervention-engine/fhir/models"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

const (
	DiabetesSnomedCode = "44054006"
	BostonCountyfp     = "025"
	BostonCousubfp     = "07000"
)

type Stats struct {
	Pop, PopMale, PopFemale int64
	PopPerSqMile, SqMiles   float64
}

type StatsTestSuite struct {
	suite.Suite
	db *sql.DB
	da StatsDataAccess
}

func (s *StatsTestSuite) SetupSuite() {
	// open a postgres database connection to the test database
	db, err := sql.Open("postgres", "postgres://fhir_test:fhir_test@localhost/fhir_test?sslmode=disable")

	if err != nil {
		log.Println("Before testing please setup the fhir_test database using pgsetup.sh")
		log.Fatal(err)
	}

	// ping the db to ensure we connected successfully
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	s.db = db

	// create new data access layer for testing
	s.da = NewPgStatsDataAccess(db)
}

func (s *StatsTestSuite) TearDownSuite() {
	// close the db connection after testing is finished
	s.db.Close()
}

func TestStatsSuite(t *testing.T) {
	// bind test suite to go test
	suite.Run(t, new(StatsTestSuite))
}

func (s *StatsTestSuite) TestAddMalePatientStat() {

	// City of Boston
	// countyfp: 025
	// cousubfp: 07000

	var err error
	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats

	// Get initial stats for comparison
	countyStats, _ = s.getCountyStats(BostonCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BostonCousubfp)

	// Add a male patient
	patient := createPatient("Boston", "male")
	s.NotPanics(func() { err = s.da.AddPatientStat(patient) }, "AddPatientStat should not panic for a valid male patient")
	s.Nil(err, "AddPatientStat should not fail for a valid male patient")

	// Check that the relevant statistics updated
	updatedCountyStats, _ = s.getCountyStats(BostonCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, 1, 0)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, 1, 0)
}

func (s *StatsTestSuite) TestAddFemalePatientStat() {

	var err error
	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats

	// Get initial stats for comparison
	countyStats, _ = s.getCountyStats(BostonCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BostonCousubfp)

	// Add a female patient
	patient := createPatient("Boston", "female")
	s.NotPanics(func() { err = s.da.AddPatientStat(patient) }, "AddPatientStat should not panic for a valid male patient")
	s.Nil(err, "AddPatientStat should not fail for a valid male patient")

	// Check that the relevant statistics updated
	updatedCountyStats, _ = s.getCountyStats(BostonCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, 0, 1)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, 0, 1)
}

func (s *StatsTestSuite) TestRemoveMalePatientStat() {

	var err error
	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats

	// Get initial stats for comparison
	countyStats, _ = s.getCountyStats(BostonCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BostonCousubfp)

	// Remove a male patient
	patient := createPatient("Boston", "male")
	s.NotPanics(func() { err = s.da.RemovePatientStat(patient) }, "AddPatientStat should not panic for a valid male patient")
	s.Nil(err, "AddPatientStat should not fail for a valid male patient")

	// Check that the relevant statistics updated
	updatedCountyStats, _ = s.getCountyStats(BostonCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, -1, 0)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, -1, 0)
}

func (s *StatsTestSuite) TestRemoveFemalePatientStat() {

	var err error
	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats

	// Get initial stats for comparison
	countyStats, _ = s.getCountyStats(BostonCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BostonCousubfp)

	// Remove a female patient
	patient := createPatient("Boston", "female")
	s.NotPanics(func() { err = s.da.RemovePatientStat(patient) }, "AddPatientStat should not panic for a valid male patient")
	s.Nil(err, "AddPatientStat should not fail for a valid male patient")

	// Check that the relevant statistics updated
	updatedCountyStats, _ = s.getCountyStats(BostonCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, 0, -1)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, 0, -1)
}

func (s *StatsTestSuite) TestAddPatientStatInvalidGender() {

	var err error
	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats

	// Test invalid gender
	patient := createPatient("Boston", "foo")
	countyStats, _ = s.getCountyStats(BostonCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.NotPanics(func() { err = s.da.AddPatientStat(patient) }, "AddPatientStat should not panic for an invalid patient gender")
	s.NotNil(err, "AddPatientStat should fail for an invalid gender")
	updatedCountyStats, _ = s.getCountyStats(BostonCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, 0, 0)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, 0, 0)
}

func (s *StatsTestSuite) TestRemovePatientStatInvalidGender() {

	var err error
	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats

	// Test invalid gender
	patient := createPatient("Boston", "foo")
	countyStats, _ = s.getCountyStats(BostonCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.NotPanics(func() { err = s.da.RemovePatientStat(patient) }, "AddPatientStat should not panic for an invalid patient gender")
	s.NotNil(err, "AddPatientStat should fail for an invalid gender")
	updatedCountyStats, _ = s.getCountyStats(BostonCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, 0, 0)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, 0, 0)
}

func (s *StatsTestSuite) TestAddAndRemovePatientStatInvalidCity() {

	var err error

	// Test invalid city
	patient := createPatient("Bar", "male")
	s.NotPanics(func() { err = s.da.AddPatientStat(patient) }, "AddPatientStat should not panic for an invalid city")
	s.NotNil(err, "AddPatientStat should fail for an invalid city")
	s.NotPanics(func() { err = s.da.RemovePatientStat(patient) }, "RemovePatientStat should not panic for an invalid city")
	s.NotNil(err, "RemovePatientStat should fail for an invalid city")
}

func (s *StatsTestSuite) TestConditionStats() {

}

func (s *StatsTestSuite) getCountyStats(countyfp string) (stats Stats, err error) {
	query := "SELECT pop, pop_male, pop_female, pop_sm, sq_mi FROM synth_ma.synth_county_stats WHERE ct_fips = $1"
	err = s.db.QueryRow(query, countyfp).Scan(&stats.Pop, &stats.PopMale, &stats.PopFemale, &stats.PopPerSqMile, &stats.SqMiles)
	return
}

func (s *StatsTestSuite) getSubdivisionStats(cousubfp string) (stats Stats, err error) {
	query := "SELECT pop, pop_male, pop_female, pop_sm, sq_mi FROM synth_ma.synth_cousub_stats WHERE cs_fips = $1"
	err = s.db.QueryRow(query, cousubfp).Scan(&stats.Pop, &stats.PopMale, &stats.PopFemale, &stats.PopPerSqMile, &stats.SqMiles)
	return
}

func (s *StatsTestSuite) assertCountyStatsChanged(county, ucounty Stats, maleDelta, femaleDelta int64) {
	s.Equal(county.Pop+(maleDelta+femaleDelta), ucounty.Pop, fmt.Sprintf("County population should change by %d", maleDelta+femaleDelta))
	s.Equal(county.PopMale+maleDelta, ucounty.PopMale, fmt.Sprintf("County male population should change by %d", maleDelta))
	s.Equal(county.PopFemale+femaleDelta, ucounty.PopFemale, fmt.Sprintf("County female population should change by %d", femaleDelta))
	newPopPerSqMile := float64(county.Pop+(maleDelta+femaleDelta)) / county.SqMiles
	s.Equal(newPopPerSqMile, ucounty.PopPerSqMile, fmt.Sprintf("County PopPerSqMile should now be %.8f", newPopPerSqMile))
}

func (s *StatsTestSuite) assertSubdivisionStatsChanged(cousub, ucousub Stats, maleDelta, femaleDelta int64) {
	s.Equal(cousub.Pop+(maleDelta+femaleDelta), ucousub.Pop, fmt.Sprintf("Subdivision population should change by %d", maleDelta+femaleDelta))
	s.Equal(cousub.PopMale+maleDelta, ucousub.PopMale, fmt.Sprintf("Subdivision male population should change by %d", maleDelta))
	s.Equal(cousub.PopFemale+femaleDelta, ucousub.PopFemale, fmt.Sprintf("Subivision female population should change by %d", femaleDelta))
	newPopPerSqMile := float64(cousub.Pop+(maleDelta+femaleDelta)) / cousub.SqMiles
	s.Equal(newPopPerSqMile, ucousub.PopPerSqMile, fmt.Sprintf("Subdivision PopPerSqMile should now be %.8f", newPopPerSqMile))
}

func createPatient(city, gender string) *models.Patient {

	return &models.Patient{
		Gender: gender,
		Address: []models.Address{
			models.Address{
				City:  city,
				State: "MA",
			},
		},
	}
}

func createCondition(snomedCode string) *models.Condition {

	return &models.Condition{
		Code: &models.CodeableConcept{
			Coding: []models.Coding{
				models.Coding{
					Code:   snomedCode,
					System: SnomedCodeSystem,
				},
			},
		},
	}
}
