package stats

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"
)

type ConditionInterceptorTestSuite struct {
	suite.Suite
	db      *sql.DB
	session *mgo.Session
	da      StatsDataAccess
	mda     server.DataAccessLayer
}

func (s *ConditionInterceptorTestSuite) SetupSuite() {
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

	// The condition interceptors also require a mongodb connection
	s.session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	mdb := s.session.DB("fhir")
	s.mda = server.NewMongoDataAccessLayer(mdb, make(map[string]server.InterceptorList))
}

func (s *ConditionInterceptorTestSuite) TearDownSuite() {
	// zero out the stats for the next test
	_, _ = s.db.Query("UPDATE synth_ma.synth_county_stats SET pop = 0, pop_male = 0, pop_female = 0, pop_sm = 0;")
	_, _ = s.db.Query("UPDATE synth_ma.synth_cousub_stats SET pop = 0, pop_male = 0, pop_female = 0, pop_sm = 0;")
	_, _ = s.db.Query("UPDATE synth_ma.synth_county_facts SET pop = 0, pop_male = 0, pop_female = 0, rate = 0;")
	_, _ = s.db.Query("UPDATE synth_ma.synth_cousub_facts SET pop = 0, pop_male = 0, pop_female = 0, rate = 0;")

	// close the db connection after testing is finished
	s.db.Close()

	// close the mongo connection
	s.session.Close()
}

func TestConditionInterceptors(t *testing.T) {
	// bind test suite to go test
	suite.Run(t, new(ConditionInterceptorTestSuite))
}

func (s *ConditionInterceptorTestSuite) TestConditionCreateInterceptor() {

	var countyFacts, updatedCountyFacts, cousubFacts, updatedCousubFacts Facts

	// Get initial facts for comparison
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)

	// Add a patient to mongo
	patient := createPatient("Boston", "male")
	patientID, _ := s.mda.Post(patient)

	condition := createConditionWithReferencedPatient(DiabetesSnomedCode, patientID)
	createInterceptor := NewConditionStatsCreateInterceptor(s.da, s.mda)
	s.NotPanics(func() { createInterceptor.After(condition) }, "Should not panic for valid condition")

	// Check that the relevant statistics updated
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, 1, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, 1, 0)

	// Add condition without referenced patient
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	condition = createCondition(DiabetesSnomedCode)
	s.NotPanics(func() { createInterceptor.After(condition) }, "Should not panic for a missing patient reference")
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, 0, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, 0, 0)

	// Add condition with abatement
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	condition = createAbatedConditionWithReferencedPatient(DiabetesSnomedCode, patientID)
	s.NotPanics(func() { createInterceptor.After(condition) }, "Should not panic for an abated condition")
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, 0, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, 0, 0)

	// Check handling of non-patient resources (fails silently)
	s.NotPanics(func() { createInterceptor.After(UnkownResource{}) }, "Should not panic for non-patient resource")

	// Delete the referenced patient resource
	_ = s.mda.Delete(patientID, "Patient")
}

func (s *ConditionInterceptorTestSuite) TestConditionUpdateInterceptor() {

	var countyFacts, updatedCountyFacts, cousubFacts, updatedCousubFacts Facts

	// Get initial facts for comparison
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)

	// Update a condition
	patient := createPatient("Boston", "male")
	patientID, _ := s.mda.Post(patient)
	condition := createConditionWithReferencedPatient(DiabetesSnomedCode, patientID)
	updatedCondition := createAbatedConditionWithReferencedPatient(DiabetesSnomedCode, patientID)

	updateInterceptor := NewConditionStatsUpdateInterceptor(s.da, s.mda)
	s.NotPanics(func() { updateInterceptor.Before(condition) }, "Should not panic for valid condition")
	s.NotPanics(func() { updateInterceptor.After(updatedCondition) }, "Should not panic for valid condition")

	// Check that the relevant statistics updated
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, -1, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, -1, 0)

	// Update condition without referenced patient
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	condition = createCondition(DiabetesSnomedCode)
	updatedCondition = createCondition(DiabetesSnomedCode)
	s.NotPanics(func() { updateInterceptor.Before(condition) }, "Should not panic for a missing patient reference")
	s.NotPanics(func() { updateInterceptor.After(updatedCondition) }, "Should not panic for a missing patient reference")
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, 0, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, 0, 0)

	// Update condition that is already abated
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	condition = createAbatedConditionWithReferencedPatient(DiabetesSnomedCode, patientID)
	updatedCondition = createAbatedConditionWithReferencedPatient(DiabetesSnomedCode, patientID)
	s.NotPanics(func() { updateInterceptor.Before(condition) }, "Should not panic for an abated condition")
	s.NotPanics(func() { updateInterceptor.After(updatedCondition) }, "Should not panic for an abated condition")
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, 0, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, 0, 0)

	// Check handling of non-patient resources (fails silently)
	s.NotPanics(func() { updateInterceptor.Before(UnkownResource{}) }, "Should not panic for non-patient resource")
	s.NotPanics(func() { updateInterceptor.After(UnkownResource{}) }, "Should not panic for non-patient resource")

	// Delete the referenced patient resource
	_ = s.mda.Delete(patientID, "Patient")
}

