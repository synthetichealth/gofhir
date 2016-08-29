package stats

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/intervention-engine/fhir/models"
	_ "github.com/lib/pq"
)

// Condition coding systems
const SnomedCodeSystem = "http://snomed.info/sct"
const ICD9CodeSystem = "http://www.icd9data.com/"
const ICD10CodeSystem = "http://www.icd10data.com/"

var InvalidGenderError = errors.New("Invalid gender")
var InvalidUpdateOperationError = errors.New("Invalid update operation")

// StatsDataAccess is the top level interface for interacting with Synthetic Mass statistics
type StatsDataAccess interface {
	AddPatientStat(patient *models.Patient) error
	RemovePatientStat(patient *models.Patient) error
	AddConditionStat(patient *models.Patient, condition *models.Condition) error
	RemoveConditionStat(patient *models.Patient, condition *models.Condition) error
}

// PgStatsDataAccess implements the StatsDataAccess interface using a Postgres database.
type PgStatsDataAccess struct {
	DB *sql.DB
}

// NewPgStatsDataAccess returns a new instance of PgStatsDataAccess.
func NewPgStatsDataAccess(db *sql.DB) *PgStatsDataAccess {
	return &PgStatsDataAccess{DB: db}
}

// AddPatientStat increments the relevant county and subdivision statistics
// based on the given patient's address.
func (da *PgStatsDataAccess) AddPatientStat(patient *models.Patient) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		err = da.updateStats(countyfp, cousubfp, patient.Gender, "increment")

	} else {
		err = InvalidGenderError
	}
	return
}

// RemovePatientStat decrements the relevant county and subdivision statistics
// based on the given patient's address.
func (da *PgStatsDataAccess) RemovePatientStat(patient *models.Patient) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		err = da.updateStats(countyfp, cousubfp, patient.Gender, "decrement")

	} else {
		err = InvalidGenderError
	}
	return
}

// AddConditionStat increments the relevant county and subdivision condition facts
// based on the given patient's address.
func (da *PgStatsDataAccess) AddConditionStat(patient *models.Patient, condition *models.Condition) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)
	diseasefp, err := da.identifyDiseaseForCondition(condition)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		err = da.updateFacts(countyfp, cousubfp, diseasefp, patient.Gender, "increment")

	} else {
		err = InvalidGenderError
	}
	return
}

// RemoveConditionStat decrements the relevant county and subdivision condition facts
// based on the given patient's address.
func (da *PgStatsDataAccess) RemoveConditionStat(patient *models.Patient, condition *models.Condition) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)
	diseasefp, err := da.identifyDiseaseForCondition(condition)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		err = da.updateFacts(countyfp, cousubfp, diseasefp, patient.Gender, "decrement")

	} else {
		err = InvalidGenderError
	}
	return
}

// identifyCountyAndSubdivisionForPatient returns the countyfp and cousubfp that
// match the subdivision in the given patient's address.
func (da *PgStatsDataAccess) identifyCountyAndSubdivisionForPatient(patient *models.Patient) (countyfp, cousubfp string, err error) {
	err = da.DB.QueryRow("SELECT countyfp, cousubfp FROM tiger.cousub WHERE name = $1", patient.Address[0].City).Scan(&countyfp, &cousubfp)
	return
}

// identifyDiseaseForCondition returns the diseasefp that matches the given
// condition's SNOMED code, if any.
func (da *PgStatsDataAccess) identifyDiseaseForCondition(condition *models.Condition) (diseasefp string, err error) {
	err = da.DB.QueryRow("SELECT diseasefp FROM synth_ma.synth_disease WHERE code_snomed = $1", getSnomedCode(condition)).Scan(&diseasefp)
	return
}

// updateStats increments or decrements a row of population counts in the county and subdivision stats tables.
func (da *PgStatsDataAccess) updateStats(countyfp, cousubfp, gender, op string) (err error) {

	var symbol string
	switch op {
	case "increment":
		symbol = "+"
	case "decrement":
		symbol = "-"
	default:
		return InvalidUpdateOperationError
	}

	var ctfp, csfp int
	countyQuery := fmt.Sprintf("UPDATE synth_ma.synth_county_stats SET pop = pop %s 1, pop_%s = pop_%s %s 1, pop_sm = ((pop %s 1) / sq_mi) WHERE ct_fips = $1 RETURNING ct_fips", symbol, gender, gender, symbol, symbol)
	cousubQuery := fmt.Sprintf("UPDATE synth_ma.synth_cousub_stats SET pop = pop %s 1, pop_%s = pop_%s %s 1, pop_sm = ((pop %s 1) / sq_mi) WHERE cs_fips = $1 RETURNING cs_fips", symbol, gender, gender, symbol, symbol)
	err = da.DB.QueryRow(countyQuery, countyfp).Scan(&ctfp)
	err = da.DB.QueryRow(cousubQuery, cousubfp).Scan(&csfp)
	return
}

// updateFacts increments or decrements a row of population counts in the county and subdivision fact tables.
func (da *PgStatsDataAccess) updateFacts(countyfp, cousubfp, diseasefp, gender, op string) (err error) {

	var symbol string
	switch op {
	case "increment":
		symbol = "+"
	case "decrement":
		symbol = "-"
	default:
		return InvalidUpdateOperationError
	}

	var ctfp, csfp int
	countyQuery := fmt.Sprintf("UPDATE synth_ma.synth_county_facts SET pop = pop %s 1, pop_%s = pop_%s %s 1 WHERE countyfp = $1 AND diseasefp = $2 RETURNING countyfp", symbol, gender, gender, symbol)
	cousubQuery := fmt.Sprintf("UPDATE synth_ma.synth_cousub_facts SET pop = pop %s 1, pop_%s = pop_%s %s 1 WHERE cousubfp = $1 AND diseasefp = $2 RETURNING cousubfp", symbol, gender, gender, symbol)
	err = da.DB.QueryRow(countyQuery, countyfp, diseasefp).Scan(&ctfp)
	err = da.DB.QueryRow(cousubQuery, cousubfp, diseasefp).Scan(&csfp)
	return
}

// patientGenderIsValid tests if the patient object provided has a valid gender.
func patientGenderIsValid(patient *models.Patient) bool {
	return (patient.Gender == "male" || patient.Gender == "female")
}

// getSnomedCode returns the condition's SNOMED code, if it exists
func getSnomedCode(condition *models.Condition) string {
	codings := condition.Code.Coding
	for _, coding := range codings {
		if coding.System == SnomedCodeSystem {
			return coding.Code
		}
	}
	return ""
}
