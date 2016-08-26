package stats

import (
	"log"
	"testing"

	"github.com/intervention-engine/fhir/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type StatTestSuite struct {
	DB  *gorm.DB
	DAL *StatsDataAccess
}

var _ = Suite(&StatTestSuite{})

func (s *StatTestSuite) SetUpSuite(c *C) {

	// Setup and connect to the test Postgres database. We test using the fhir_test
	// database. A script to setup this database is available in the postgres/ folder.

	// configure the GORM Postgres driver and database connection
	db, err := gorm.Open("postgres", "postgres://fhir_test:fhir_test@localhost/fhir_test?sslmode=disable")
	db.SingularTable(true) // disable table name pluralization globally

	if err != nil {
		log.Println("Make sure you run pgsetup.sh to create the fhir_test database before testing.")
		log.Fatal(err)
	}

	// ping the db to ensure we connected successfully
	if err := db.DB().Ping(); err != nil {
		log.Fatal(err)
	}
	s.DB = db
	s.DAL = NewPgStatsDataAccess(db)
}

func (s *StatTestSuite) TearDownSuite(c *C) {

	s.DB.Close()
}

func (s *StatTestSuite) TestCountyDataAccess(c *C) {

	// Get a county we know is in the database, by name
	county := s.DAL.Counties.GetCountyByName("Suffolk")

	// Make sure the county is as we expect
	c.Assert(county.Name, Equals, "Suffolk")
	c.Assert(county.CountyFp, Equals, "025")
	c.Assert(county.CountyIdFp, Equals, "25025")
	c.Assert(county.StateFp, Equals, "25")

	// Get another county by CountyFp
	county = s.DAL.Counties.GetCountyById("017") // Middlesex county

	// Make sure the county is as we expect
	c.Assert(county.Name, Equals, "Middlesex")
	c.Assert(county.CountyFp, Equals, "017")
	c.Assert(county.CountyIdFp, Equals, "25017")
	c.Assert(county.StateFp, Equals, "25")
}

func (s *StatTestSuite) TestSubdivisionDataAccess(c *C) {

	// Get a subdivision we know is in the database, by name
	cousub := s.DAL.Subdivisions.GetSubdivisionByName("Boston")

	// Make sure the subdivision is as we expect
	c.Assert(cousub.Name, Equals, "Boston")
	c.Assert(cousub.CountyFp, Equals, "025")
	c.Assert(cousub.CousubFp, Equals, "07000")
	c.Assert(cousub.CosbidFp, Equals, "2502507000")
	c.Assert(cousub.StateFp, Equals, "25")

	// Get another subdivision by CousubFp
	cousub = s.DAL.Subdivisions.GetSubdivisionById("17405") // Dover, MA

	// Make sure the subdivision is as we expect
	c.Assert(cousub.Name, Equals, "Dover")
	c.Assert(cousub.CountyFp, Equals, "021")
	c.Assert(cousub.CousubFp, Equals, "17405")
	c.Assert(cousub.CosbidFp, Equals, "2502117405")
	c.Assert(cousub.StateFp, Equals, "25")
}

func (s *StatTestSuite) TestSyntheticDiseaseDataAccess(c *C) {

	// Get a disease we know is in the database
	disease := s.DAL.SyntheticDiseases.GetSyntheticDiseaseByCondition("diabetes")

	// Make sure the disease is as we expect
	c.Assert(disease.StatName, Equals, "diabetes")
	c.Assert(disease.ConditionName, Equals, "diabetes")
	c.Assert(disease.ICD9Code, Equals, "")
	c.Assert(disease.ICD10Code, Equals, "")
	c.Assert(disease.SnomedCode, Equals, "44054006")
}

