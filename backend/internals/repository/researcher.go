package repository

import (
	"log"

	"github.com/jmoiron/sqlx"
)

type ResearcherRepository struct {
	Db                    *sqlx.DB
	upsertResearcherQuery *sqlx.Stmt
	hasImportedQuery      *sqlx.Stmt
}

func prepareQuery(query string, db *sqlx.DB) *sqlx.Stmt {
	stmt, err := db.Preparex(query)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

// Attach creates a new repository and sets up needed bits
func Attach(schema string, db *sqlx.DB, driver string) *ResearcherRepository {
	a := ResearcherRepository{
		Db: db,
	}

	a.upsertResearcherQuery = prepareQuery(`
		INSERT INTO researcher (uuid, name, email)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid)
		DO NOTHING
	`, db)

	a.hasImportedQuery = prepareQuery(`
		SELECT date_created FROM highlight
		WHERE uuid = $1 AND researcher_uuid = $2
	`, db)

	return &a
}

func (r *ResearcherRepository) Begin() (*sqlx.Tx, error) {
	return r.Db.Beginx()
}

func (r *ResearcherRepository) UpsertResearcher(researcherId string, name string, email string) error {
	_, err := r.upsertResearcherQuery.Exec(researcherId, name, email)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResearcherRepository) HasImported(highlightId string, researcherId string) (bool, error) {
	res, err := r.hasImportedQuery.Query(highlightId, researcherId)
	defer res.Close()

	if err != nil {
		return false, err
	}
	exists := res.Next()
	return exists, nil
}
