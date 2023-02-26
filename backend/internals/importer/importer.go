package importer

import (
	"gitlab.com/robrohan/knotset/internals/repository"
)

type Importer interface {
	Import(file string, repo repository.ResearcherRepository) error
}

// Check to see if this activity has already been imported. Warning: this is based on on the file name on the device
func HasImportedAHighlight(name string, repo *repository.ResearcherRepository, researcherId string) (bool, error) {
	// Check if this file has already been imported
	have, err := repo.HasImported(name, researcherId)
	if err != nil {
		return false, err
	}

	return have, nil
}