func (s *StatTestSuite) TestSyntheticCountyStatDataAccess(c *C) {

	var stat, ustat SyntheticCountyStat

	// Get a county for testing
	county := s.DAL.Counties.GetCountyByName("Suffolk")

	// Get a stat we know is in the database
	stat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)

	// Make sure the stat is the one we expect
	c.Assert(stat.CountyName, Equals, "Suffolk")
	c.Assert(stat.CountyFp, Equals, "025")

	// Add a male to the stat
	s.DAL.SyntheticCountyStats.AddMaleToStat(stat)
	ustat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	assertCountyStatChanged(c, stat, ustat, 1, 0)

	// Add a female to the stat
	stat = ustat
	s.DAL.SyntheticCountyStats.AddFemaleToStat(stat)
	ustat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	assertCountyStatChanged(c, stat, ustat, 0, 1)

	// Remove a male from the stat
	stat = ustat
	s.DAL.SyntheticCountyStats.RemoveMaleFromStat(stat)
	ustat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	assertCountyStatChanged(c, stat, ustat, -1, 0)

	// Remove a female from the stat
	stat = ustat
	s.DAL.SyntheticCountyStats.RemoveFemaleFromStat(stat)
	ustat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	assertCountyStatChanged(c, stat, ustat, 0, -1)
}

func (s *StatTestSuite) TestSyntheticSubdivisionStatDataAccess(c *C) {

	var stat, ustat SyntheticSubdivisionStat

	// Get a subdivision for testing
	cousub := s.DAL.Subdivisions.GetSubdivisionByName("Bedford")

	// Get a stat we know is in the database
	stat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)

	// Make sure the stat is the one we expect
	c.Assert(stat.SubdivisionName, Equals, "Bedford")
	c.Assert(stat.CountyFp, Equals, "017") // Middlesex County
	c.Assert(stat.CousubFp, Equals, "04615")

	// Add a male to the stat
	s.DAL.SyntheticSubdivisionStats.AddMaleToStat(stat)
	ustat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertSubdivisionStatChanged(c, stat, ustat, 1, 0)

	// Add a female to the stat
	stat = ustat
	s.DAL.SyntheticSubdivisionStats.AddFemaleToStat(stat)
	ustat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertSubdivisionStatChanged(c, stat, ustat, 0, 1)

	// Remove a male from the stat
	stat = ustat
	s.DAL.SyntheticSubdivisionStats.RemoveMaleFromStat(stat)
	ustat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertSubdivisionStatChanged(c, stat, ustat, -1, 0)

	// Remove a female from the stat
	stat = ustat
	s.DAL.SyntheticSubdivisionStats.RemoveFemaleFromStat(stat)
	ustat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertSubdivisionStatChanged(c, stat, ustat, 0, -1)
}

func (s *StatTestSuite) TestSyntheticCountyFactDataAccess(c *C) {

	var fact, ufact SyntheticCountyFact

	// Get a county for testing
	county := s.DAL.Counties.GetCountyByName("Middlesex")

	// Get a fact we know is in the database
	fact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, "diabetes")

	// Make sure the fact is the one we expect
	c.Assert(fact.CountyIdFp, Equals, county.CountyIdFp)
	c.Assert(fact.DiseaseId, Equals, "1") // 1 = diabetes

	// Add a male to the fact
	s.DAL.SyntheticCountyFacts.AddMaleToFact(fact)
	ufact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, "diabetes")
	assertCountyFactChanged(c, fact, ufact, 1, 0)

	// Add a female to the fact
	fact = ufact
	s.DAL.SyntheticCountyFacts.AddFemaleToFact(fact)
	ufact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, "diabetes")
	assertCountyFactChanged(c, fact, ufact, 0, 1)

	// Remove a male from the fact
	fact = ufact
	s.DAL.SyntheticCountyFacts.RemoveMaleFromFact(fact)
	ufact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, "diabetes")
	assertCountyFactChanged(c, fact, ufact, -1, 0)

	// Remove a female from the fact
	fact = ufact
	s.DAL.SyntheticCountyFacts.RemoveFemaleFromFact(fact)
	ufact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, "diabetes")
	assertCountyFactChanged(c, fact, ufact, 0, -1)
}

