package stats

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type PatientInterceptorTestSuite struct {
	suite.Suite
	db *sql.DB
	da StatsDataAccess
}

type UnkownResource struct{}

func (s *PatientInterceptorTestSuite) SetupSuite() {
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

func (s *PatientInterceptorTestSuite) TearDownSuite() {

	// zero out the stats for the next test
	_, _ = s.db.Query("UPDATE synth_ma.synth_county_stats SET pop = 0, pop_male = 0, pop_female = 0, pop_sm = 0;")
	_, _ = s.db.Query("UPDATE synth_ma.synth_cousub_stats SET pop = 0, pop_male = 0, pop_female = 0, pop_sm = 0;")
	_, _ = s.db.Query("UPDATE synth_ma.synth_county_facts SET pop = 0, pop_male = 0, pop_female = 0, rate = 0;")
	_, _ = s.db.Query("UPDATE synth_ma.synth_cousub_facts SET pop = 0, pop_male = 0, pop_female = 0, rate = 0;")

	// close the db connection after testing is finished
	s.db.Close()
}

func TestPatientInterceptors(t *testing.T) {
	// bind test suite to go test
	suite.Run(t, new(PatientInterceptorTestSuite))
}

func (s *PatientInterceptorTestSuite) TestPatientCreateInterceptor() {

	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats

	// Get initial stats for comparison
	countyStats, _ = s.getCountyStats(BostonCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BostonCousubfp)

	// Add a patient
	patient := createPatient("Boston", "male")
	createInterceptor := NewPatientStatsCreateInterceptor(s.da)
	s.NotPanics(func() { createInterceptor.After(patient) }, "Should not panic for valid patient")

	// Check that the relevant statistics updated
	updatedCountyStats, _ = s.getCountyStats(BostonCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BostonCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, 1, 0)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, 1, 0)

	// Check handling of non-patient resources (fails silently)
	s.NotPanics(func() { createInterceptor.After(UnkownResource{}) }, "Should not panic for non-patient resource")
}

func (s *PatientInterceptorTestSuite) TestPatientDeleteInterceptor() {

	var countyStats, updatedCountyStats, cousubStats, updatedCousubStats Stats
	patient := createPatient("Bedford", "female")
	deleteInterceptor := NewPatientStatsDeleteInterceptor(s.da)

	countyStats, _ = s.getCountyStats(BedfordCountyfp)
	cousubStats, _ = s.getSubdivisionStats(BedfordCousubfp)

	s.NotPanics(func() { deleteInterceptor.After(patient) }, "Should not panic for valid patient")

	// Check that the stats changed
	updatedCountyStats, _ = s.getCountyStats(BedfordCountyfp)
	updatedCousubStats, _ = s.getSubdivisionStats(BedfordCousubfp)
	s.assertCountyStatsChanged(countyStats, updatedCountyStats, 0, -1)
	s.assertSubdivisionStatsChanged(cousubStats, updatedCousubStats, 0, -1)

	s.NotPanics(func() { deleteInterceptor.After(UnkownResource{}) }, "Should not panic for non-patient resource")
}

func (s *PatientInterceptorTestSuite) getCountyStats(countyfp string) (stats Stats, err error) {
	query := "SELECT pop, pop_male, pop_female, pop_sm, sq_mi FROM synth_ma.synth_county_stats WHERE ct_fips = $1"
	err = s.db.QueryRow(query, countyfp).Scan(&stats.Pop, &stats.PopMale, &stats.PopFemale, &stats.PopPerSqMile, &stats.SqMiles)
	return
}

func (s *PatientInterceptorTestSuite) getSubdivisionStats(cousubfp string) (stats Stats, err error) {
	query := "SELECT pop, pop_male, pop_female, pop_sm, sq_mi FROM synth_ma.synth_cousub_stats WHERE cs_fips = $1"
	err = s.db.QueryRow(query, cousubfp).Scan(&stats.Pop, &stats.PopMale, &stats.PopFemale, &stats.PopPerSqMile, &stats.SqMiles)
	return
}

func (s *PatientInterceptorTestSuite) assertCountyStatsChanged(county, ucounty Stats, maleDelta, femaleDelta int64) {
	s.Equal(county.Pop+(maleDelta+femaleDelta), ucounty.Pop, fmt.Sprintf("County population should change by %d", maleDelta+femaleDelta))
	s.Equal(county.PopMale+maleDelta, ucounty.PopMale, fmt.Sprintf("County male population should change by %d", maleDelta))
	s.Equal(county.PopFemale+femaleDelta, ucounty.PopFemale, fmt.Sprintf("County female population should change by %d", femaleDelta))
	newPopPerSqMile := float64(county.Pop+(maleDelta+femaleDelta)) / county.SqMiles
	s.Equal(newPopPerSqMile, ucounty.PopPerSqMile, fmt.Sprintf("County PopPerSqMile should now be %.8f", newPopPerSqMile))
}

func (s *PatientInterceptorTestSuite) assertSubdivisionStatsChanged(cousub, ucousub Stats, maleDelta, femaleDelta int64) {
	s.Equal(cousub.Pop+(maleDelta+femaleDelta), ucousub.Pop, fmt.Sprintf("Subdivision population should change by %d", maleDelta+femaleDelta))
	s.Equal(cousub.PopMale+maleDelta, ucousub.PopMale, fmt.Sprintf("Subdivision male population should change by %d", maleDelta))
	s.Equal(cousub.PopFemale+femaleDelta, ucousub.PopFemale, fmt.Sprintf("Subivision female population should change by %d", femaleDelta))
	newPopPerSqMile := float64(cousub.Pop+(maleDelta+femaleDelta)) / cousub.SqMiles
	s.Equal(newPopPerSqMile, ucousub.PopPerSqMile, fmt.Sprintf("Subdivision PopPerSqMile should now be %.8f", newPopPerSqMile))
}
