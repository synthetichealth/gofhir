package stats

import (
	"errors"
	"log"

	"github.com/intervention-engine/fhir/models"
)

// PatientStatsCreateInterceptor intercepts any new patient resources added to the database
// and updates the Synthetic Mass population statistics based on that patient's address.
type PatientStatsCreateInterceptor struct {
	DataAccess *StatsDataAccess
}

// Before is unused
func (s *PatientStatsCreateInterceptor) Before(resource interface{}) {}

// After increments population statistics after a patient resource is created.
func (s *PatientStatsCreateInterceptor) After(resource interface{}) {
	patient, ok := resource.(*models.Patient)

	if ok {
		gender := patient.Gender
		switch gender {
		case "male":
			s.DataAccess.AddMalePatient(patient)
		case "female":
			s.DataAccess.AddFemalePatient(patient)
		default:
			log.Printf("PatientStatsCreateInterceptor: Invalid gender for patient %s\n", patient.Id)
		}
	}
}

// OnError is unused
func (s *PatientStatsCreateInterceptor) OnError(err error, resource interface{}) {}

// PatientStatsUpdateInterceptor intercepts any updated patient resources
// and updates the Synthetic Mass population statistics based on that patient's address.
type PatientStatsUpdateInterceptor struct {
	DataAccess *StatsDataAccess
	// The state of the patient before the database update, for comparison after the database update
	patientBefore *models.Patient
	// Tracks if the interceptor failed to cache the patient model before the update
	cacheError error
}

// Before caches a patient resource before it's updated, for comparison after the update.
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

// After compares the updated patient resource to the cached patient resource (from before the update), then
// updates population statistics based on that patient's address.
func (s *PatientStatsUpdateInterceptor) After(resource interface{}) {
	patientAfter, ok := resource.(*models.Patient)

	if ok && s.cacheError == nil {
		// see if the patient's address (or at least, his/her city) changed, and update stats
		if s.patientBefore.Address[0].City != "" && s.patientBefore.Address[0].City != patientAfter.Address[0].City {

			switch patientAfter.Gender {
			case "male":
				s.DataAccess.RemoveMalePatient(s.patientBefore)
				s.DataAccess.AddMalePatient(patientAfter)
			case "female":
				s.DataAccess.RemoveFemalePatient(s.patientBefore)
				s.DataAccess.AddFemalePatient(patientAfter)
			default:
				log.Printf("PatientStatsUpdateInterceptor: Invalid gender for patient %s\n", patientAfter.Id)
			}
		}
	}
}

// OnError is unused
func (s *PatientStatsUpdateInterceptor) OnError(err error, resource interface{}) {}

// PatientStatsDeleteInterceptor intercepts any deleted patient resources
// and updates the Synthetic Mass population statistics based on that patient's address.
type PatientStatsDeleteInterceptor struct {
	DataAccess *StatsDataAccess
}

// Before is unused
func (s *PatientStatsDeleteInterceptor) Before(resource interface{}) {}

// After decrements population statistics after a patient resource is deleted.
func (s *PatientStatsDeleteInterceptor) After(resource interface{}) {
	patient, ok := resource.(*models.Patient)

	if ok {
		switch patient.Gender {
		case "male":
			s.DataAccess.RemoveMalePatient(patient)
		case "female":
			s.DataAccess.RemoveFemalePatient(patient)
		default:
			log.Printf("PatientStatsDeleteInterceptor: Invalid gender for patient %s\n", patient.Id)
		}
	}
}

// OnError is unused
func (s *PatientStatsDeleteInterceptor) OnError(err error, resource interface{}) {}