func (s *StatTestSuite) TestSyntheticSubdivisionFactDataAccess(c *C) {

	var fact, ufact SyntheticSubdivisionFact

	// Get a subdivision for testing
	cousub := s.DAL.Subdivisions.GetSubdivisionByName("Haverhill")

	// Get a fact we know is in the database
	fact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, "diabetes")

	// Make sure the fact is the one we expect
	c.Assert(fact.CosbidFp, Equals, cousub.CosbidFp)
	c.Assert(fact.DiseaseId, Equals, "1") // 1 = diabetes

	// Add a male to the fact
	s.DAL.SyntheticSubdivisionFacts.AddMaleToFact(fact)
	ufact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, "diabetes")
	assertSubdivisionFactChanged(c, fact, ufact, 1, 0)

	// Add a female to the fact
	fact = ufact
	s.DAL.SyntheticSubdivisionFacts.AddFemaleToFact(fact)
	ufact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, "diabetes")
	assertSubdivisionFactChanged(c, fact, ufact, 0, 1)

	// Remove a male from the fact
	fact = ufact
	s.DAL.SyntheticSubdivisionFacts.RemoveMaleFromFact(fact)
	ufact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, "diabetes")
	assertSubdivisionFactChanged(c, fact, ufact, -1, 0)

	// Remove a female from the fact
	fact = ufact
	s.DAL.SyntheticSubdivisionFacts.RemoveFemaleFromFact(fact)
	ufact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, "diabetes")
	assertSubdivisionFactChanged(c, fact, ufact, 0, -1)
}

func (s *StatTestSuite) TestTopLevelPatientDataAccess(c *C) {

	var countyStat, ucountyStat SyntheticCountyStat
	var cousubStat, ucousubStat SyntheticSubdivisionStat

	// Create patient for testing
	patient := &models.Patient{
		Gender: "male",
		Address: []models.Address{
			models.Address{
				City:       "Dover",
				State:      "MA",
				PostalCode: "02030",
			},
		},
	}

	county, cousub := s.DAL.GetCountyAndSubdivisionForPatient(patient)
	cousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	countyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)

	// Add a male patient
	s.DAL.AddMalePatient(patient)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)

	// Test that both the county and subdivision stats for the patient were updated
	assertCountyStatChanged(c, countyStat, ucountyStat, 1, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 1, 0)

	// Add a female patient
	patient.Gender = "female"
	countyStat = ucountyStat
	cousubStat = ucousubStat
	s.DAL.AddFemalePatient(patient)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, 1)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, 1)

	// Remove a female patient
	countyStat = ucountyStat
	cousubStat = ucousubStat
	s.DAL.RemoveFemalePatient(patient)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, -1)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, -1)

	// Remove a male patient
	patient.Gender = "male"
	countyStat = ucountyStat
	cousubStat = ucousubStat
	s.DAL.RemoveMalePatient(patient)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	assertCountyStatChanged(c, countyStat, ucountyStat, -1, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, -1, 0)
}

