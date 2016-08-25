package stats

import (
	"github.com/intervention-engine/fhir/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// County is a county in the United States of America.
type County struct {
	CountyIdFp string `gorm:"column:cntyidfp;primary_key"`
	StateFp    string `gorm:"column:statefp"`
	CountyFp   string `gorm:"column:countyfp"`
	Name       string `gorm:"column:name"`
}

func (County) TableName() string {
	return "tiger.county"
}

// CountySubdivision is a county subdivision (city or town) in the United States of America.
type Subdivision struct {
	CosbidFp string `gorm:"column:cosbidfp;primary_key"`
	StateFp  string `gorm:"column:statefp"`
	CountyFp string `gorm:"column:countyfp"`
	CousubFp string `gorm:"column:cousubfp"`
	Name     string `gorm:"column:name"`
}

func (Subdivision) TableName() string {
	return "tiger.cousub"
}

// SyntheticDisease is a patient condition that we track statistics for (e.g. diabetes, heart disease).
// A single stat, e.g. 'heart_disease', may track several conditions (e.g. myocardial_infarction
// AND coronary_heart_disease).
type SyntheticDisease struct {
	Id            string `gorm:"column:diseasefp;primary_key"`
	StatName      string `gorm:"column:stat_name"`
	ConditionName string `gorm:"column:condition_name"`
	ICD9Code      string `gorm:"column:code_icd9"`
	ICD10Code     string `gorm:"column:code_icd10"`
	SnomedCode    string `gorm:"column:code_snomed"`
}

func (SyntheticDisease) TableName() string {
	return "synth_ma.synth_disease"
}

// SyntheticCountyStat is a set of statistics for a given county.
type SyntheticCountyStat struct {
	CountyName              string  `gorm:"column:ct_name"`
	CountyFp                string  `gorm:"column:ct_fips;primary_key"`
	SquareMiles             float64 `gorm:"column:sq_mi"`
	Population              int64   `gorm:"column:pop"`
	PopulationMale          int64   `gorm:"column:pop_male"`
	PopulationFemale        int64   `gorm:"column:pop_female"`
	PopulationPerSquareMile float64 `gorm:"column:pop_sm"`
}

func (SyntheticCountyStat) TableName() string {
	return "synth_ma.synth_county_stats"
}

// SyntheticSubdivisionStat is a set of statistics for a given subdivision.
type SyntheticSubdivisionStat struct {
	SubdivisionName         string  `gorm:"column:cs_name"`
	CountyFp                string  `gorm:"column:ct_fips;primary_key"`
	CousubFp                string  `gorm:"column:cs_fips;primary_key"`
	SquareMiles             float64 `gorm:"column:sq_mi"`
	Population              int64   `gorm:"column:pop"`
	PopulationMale          int64   `gorm:"column:pop_male"`
	PopulationFemale        int64   `gorm:"column:pop_female"`
	PopulationPerSquareMile float64 `gorm:"column:pop_sm"`
}

func (SyntheticSubdivisionStat) TableName() string {
	return "synth_ma.synth_cousub_stats"
}

// SyntheticCountyFact is a set of statistics on a particular disease in a particular county.
type SyntheticCountyFact struct {
	Id               string `gorm:"column:factid;primary_key"`
	CountyIdFp       string `gorm:"column:cntyidfp"`
	DiseaseId        string `gorm:"column:diseasefp"`
	Population       int64  `gorm:"column:pop"`
	PopulationMale   int64  `gorm:"column:pop_male"`
	PopulationFemale int64  `gorm:"column:pop_female"`
}

func (SyntheticCountyFact) TableName() string {
	return "synth_ma.synth_county_facts"
}

// CountySyntheticSubdivisionFact is a set of statistics on a particular disease in a particular county subdivision.
type SyntheticSubdivisionFact struct {
	Id               string `gorm:"column:factid;primary_key"`
	CosbidFp         string `gorm:"column:cosbidfp"`
	DiseaseId        string `gorm:"column:diseasefp"`
	Population       int64  `gorm:"column:pop"`
	PopulationMale   int64  `gorm:"column:pop_male"`
	PopulationFemale int64  `gorm:"column:pop_female"`
}

func (SyntheticSubdivisionFact) TableName() string {
	return "synth_ma.synth_cousub_facts"
}

// CountyDataAccess is an interface for interacting with the County data object.
type CountyDataAccess interface {
	GetCountyById(countyfp string) County
	GetCountyByName(name string) County
}

// SubdivisionDataAccess is an interface for interacting with the Subdivision data object.
type SubdivisionDataAccess interface {
	GetSubdivisionById(cousubfp string) Subdivision
	GetSubdivisionByName(name string) Subdivision
}

// SyntheticDiseaseDataAccess is an interface for interacting with the SyntheticDisease data object.
type SyntheticDiseaseDataAccess interface {
	GetSyntheticDiseaseByCondition(conditionName string) SyntheticDisease
}

// SyntheticCountyStatDataAccess is an interface for interacting with the SyntheticCountyStat data object.
type SyntheticCountyStatDataAccess interface {
	GetStatByCounty(county County) SyntheticCountyStat
	AddMaleToStat(stat SyntheticCountyStat)
	AddFemaleToStat(stat SyntheticCountyStat)
	RemoveMaleFromStat(stat SyntheticCountyStat)
	RemoveFemaleFromStat(stat SyntheticCountyStat)
}

// SyntheticSubdivisionStatDataAccess is an interface for interacting with the SyntheticSubdivisionStat data object.
type SyntheticSubdivisionStatDataAccess interface {
	GetStatBySubdivision(cousub Subdivision) SyntheticSubdivisionStat
	AddMaleToStat(stat SyntheticSubdivisionStat)
	AddFemaleToStat(stat SyntheticSubdivisionStat)
	RemoveMaleFromStat(stat SyntheticSubdivisionStat)
	RemoveFemaleFromStat(stat SyntheticSubdivisionStat)
}

// SyntheticCountyFactDataAccess is an interface for interacting with the SyntheticCountyFact data object.
type SyntheticCountyFactDataAccess interface {
	GetFactByCountyAndCondition(county County, conditionName string) SyntheticCountyFact
	AddMaleToFact(fact SyntheticCountyFact)
	AddFemaleToFact(fact SyntheticCountyFact)
	RemoveMaleFromFact(fact SyntheticCountyFact)
	RemoveFemaleFromFact(fact SyntheticCountyFact)
}

// SyntheticSubdivisionFactDataAccess is an interface for interacting with the SyntheticSubdivisionFact data object.
type SyntheticSubdivisionFactDataAccess interface {
	GetFactBySubdivisionAndCondition(cousub Subdivision, conditionName string) SyntheticSubdivisionFact
	AddMaleToFact(fact SyntheticSubdivisionFact)
	AddFemaleToFact(fact SyntheticSubdivisionFact)
	RemoveMaleFromFact(fact SyntheticSubdivisionFact)
	RemoveFemaleFromFact(fact SyntheticSubdivisionFact)
}

// PgCountyDataAccess implements the CountyDataAccess interface using a Postgres data connection.
type PgCountyDataAccess struct {
	DB *gorm.DB
}

func (da *PgCountyDataAccess) GetCountyById(countyfp string) County {
	var county County
	da.DB.Where(&County{CountyFp: countyfp}).First(&county)
	return county
}

func (da *PgCountyDataAccess) GetCountyByName(name string) County {
	var county County
	da.DB.Where(&County{Name: name}).First(&county)
	return county
}

// PgSubdivisionDataAccess implements the SubdivisionDataAccess interface using a Postgres data connection.
type PgSubdivisionDataAccess struct {
	DB *gorm.DB
}

func (da *PgSubdivisionDataAccess) GetSubdivisionById(cousubfp string) Subdivision {
	var cousub Subdivision
	da.DB.Where(&Subdivision{CousubFp: cousubfp}).First(&cousub)
	return cousub
}

func (da *PgSubdivisionDataAccess) GetSubdivisionByName(name string) Subdivision {
	var cousub Subdivision
	da.DB.Where(&Subdivision{Name: name}).First(&cousub)
	return cousub
}

// PgSyntheticDiseaseDataAccess implements the SyntheticDiseaseDataAccess interface using a Postgres data connection.
type PgSyntheticDiseaseDataAccess struct {
	DB *gorm.DB
}

// GetSyntheticDiseaseByCondition returns a disease that matches the given condition (e.g. 'diabetes').
func (da *PgSyntheticDiseaseDataAccess) GetSyntheticDiseaseByCondition(conditionName string) SyntheticDisease {
	var disease SyntheticDisease
	da.DB.Where(&SyntheticDisease{ConditionName: conditionName}).First(&disease)
	return disease
}

// PgSyntheticCountyStatDataAccess implements the SyntheticCountyStatDataAccess interface using a Postgres database connection.
type PgSyntheticCountyStatDataAccess struct {
	DB *gorm.DB
}

func (da *PgSyntheticCountyStatDataAccess) GetStatByCounty(county County) SyntheticCountyStat {
	var stat SyntheticCountyStat
	da.DB.Where(&SyntheticCountyStat{CountyFp: county.CountyFp}).First(&stat)
	return stat
}

func (da *PgSyntheticCountyStatDataAccess) AddMaleToStat(stat SyntheticCountyStat) {
	da.modifyPopulationCount(stat, 1, 0)
}

func (da *PgSyntheticCountyStatDataAccess) AddFemaleToStat(stat SyntheticCountyStat) {
	da.modifyPopulationCount(stat, 0, 1)
}

func (da *PgSyntheticCountyStatDataAccess) RemoveMaleFromStat(stat SyntheticCountyStat) {
	da.modifyPopulationCount(stat, -1, 0)
}

func (da *PgSyntheticCountyStatDataAccess) RemoveFemaleFromStat(stat SyntheticCountyStat) {
	da.modifyPopulationCount(stat, 0, -1)
}

func (da *PgSyntheticCountyStatDataAccess) modifyPopulationCount(stat SyntheticCountyStat, maleDelta, femaleDelta int64) {
	stat.Population += (maleDelta + femaleDelta)
	stat.PopulationMale += maleDelta
	stat.PopulationFemale += femaleDelta
	stat.PopulationPerSquareMile = float64(stat.Population) / stat.SquareMiles
	da.DB.Model(&stat).UpdateColumns(SyntheticCountyStat{
		Population:              stat.Population,
		PopulationMale:          stat.PopulationMale,
		PopulationFemale:        stat.PopulationFemale,
		PopulationPerSquareMile: stat.PopulationPerSquareMile,
	})
}

// PgSyntheticSubdivisionStatDataAccess implements the SubdivisionDataAccess interface using a Postgres database connection.
type PgSyntheticSubdivisionStatDataAccess struct {
	DB *gorm.DB
}

func (da *PgSyntheticSubdivisionStatDataAccess) GetStatBySubdivision(cousub Subdivision) SyntheticSubdivisionStat {
	var stat SyntheticSubdivisionStat
	da.DB.Where(&SyntheticSubdivisionStat{CountyFp: cousub.CountyFp, CousubFp: cousub.CousubFp}).First(&stat)
	return stat
}

func (da *PgSyntheticSubdivisionStatDataAccess) AddMaleToStat(stat SyntheticSubdivisionStat) {
	da.modifyPopulationCount(stat, 1, 0)
}

func (da *PgSyntheticSubdivisionStatDataAccess) AddFemaleToStat(stat SyntheticSubdivisionStat) {
	da.modifyPopulationCount(stat, 0, 1)
}

func (da *PgSyntheticSubdivisionStatDataAccess) RemoveMaleFromStat(stat SyntheticSubdivisionStat) {
	da.modifyPopulationCount(stat, -1, 0)
}

func (da *PgSyntheticSubdivisionStatDataAccess) RemoveFemaleFromStat(stat SyntheticSubdivisionStat) {
	da.modifyPopulationCount(stat, 0, -1)
}

func (da *PgSyntheticSubdivisionStatDataAccess) modifyPopulationCount(stat SyntheticSubdivisionStat, maleDelta, femaleDelta int64) {
	stat.Population += (maleDelta + femaleDelta)
	stat.PopulationMale += maleDelta
	stat.PopulationFemale += femaleDelta
	stat.PopulationPerSquareMile = float64(stat.Population) / stat.SquareMiles
	da.DB.Model(&stat).UpdateColumns(SyntheticSubdivisionStat{
		Population:              stat.Population,
		PopulationMale:          stat.PopulationMale,
		PopulationFemale:        stat.PopulationFemale,
		PopulationPerSquareMile: stat.PopulationPerSquareMile,
	})
}

// PgSyntheticCountyFactDataAccess implements the SyntheticCountyFactDataAccess interface using a Postgres data connection.
type PgSyntheticCountyFactDataAccess struct {
	DB *gorm.DB
}

// GetFactByCountyAndCondition returns a specific fact for a given county and condition.
func (da *PgSyntheticCountyFactDataAccess) GetFactByCountyAndCondition(county County, conditionName string) SyntheticCountyFact {
	var disease SyntheticDisease
	var fact SyntheticCountyFact
	da.DB.Where(&SyntheticDisease{ConditionName: conditionName}).First(&disease)
	da.DB.Where(&SyntheticCountyFact{CountyIdFp: county.CountyIdFp, DiseaseId: disease.Id}).First(&fact)
	return fact
}

// AddMaleToFact increments the male and total population counts for the given SyntheticCountyFact (county and disease)
func (da *PgSyntheticCountyFactDataAccess) AddMaleToFact(fact SyntheticCountyFact) {
	fact.Population += 1
	fact.PopulationMale += 1
	da.DB.Model(&fact).UpdateColumns(SyntheticCountyFact{Population: fact.Population, PopulationMale: fact.PopulationMale})
}

// AddFemaleToFact increments the female and total population counts for the given SyntheticCountyFact (county and disease)
func (da *PgSyntheticCountyFactDataAccess) AddFemaleToFact(fact SyntheticCountyFact) {
	fact.Population += 1
	fact.PopulationFemale += 1
	da.DB.Model(&fact).UpdateColumns(SyntheticCountyFact{Population: fact.Population, PopulationFemale: fact.PopulationFemale})
}

// RemoveMaleFromFact decrements the male and total population counts for the given SyntheticCountyFact (county and disease)
func (da *PgSyntheticCountyFactDataAccess) RemoveMaleFromFact(fact SyntheticCountyFact) {
	fact.Population -= 1
	fact.PopulationMale -= 1
	da.DB.Model(&fact).UpdateColumns(SyntheticCountyFact{Population: fact.Population, PopulationMale: fact.PopulationMale})
}

// RemoveFemaleFromFact decrements the female and total population counts for the given SyntheticCountyFact (county and disease)
func (da *PgSyntheticCountyFactDataAccess) RemoveFemaleFromFact(fact SyntheticCountyFact) {
	fact.Population -= 1
	fact.PopulationFemale -= 1
	da.DB.Model(&fact).UpdateColumns(SyntheticCountyFact{Population: fact.Population, PopulationFemale: fact.PopulationFemale})
}

// PgSyntheticSubdivisionFactDataAccess implements the SyntheticSubdivisionFactDataAccess interface using a Postgres data connection.
type PgSyntheticSubdivisionFactDataAccess struct {
	DB *gorm.DB
}

// GetFactBySubdivisionAndCondition returns a specific fact for a given subdivision and condition.
func (da *PgSyntheticSubdivisionFactDataAccess) GetFactBySubdivisionAndCondition(cousub Subdivision, conditionName string) SyntheticSubdivisionFact {
	var disease SyntheticDisease
	var fact SyntheticSubdivisionFact
	da.DB.Where(&SyntheticDisease{ConditionName: conditionName}).First(&disease)
	da.DB.Where(&SyntheticSubdivisionFact{CosbidFp: cousub.CosbidFp, DiseaseId: disease.Id}).First(&fact)
	return fact
}

// AddMaleToFact increments the male and total population counts for the given SyntheticSubdivisionFact (subdivision and disease)
func (da *PgSyntheticSubdivisionFactDataAccess) AddMaleToFact(fact SyntheticSubdivisionFact) {
	fact.Population += 1
	fact.PopulationMale += 1
	da.DB.Model(&fact).UpdateColumns(SyntheticSubdivisionFact{Population: fact.Population, PopulationMale: fact.PopulationMale})

}

// AddFemaleToFact increments the female and total population counts for the given SyntheticSubdivisionFact (subdivision and disease)
func (da *PgSyntheticSubdivisionFactDataAccess) AddFemaleToFact(fact SyntheticSubdivisionFact) {
	fact.Population += 1
	fact.PopulationFemale += 1
	da.DB.Model(&fact).UpdateColumns(SyntheticSubdivisionFact{Population: fact.Population, PopulationFemale: fact.PopulationFemale})
}

// RemoveMaleFromFact decrements the male and total population counts for the given SyntheticSubdivisionFact (subdivision and disease)
func (da *PgSyntheticSubdivisionFactDataAccess) RemoveMaleFromFact(fact SyntheticSubdivisionFact) {
	fact.Population -= 1
	fact.PopulationMale -= 1
	da.DB.Model(&fact).UpdateColumns(SyntheticSubdivisionFact{Population: fact.Population, PopulationMale: fact.PopulationMale})
}

// RemoveFemaleFromFact decrements the female and total population counts for the given SyntheticSubdivisionFact (subdivision and disease)
func (da *PgSyntheticSubdivisionFactDataAccess) RemoveFemaleFromFact(fact SyntheticSubdivisionFact) {
	fact.Population -= 1
	fact.PopulationFemale -= 1
	da.DB.Model(&fact).UpdateColumns(SyntheticSubdivisionFact{Population: fact.Population, PopulationFemale: fact.PopulationFemale})
}

// getDiseaseIds returns a slice of disease IDs from a slice of diseases
func getDiseaseIds(diseases []SyntheticDisease) []string {
	diseaseIds := make([]string, len(diseases))
	for i, disease := range diseases {
		diseaseIds[i] = disease.Id
	}
	return diseaseIds
}

// StatsDataAccess is the top-level interface for interacting with the Postgres database.
type StatsDataAccess struct {
	Counties                  CountyDataAccess
	Subdivisions              SubdivisionDataAccess
	SyntheticDiseases         SyntheticDiseaseDataAccess
	SyntheticCountyStats      SyntheticCountyStatDataAccess
	SyntheticSubdivisionStats SyntheticSubdivisionStatDataAccess
	SyntheticCountyFacts      SyntheticCountyFactDataAccess
	SyntheticSubdivisionFacts SyntheticSubdivisionFactDataAccess
}

// AddMale increments the male and total population counts for the patient's county and subdivision
func (da *StatsDataAccess) AddMalePatient(patient *models.Patient) {
	county, cousub := da.getCountyAndSubdivisionForPatient(patient)
	countyStat := da.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat := da.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	da.SyntheticCountyStats.AddMaleToStat(countyStat)
	da.SyntheticSubdivisionStats.AddMaleToStat(cousubStat)
}

// AddFemale increments the female and total population counts for the patient's county and subdivision
func (da *StatsDataAccess) AddFemalePatient(patient *models.Patient) {
	county, cousub := da.getCountyAndSubdivisionForPatient(patient)
	countyStat := da.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat := da.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	da.SyntheticCountyStats.AddFemaleToStat(countyStat)
	da.SyntheticSubdivisionStats.AddFemaleToStat(cousubStat)
}

// RemoveMale decrements the male and total population counts for the patient's county and subdivision
func (da *StatsDataAccess) RemoveMalePatient(patient *models.Patient) {
	county, cousub := da.getCountyAndSubdivisionForPatient(patient)
	countyStat := da.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat := da.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	da.SyntheticCountyStats.RemoveMaleFromStat(countyStat)
	da.SyntheticSubdivisionStats.RemoveMaleFromStat(cousubStat)
}

// RemoveFemale decrements the female and total population counts for the patient's county and subdivision
func (da *StatsDataAccess) RemoveFemalePatient(patient *models.Patient) {
	county, cousub := da.getCountyAndSubdivisionForPatient(patient)
	countyStat := da.SyntheticCountyStats.GetStatByCounty(county)
	cousubStat := da.SyntheticSubdivisionStats.GetStatBySubdivision(cousub)
	da.SyntheticCountyStats.RemoveFemaleFromStat(countyStat)
	da.SyntheticSubdivisionStats.RemoveFemaleFromStat(cousubStat)
}

// getCountyAndSubdivisionForPatient returns the county and subdivision objects given the patient's address
func (da *StatsDataAccess) getCountyAndSubdivisionForPatient(patient *models.Patient) (county County, cousub Subdivision) {
	cousub = da.Subdivisions.GetSubdivisionByName(patient.Address[0].City)
	county = da.Counties.GetCountyById(cousub.CountyFp)
	return county, cousub
}

// NewPgStatsDataAccess returns a new StatsDataAccess object with each *DataAccess interface initialized.
func NewPgStatsDataAccess(db *gorm.DB) *StatsDataAccess {
	return &StatsDataAccess{
		Counties:                  &PgCountyDataAccess{DB: db},
		Subdivisions:              &PgSubdivisionDataAccess{DB: db},
		SyntheticDiseases:         &PgSyntheticDiseaseDataAccess{DB: db},
		SyntheticCountyStats:      &PgSyntheticCountyStatDataAccess{DB: db},
		SyntheticSubdivisionStats: &PgSyntheticSubdivisionStatDataAccess{DB: db},
		SyntheticCountyFacts:      &PgSyntheticCountyFactDataAccess{DB: db},
		SyntheticSubdivisionFacts: &PgSyntheticSubdivisionFactDataAccess{DB: db},
	}
}
