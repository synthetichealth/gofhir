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

var ErrInvalidGender = errors.New("Invalid gender")
var ErrInvalidUpdateOperation = errors.New("Invalid update operation")
var ErrNoAddress = errors.New("No address found")
var ErrNoSnomedCode = errors.New("No snomed code found")

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

	return da.updateStats(patient, "increment")
}

// RemovePatientStat decrements the relevant county and subdivision statistics
// based on the given patient's address.
func (da *PgStatsDataAccess) RemovePatientStat(patient *models.Patient) (err error) {

	return da.updateStats(patient, "decrement")
}

// AddConditionStat increments the relevant county and subdivision condition facts
// based on the given patient's address.
func (da *PgStatsDataAccess) AddConditionStat(patient *models.Patient, condition *models.Condition) (err error) {

	return da.updateFacts(patient, condition, "increment")
}

// RemoveConditionStat decrements the relevant county and subdivision condition facts
// based on the given patient's address.
func (da *PgStatsDataAccess) RemoveConditionStat(patient *models.Patient, condition *models.Condition) (err error) {

	return da.updateFacts(patient, condition, "decrement")
}

// identifyCountyAndSubdivisionForPatient returns the countyfp and cousubfp that
// match the subdivision in the given patient's address.
func (da *PgStatsDataAccess) identifyCountyAndSubdivisionForPatient(patient *models.Patient) (countyfp, cousubfp string, err error) {

	if !patientAddressIsValid(patient) {
		return "", "", ErrNoAddress
	}

	err = da.DB.QueryRow("SELECT countyfp, cousubfp FROM tiger.cousub WHERE name = $1", patient.Address[0].City).Scan(&countyfp, &cousubfp)
	return
}

// identifyDiseaseForCondition returns the diseasefp that matches the given
// condition's SNOMED code, if any.
func (da *PgStatsDataAccess) identifyDiseaseForCondition(condition *models.Condition) (diseasefp string, err error) {
	snomedCode, err := getSnomedCode(condition)

	if err != nil || snomedCode == "" {
		return "", ErrNoSnomedCode
	}
	err = da.DB.QueryRow("SELECT diseasefp FROM synth_ma.synth_disease WHERE code_snomed = $1", snomedCode).Scan(&diseasefp)
	return
}

// updateStats increments or decrements a row of population counts in the county and subdivision stats tables.
func (da *PgStatsDataAccess) updateStats(patient *models.Patient, op string) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)
	if err != nil {
		return
	}

	if !patientGenderIsValid(patient) {
		err = ErrInvalidGender
		return
	}

	var symbol string
	switch op {
	case "increment":
		symbol = "+"
	case "decrement":
		symbol = "-"
	default:
		return ErrInvalidUpdateOperation
	}

	var ctfp, csfp int
	countyQuery := fmt.Sprintf("UPDATE synth_ma.synth_county_stats SET pop = pop %s 1, pop_%s = pop_%s %s 1, pop_sm = ((pop %s 1) / sq_mi) WHERE ct_fips = $1 RETURNING ct_fips", symbol, patient.Gender, patient.Gender, symbol, symbol)
	cousubQuery := fmt.Sprintf("UPDATE synth_ma.synth_cousub_stats SET pop = pop %s 1, pop_%s = pop_%s %s 1, pop_sm = ((pop %s 1) / sq_mi) WHERE cs_fips = $1 RETURNING cs_fips", symbol, patient.Gender, patient.Gender, symbol, symbol)

	err = da.DB.QueryRow(countyQuery, countyfp).Scan(&ctfp)
	if err != nil {
		return
	}

	err = da.DB.QueryRow(cousubQuery, cousubfp).Scan(&csfp)
	return
}

// updateFacts increments or decrements a row of population counts in the county and subdivision fact tables.
func (da *PgStatsDataAccess) updateFacts(patient *models.Patient, condition *models.Condition, op string) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)
	if err != nil {
		return
	}

	diseasefp, err := da.identifyDiseaseForCondition(condition)
	if err != nil {
		return
	}

	if !patientGenderIsValid(patient) {
		err = ErrInvalidGender
		return
	}

	var symbol string
	switch op {
	case "increment":
		symbol = "+"
	case "decrement":
		symbol = "-"
	default:
		return ErrInvalidUpdateOperation
	}

	var ctfp, csfp int
	countyQuery := fmt.Sprintf("UPDATE synth_ma.synth_county_facts SET pop = pop %s 1, pop_%s = pop_%s %s 1 WHERE countyfp = $1 AND diseasefp = $2 RETURNING countyfp", symbol, patient.Gender, patient.Gender, symbol)
	cousubQuery := fmt.Sprintf("UPDATE synth_ma.synth_cousub_facts SET pop = pop %s 1, pop_%s = pop_%s %s 1 WHERE cousubfp = $1 AND diseasefp = $2 RETURNING cousubfp", symbol, patient.Gender, patient.Gender, symbol)

	err = da.DB.QueryRow(countyQuery, countyfp, diseasefp).Scan(&ctfp)
	if err != nil {
		return
	}

	err = da.DB.QueryRow(cousubQuery, cousubfp, diseasefp).Scan(&csfp)
	return
}

// patientGenderIsValid tests if the patient object provided has a valid gender.
func patientGenderIsValid(patient *models.Patient) bool {
	return (patient.Gender == "male" || patient.Gender == "female")
}

// patientAddressIsValid tests if the patient's address is valid
func patientAddressIsValid(patient *models.Patient) bool {
	if len(patient.Address) == 0 {
		return false
	}

	if patient.Address[0].City == "" {
		return false
	}
	return true
}

// getSnomedCode returns the condition's SNOMED code, if it exists
func getSnomedCode(condition *models.Condition) (code string, err error) {
	codings := condition.Code.Coding
	for _, coding := range codings {
		if coding.System == SnomedCodeSystem {
			return coding.Code, nil
		}
	}
	return "", ErrNoSnomedCode
}
