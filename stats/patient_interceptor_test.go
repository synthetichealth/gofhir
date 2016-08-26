package stats

import (
	"log"

	"github.com/intervention-engine/fhir/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "gopkg.in/check.v1"
)

type InterceptorTestSuite struct {
	DB  *gorm.DB
	DAL *StatsDataAccess
}

var _ = Suite(&InterceptorTestSuite{})

func (s *InterceptorTestSuite) SetUpSuite(c *C) {

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

func (s *InterceptorTestSuite) TearDownSuite(c *C) {

	s.DB.Close()
}

// TestPatientStatsCreateInterceptor tests the PatientStatsCreateInterceptor's ability to update
// patient statistics after a new patient is added to the database
func (s *InterceptorTestSuite) TestPatientStatsCreateInterceptor(c *C) {

	var patient *models.Patient
	var county County
	var cousub Subdivision
	var countyStat, ucountyStat SyntheticCountyStat
	var cousubStat, ucousubStat SyntheticSubdivisionStat

	createInterceptor := &PatientStatsCreateInterceptor{
		DataAccess: s.DAL,
	}

	patient = &models.Patient{
		Gender: "male",
		Address: []models.Address{
			models.Address{
				City:       "Boston",
				State:      "MA",
				PostalCode: "02215",
			},
		},
	}

	// Get existing stats
	cousub = s.DAL.Subdivisions.GetSubdivisionByName(patient.Address[0].City)
	county = s.DAL.Counties.GetCountyById(cousub.CountyFp)
	countyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)

	// 1. Test a CreateMale operation
	// ------------------------------------------------------------------------
	createInterceptor.After(patient)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 1, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 1, 0)

	// 2. Test a CreateFemale operation
	// ------------------------------------------------------------------------
	patient.Gender = "female"
	countyStat = ucountyStat
	cousubStat = ucousubStat

	createInterceptor.After(patient)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, 1)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, 1)

}

// TestPatientStatsUpdateInterceptor tests the PatientStatsUpdateInterceptor's ability to update
// patient statistics after a new patient is added to the database
func (s *InterceptorTestSuite) TestPatientStatsUpdateInterceptor(c *C) {

	var patient, updatedPatient *models.Patient
	var county, county2 County
	var cousub, cousub2 Subdivision
	var countyStat, ucountyStat, countyStat2, ucountyStat2 SyntheticCountyStat
	var cousubStat, ucousubStat, cousubStat2, ucousubStat2 SyntheticSubdivisionStat

	updateInterceptor := &PatientStatsUpdateInterceptor{
		DataAccess: s.DAL,
	}

	patient = &models.Patient{
		Gender: "male",
		Address: []models.Address{
			models.Address{
				City:       "Boston",
				State:      "MA",
				PostalCode: "02215",
			},
		},
	}

	updatedPatient = &models.Patient{
		Gender: "male",
		Address: []models.Address{
			models.Address{
				City:       "Bedford",
				State:      "MA",
				PostalCode: "01730",
			},
		},
	}

	// Get existing stats
	cousub = s.DAL.Subdivisions.GetSubdivisionByName(patient.Address[0].City)
	county = s.DAL.Counties.GetCountyById(cousub.CountyFp)
	countyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)

	// 3. Test a UpdateMale operation with no relevant changes
	// ------------------------------------------------------------------------
	updateInterceptor.Before(patient)
	updateInterceptor.After(patient)

	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, 0)

	// 4. Test a UpdateFemale operation with no relevant changes
	// ------------------------------------------------------------------------
	patient.Gender = "female"
	countyStat = ucountyStat
	cousubStat = ucousubStat

	updateInterceptor.Before(patient)
	updateInterceptor.After(patient)

	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, 0)

	// 5. Test a UpdateFemale operation with relevant statistic changes
	// ------------------------------------------------------------------------
	updatedPatient.Gender = "female"
	countyStat = ucountyStat
	cousubStat = ucousubStat

	// The patient's City will change from Boston to Bedford, so we need to get Bedford's stats too
	cousub2 = s.DAL.Subdivisions.GetSubdivisionByName(updatedPatient.Address[0].City)
	county2 = s.DAL.Counties.GetCountyById(cousub2.CountyFp)
	countyStat2 = s.DAL.SyntheticCountyStats.GetStatByCounty(county2)
	cousubStat2 = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub2)

	updateInterceptor.Before(patient)
	updateInterceptor.After(updatedPatient)

	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	ucountyStat2 = s.DAL.SyntheticCountyStats.GetStatByCounty(county2)
	ucousubStat2 = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub2)

	// Boston stats should have decremented
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, -1)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, -1)

	// Bedford stats should have incremented
	assertCountyStatChanged(c, countyStat2, ucountyStat2, 0, 1)
	assertSubdivisionStatChanged(c, cousubStat2, ucousubStat2, 0, 1)

	// 6. Test a UpdateMale operation with relevant statistic changes
	// ------------------------------------------------------------------------
	patient.Gender = "male"
	updatedPatient.Gender = "male"
	countyStat = ucountyStat
	cousubStat = ucousubStat
	countyStat2 = ucountyStat2
	cousubStat2 = ucousubStat2

	updateInterceptor.Before(patient)
	updateInterceptor.After(updatedPatient)

	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	ucountyStat2 = s.DAL.SyntheticCountyStats.GetStatByCounty(county2)
	ucousubStat2 = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub2)

	// Boston stats should have decremented
	assertCountyStatChanged(c, countyStat, ucountyStat, -1, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, -1, 0)

	// Bedford stats should have incremented
	assertCountyStatChanged(c, countyStat2, ucountyStat2, 1, 0)
	assertSubdivisionStatChanged(c, cousubStat2, ucousubStat2, 1, 0)
}