func (s *ConditionInterceptorTestSuite) TestConditionDeleteInterceptor() {

	var countyFacts, updatedCountyFacts, cousubFacts, updatedCousubFacts Facts

	// Get initial facts for comparison
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)

	// Add a patient to mongo
	patient := createPatient("Boston", "male")
	patientID, _ := s.mda.Post(patient)
	condition := createConditionWithReferencedPatient(DiabetesSnomedCode, patientID)
	deleteInterceptor := NewConditionStatsDeleteInterceptor(s.da, s.mda)
	s.NotPanics(func() { deleteInterceptor.After(condition) }, "Should not panic for valid condition")

	// Check that the relevant statistics updated
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, -1, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, -1, 0)

	// Delete condition without referenced patient
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	condition = createCondition(DiabetesSnomedCode)
	s.NotPanics(func() { deleteInterceptor.After(condition) }, "Should not panic for a missing patient reference")
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, 0, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, 0, 0)

	// Delete condition with abatement
	countyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	cousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	condition = createAbatedConditionWithReferencedPatient(DiabetesSnomedCode, patientID)
	s.NotPanics(func() { deleteInterceptor.After(condition) }, "Should not panic for an abated condition")
	updatedCountyFacts, _ = s.getCountyFacts(BostonCountyfp, Diabetesfp)
	updatedCousubFacts, _ = s.getSubdivisionFacts(BostonCousubfp, Diabetesfp)
	s.assertCountyFactsChanged(countyFacts, updatedCountyFacts, 0, 0)
	s.assertSubdivisionFactsChanged(cousubFacts, updatedCousubFacts, 0, 0)

	// Check handling of non-patient resources (fails silently)
	s.NotPanics(func() { deleteInterceptor.After(UnkownResource{}) }, "Should not panic for non-patient resource")

	// Delete the referenced patient resource
	_ = s.mda.Delete(patientID, "Patient")
}

func (s *ConditionInterceptorTestSuite) getCountyFacts(countyfp, diseasefp string) (facts Facts, err error) {
	query := "SELECT pop, pop_male, pop_female FROM synth_ma.synth_county_facts WHERE countyfp = $1 AND diseasefp = $2"
	err = s.db.QueryRow(query, countyfp, diseasefp).Scan(&facts.Pop, &facts.PopMale, &facts.PopFemale)
	return
}

func (s *ConditionInterceptorTestSuite) getSubdivisionFacts(cousubfp, diseasefp string) (facts Facts, err error) {
	query := "SELECT pop, pop_male, pop_female FROM synth_ma.synth_cousub_facts WHERE cousubfp = $1 AND diseasefp = $2"
	err = s.db.QueryRow(query, cousubfp, diseasefp).Scan(&facts.Pop, &facts.PopMale, &facts.PopFemale)
	return
}

func (s *ConditionInterceptorTestSuite) assertCountyFactsChanged(county, ucounty Facts, maleDelta, femaleDelta int64) {
	s.Equal(county.Pop+(maleDelta+femaleDelta), ucounty.Pop, fmt.Sprintf("County population should change by %d", maleDelta+femaleDelta))
	s.Equal(county.PopMale+maleDelta, ucounty.PopMale, fmt.Sprintf("County male population should change by %d", maleDelta))
	s.Equal(county.PopFemale+femaleDelta, ucounty.PopFemale, fmt.Sprintf("County female population should change by %d", femaleDelta))
}

func (s *ConditionInterceptorTestSuite) assertSubdivisionFactsChanged(cousub, ucousub Facts, maleDelta, femaleDelta int64) {
	s.Equal(cousub.Pop+(maleDelta+femaleDelta), ucousub.Pop, fmt.Sprintf("Subdivision population should change by %d", maleDelta+femaleDelta))
	s.Equal(cousub.PopMale+maleDelta, ucousub.PopMale, fmt.Sprintf("Subdivision male population should change by %d", maleDelta))
	s.Equal(cousub.PopFemale+femaleDelta, ucousub.PopFemale, fmt.Sprintf("Subivision female population should change by %d", femaleDelta))
}

func createConditionWithReferencedPatient(snomedCode string, patientID string) *models.Condition {

	return &models.Condition{
		Code: &models.CodeableConcept{
			Coding: []models.Coding{
				models.Coding{
					Code:   snomedCode,
					System: SnomedCodeSystem,
				},
			},
		},
		Subject: &models.Reference{
			Reference:    fmt.Sprintf("Patient/%s", patientID),
			ReferencedID: patientID,
		},
	}
}

func createAbatedConditionWithReferencedPatient(snomedCode string, patientID string) *models.Condition {

	var age = 20.0
	var isAbated = true

	condition := createConditionWithReferencedPatient(snomedCode, patientID)
	condition.AbatementAge = &models.Quantity{Value: &age, Unit: "years"}
	condition.AbatementBoolean = &isAbated
	condition.AbatementDateTime = &models.FHIRDateTime{Time: time.Now(), Precision: "ms"}
	condition.AbatementPeriod = &models.Period{
		Start: &models.FHIRDateTime{Time: time.Now(), Precision: "ms"},
		End:   &models.FHIRDateTime{Time: time.Now(), Precision: "ms"},
	}
	condition.AbatementString = "abated"
	return condition
}