func (s *StatTestSuite) TestTopLevelConditionDataAcess(c *C) {

	var countyFact, ucountyFact SyntheticCountyFact
	var cousubFact, ucousubFact SyntheticSubdivisionFact

	patient := &models.Patient{
		Gender: "male",
		Address: []models.Address{
			models.Address{
				City:       "Marblehead",
				State:      "MA",
				PostalCode: "01945",
			},
		},
	}

	condition := &models.Condition{
		Code: &models.CodeableConcept{
			Coding: []models.Coding{
				models.Coding{
					Code:   "44054006",
					System: SnomedCodeSystem,
				},
			},
		},
	}

	county, cousub := s.DAL.GetCountyAndSubdivisionForPatient(patient)
	conditionName := s.DAL.GetConditionName(condition)
	countyFact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, conditionName)
	cousubFact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, conditionName)

	// Add a male condition
	s.DAL.AddMaleCondition(patient, condition)
	ucountyFact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, conditionName)
	ucousubFact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, conditionName)
	assertCountyFactChanged(c, countyFact, ucountyFact, 1, 0)
	assertSubdivisionFactChanged(c, cousubFact, ucousubFact, 1, 0)

	// Add a female condition
	countyFact = ucountyFact
	cousubFact = ucousubFact
	patient.Gender = "female"
	s.DAL.AddFemaleCondition(patient, condition)
	ucountyFact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, conditionName)
	ucousubFact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, conditionName)
	assertCountyFactChanged(c, countyFact, ucountyFact, 0, 1)
	assertSubdivisionFactChanged(c, cousubFact, ucousubFact, 0, 1)

	// Remove a female condition
	countyFact = ucountyFact
	cousubFact = ucousubFact
	s.DAL.RemoveFemaleCondition(patient, condition)
	ucountyFact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, conditionName)
	ucousubFact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, conditionName)
	assertCountyFactChanged(c, countyFact, ucountyFact, 0, -1)
	assertSubdivisionFactChanged(c, cousubFact, ucousubFact, 0, -1)

	// Remove a male condition
	countyFact = ucountyFact
	cousubFact = ucousubFact
	patient.Gender = "male"
	s.DAL.RemoveMaleCondition(patient, condition)
	ucountyFact = s.DAL.SyntheticCountyFacts.GetFactByCountyAndCondition(county, conditionName)
	ucousubFact = s.DAL.SyntheticSubdivisionFacts.GetFactBySubdivisionAndCondition(cousub, conditionName)
	assertCountyFactChanged(c, countyFact, ucountyFact, -1, 0)
	assertSubdivisionFactChanged(c, cousubFact, ucousubFact, -1, 0)
}

func assertCountyStatChanged(c *C, stat, ustat SyntheticCountyStat, maleDelta, femaleDelta int64) {
	c.Assert(ustat.Population, Equals, stat.Population+(maleDelta+femaleDelta))
	c.Assert(ustat.PopulationMale, Equals, stat.PopulationMale+maleDelta)
	c.Assert(ustat.PopulationFemale, Equals, stat.PopulationFemale+femaleDelta)
	newPopPerSqMile := float64(stat.Population+(maleDelta+femaleDelta)) / stat.SquareMiles
	c.Assert(ustat.PopulationPerSquareMile, Equals, newPopPerSqMile)
}

func assertSubdivisionStatChanged(c *C, stat, ustat SyntheticSubdivisionStat, maleDelta, femaleDelta int64) {
	c.Assert(ustat.Population, Equals, stat.Population+(maleDelta+femaleDelta))
	c.Assert(ustat.PopulationMale, Equals, stat.PopulationMale+maleDelta)
	c.Assert(ustat.PopulationFemale, Equals, stat.PopulationFemale+femaleDelta)
	newPopPerSqMile := float64(stat.Population+(maleDelta+femaleDelta)) / stat.SquareMiles
	c.Assert(ustat.PopulationPerSquareMile, Equals, newPopPerSqMile)
}

func assertCountyFactChanged(c *C, fact, ufact SyntheticCountyFact, maleDelta, femaleDelta int64) {
	c.Assert(ufact.Population, Equals, fact.Population+(maleDelta+femaleDelta))
	c.Assert(ufact.PopulationMale, Equals, fact.PopulationMale+maleDelta)
	c.Assert(ufact.PopulationFemale, Equals, fact.PopulationFemale+femaleDelta)
}

func assertSubdivisionFactChanged(c *C, fact, ufact SyntheticSubdivisionFact, maleDelta, femaleDelta int64) {
	c.Assert(ufact.Population, Equals, fact.Population+(maleDelta+femaleDelta))
	c.Assert(ufact.PopulationMale, Equals, fact.PopulationMale+maleDelta)
	c.Assert(ufact.PopulationFemale, Equals, fact.PopulationFemale+femaleDelta)
}