// TestPatientStatsDeleteInterceptor tests the PatientStatsDeleteInterceptor's ability to update
// patient statistics after a new patient is added to the database
func (s *InterceptorTestSuite) TestPatientStatsDeleteInterceptor(c *C) {

	var patient *models.Patient
	var county County
	var cousub Subdivision
	var countyStat, ucountyStat SyntheticCountyStat
	var cousubStat, ucousubStat SyntheticSubdivisionStat

	deleteInterceptor := &PatientStatsDeleteInterceptor{
		DataAccess: s.DAL,
	}

	patient = &models.Patient{
		Gender: "male",
		Address: []models.Address{
			models.Address{
				City:       "Boston",
				State:      "MA",
				PostalCode: "02215",
			},
		},
	}

	// Get existing stats
	cousub = s.DAL.Subdivisions.GetSubdivisionByName(patient.Address[0].City)
	county = s.DAL.Counties.GetCountyById(cousub.CountyFp)
	countyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)

	// 7. Test a DeleteMale operation
	// ------------------------------------------------------------------------
	deleteInterceptor.After(patient)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, -1, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, -1, 0)

	// 8. Test a DeleteFemale operation
	// ------------------------------------------------------------------------
	patient.Gender = "female"
	countyStat = ucountyStat
	cousubStat = ucousubStat

	deleteInterceptor.After(patient)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, -1)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, -1)
}

// TestPatientStatsInterceptorErrorHandling tests that the interceptors and underlying
// data access layers handle errors as expected
func (s *InterceptorTestSuite) TestPatientStatsInterceptorErrorHandling(c *C) {

	var countyStat, ucountyStat SyntheticCountyStat
	var cousubStat, ucousubStat SyntheticSubdivisionStat

	patient := &models.Patient{
		Gender: "foo",
		Address: []models.Address{
			models.Address{
				City:       "Boston",
				State:      "MA",
				PostalCode: "02215",
			},
		},
	}

	createInterceptor := &PatientStatsCreateInterceptor{
		DataAccess: s.DAL,
	}

	updateInterceptor := &PatientStatsUpdateInterceptor{
		DataAccess: s.DAL,
	}

	deleteInterceptor := &PatientStatsDeleteInterceptor{
		DataAccess: s.DAL,
	}

	cousub := s.DAL.Subdivisions.GetSubdivisionByName(patient.Address[0].City)
	county := s.DAL.Counties.GetCountyById(cousub.CountyFp)
	countyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)

	// 9. Test an invalid gender (fails silently and doesn't modify DB)
	// ------------------------------------------------------------------------
	createInterceptor.After(patient)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, 0)

	countyStat = ucountyStat
	cousubStat = ucousubStat
	updateInterceptor.Before(patient)
	updateInterceptor.After(patient)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, 0)

	countyStat = ucountyStat
	cousubStat = ucousubStat
	deleteInterceptor.After(patient)
	ucountyStat = s.DAL.SyntheticCountyStats.GetStatByCounty(county)
	ucousubStat = s.DAL.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	assertCountyStatChanged(c, countyStat, ucountyStat, 0, 0)
	assertSubdivisionStatChanged(c, cousubStat, ucousubStat, 0, 0)

	// 10. Test a city that does not exist (fails silently)
	// ------------------------------------------------------------------------
	patient.Gender = "male"
	patient.Address[0].City = "foo"

	createInterceptor.After(patient)
	updateInterceptor.After(patient)
	deleteInterceptor.After(patient)
}
