package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/robrohan/rigormining/internals/models"
)

// DataRepository holds prepared statements for every table the app uses.
type DataRepository struct {
	Db *sqlx.DB

	upsertUserQuery     *sqlx.Stmt
	getUserByEmailQuery *sqlx.Stmt
	getUserByIdQuery    *sqlx.Stmt
	getUserByTokenQuery *sqlx.Stmt

	createTokenQuery       *sqlx.Stmt
	getTokensByUserIdQuery *sqlx.Stmt
	deleteTokenQuery       *sqlx.Stmt

	createItemQuery      *sqlx.Stmt
	updateItemQuery       *sqlx.Stmt
	updateItemFileQuery   *sqlx.Stmt
	deleteItemQuery       *sqlx.Stmt
	getItemByIdQuery      *sqlx.Stmt
	getItemByHashQuery      *sqlx.Stmt
	setItemDeliveredQuery   *sqlx.Stmt
	clearItemDeliveredQuery *sqlx.Stmt

	upsertTagQuery      *sqlx.Stmt
	attachTagQuery      *sqlx.Stmt
	detachTagQuery      *sqlx.Stmt
	getTagsForItemQuery *sqlx.Stmt
}

func prepareQuery(query string, db *sqlx.DB) *sqlx.Stmt {
	stmt, err := db.Preparex(query)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

// Attach creates a new repository and sets up prepared statements.
func Attach(schema string, db *sqlx.DB, driver string) *DataRepository {
	r := DataRepository{Db: db}

	r.upsertUserQuery = prepareQuery(`
		INSERT INTO users (uuid, authid, email, picture, salt)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE
			SET picture = $4,
			salt = $5
	`, db)

	r.getUserByEmailQuery = prepareQuery(`
		SELECT uuid, email, username, picture, authid, salt
		FROM users WHERE email = $1
	`, db)

	r.getUserByIdQuery = prepareQuery(`
		SELECT uuid, email, username, picture, authid, salt
		FROM users WHERE uuid = $1
	`, db)

	r.getUserByTokenQuery = prepareQuery(`
		SELECT u.uuid, u.email, u.username, u.picture, u.authid, u.salt
		FROM users u
		JOIN token t ON u.uuid = t.user_uuid
		WHERE t.value = $1
	`, db)

	r.createTokenQuery = prepareQuery(`
		INSERT INTO token (uuid, user_uuid, name, value, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, db)

	r.getTokensByUserIdQuery = prepareQuery(`
		SELECT uuid, user_uuid, name, value, created_at
		FROM token WHERE user_uuid = $1
		ORDER BY created_at DESC
	`, db)

	r.deleteTokenQuery = prepareQuery(`
		DELETE FROM token WHERE uuid = $1 AND user_uuid = $2
	`, db)

	r.createItemQuery = prepareQuery(`
		INSERT INTO library_item (
			uuid, user_uuid, title, authors, doi, isbn, year, source_url,
			file_path, file_type, file_hash, added_date, notes,
			item_type, venue, volume, number, pages, publisher
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19
		)
	`, db)

	// NOTE on placeholder numbering: SQLite treats $N as a NAMED parameter
	// and go-sqlite3 binds Go's positional Exec(args...) values in order of
	// each distinct $N's FIRST APPEARANCE in the SQL text - not by the
	// numeral itself (Postgres, by contrast, binds strictly by the number).
	// So every $N below must appear in the text in the same order its
	// corresponding Exec() argument is passed, or SQLite silently binds
	// the wrong value to the wrong placeholder (the WHERE clause ends up
	// comparing against garbage, matches zero rows, and Exec reports no
	// error at all since "0 rows updated" isn't a Go error - this bit us
	// for real once already, see git history).
	r.updateItemQuery = prepareQuery(`
		UPDATE library_item SET
			title = $1, authors = $2, doi = $3, isbn = $4, year = $5,
			source_url = $6, notes = $7, item_type = $8, venue = $9,
			volume = $10, number = $11, pages = $12, publisher = $13
		WHERE uuid = $14 AND user_uuid = $15
	`, db)

	r.updateItemFileQuery = prepareQuery(`
		UPDATE library_item SET file_path = $1, file_type = $2, file_hash = $3
		WHERE uuid = $4 AND user_uuid = $5
	`, db)

	r.deleteItemQuery = prepareQuery(`
		DELETE FROM library_item WHERE uuid = $1 AND user_uuid = $2
	`, db)

	r.getItemByIdQuery = prepareQuery(`
		SELECT * FROM library_item WHERE uuid = $1
	`, db)

	r.getItemByHashQuery = prepareQuery(`
		SELECT * FROM library_item WHERE file_hash = $1 AND user_uuid = $2
	`, db)

	r.setItemDeliveredQuery = prepareQuery(`
		UPDATE library_item SET delivered_at = $1 WHERE uuid = $2 AND user_uuid = $3
	`, db)

	r.clearItemDeliveredQuery = prepareQuery(`
		UPDATE library_item SET delivered_at = NULL WHERE uuid = $1 AND user_uuid = $2
	`, db)

	r.upsertTagQuery = prepareQuery(`
		INSERT INTO tag (uuid, name) VALUES ($1, $2)
		ON CONFLICT (name) DO NOTHING
	`, db)

	r.attachTagQuery = prepareQuery(`
		INSERT INTO item_tag (item_uuid, tag_uuid) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, db)

	r.getTagsForItemQuery = prepareQuery(`
		SELECT t.uuid, t.name FROM tag t
		JOIN item_tag it ON it.tag_uuid = t.uuid
		WHERE it.item_uuid = $1
	`, db)

	r.detachTagQuery = prepareQuery(`
		DELETE FROM item_tag WHERE item_uuid = $1 AND tag_uuid = $2
	`, db)

	return &r
}

// -------------------------------------------------------------------------
// Users

func (r *DataRepository) UpsertUser(user *models.User, salt string) error {
	_, err := r.upsertUserQuery.Exec(user.UUID, user.AuthId, user.Email, user.Picture, salt)
	return err
}

func (r *DataRepository) GetUser(email string) (*models.User, error) {
	return scanOneUser(r.getUserByEmailQuery.Queryx(email))
}

func (r *DataRepository) GetUserById(id string) (*models.User, error) {
	return scanOneUser(r.getUserByIdQuery.Queryx(id))
}

func (r *DataRepository) GetUserByToken(value string) (*models.User, error) {
	return scanOneUser(r.getUserByTokenQuery.Queryx(value))
}

func scanOneUser(rows *sqlx.Rows, err error) (*models.User, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := models.User{}
	found := false
	for rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			return nil, err
		}
		found = true
	}
	if !found {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// -------------------------------------------------------------------------
// API tokens

func (r *DataRepository) CreateToken(token *models.Token) error {
	_, err := r.createTokenQuery.Exec(token.UUID, token.UserId, token.Name, token.Value, token.CreatedAt)
	return err
}

func (r *DataRepository) GetTokensByUserId(userId string) ([]models.Token, error) {
	rows, err := r.getTokensByUserIdQuery.Queryx(userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]models.Token, 0)
	for rows.Next() {
		t := models.Token{}
		if err := rows.StructScan(&t); err != nil {
			return nil, err
		}
		// Never send the raw token value back out once it's been created.
		t.Value = ""
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (r *DataRepository) DeleteToken(tokenId string, userId string) error {
	_, err := r.deleteTokenQuery.Exec(tokenId, userId)
	return err
}

// -------------------------------------------------------------------------
// Library items

func (r *DataRepository) CreateItem(item *models.LibraryItem) error {
	_, err := r.createItemQuery.Exec(
		item.UUID, item.UserId, item.Title, item.Authors, item.Doi, item.Isbn,
		item.Year, item.SourceUrl, item.FilePath, item.FileType, item.FileHash,
		item.AddedDate, item.Notes,
		item.ItemType, item.Venue, item.Volume, item.Number, item.Pages, item.Publisher,
	)
	return err
}

func (r *DataRepository) UpdateItem(item *models.LibraryItem) error {
	_, err := r.updateItemQuery.Exec(
		item.Title, item.Authors, item.Doi, item.Isbn,
		item.Year, item.SourceUrl, item.Notes,
		item.ItemType, item.Venue, item.Volume, item.Number, item.Pages, item.Publisher,
		item.UUID, item.UserId,
	)
	return err
}

func (r *DataRepository) UpdateItemFile(item *models.LibraryItem) error {
	_, err := r.updateItemFileQuery.Exec(item.FilePath, item.FileType, item.FileHash, item.UUID, item.UserId)
	return err
}

func (r *DataRepository) DeleteItem(itemId string, userId string) error {
	_, err := r.deleteItemQuery.Exec(itemId, userId)
	return err
}

func (r *DataRepository) GetItemById(itemId string) (*models.LibraryItem, error) {
	rows, err := r.getItemByIdQuery.Queryx(itemId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	item := models.LibraryItem{}
	found := false
	for rows.Next() {
		if err := rows.StructScan(&item); err != nil {
			return nil, err
		}
		found = true
	}
	if !found {
		return nil, errors.New("item not found")
	}
	return &item, nil
}

func (r *DataRepository) GetItemByHash(hash string, userId string) (*models.LibraryItem, error) {
	rows, err := r.getItemByHashQuery.Queryx(hash, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	item := models.LibraryItem{}
	if rows.Next() {
		if err := rows.StructScan(&item); err != nil {
			return nil, err
		}
		return &item, nil
	}
	return nil, sql.ErrNoRows
}

func (r *DataRepository) SetItemDelivered(itemId string, userId string, deliveredAt string) error {
	_, err := r.setItemDeliveredQuery.Exec(deliveredAt, itemId, userId)
	return err
}

// ClearItemDelivered resets delivered_at to null - used when the Sync
// screen finds a file it previously copied is no longer actually present
// on the device (deleted by the user directly on the Kobo), so the item
// goes back to "pending" instead of the app trusting a stale flag forever.
func (r *DataRepository) ClearItemDelivered(itemId string, userId string) error {
	_, err := r.clearItemDeliveredQuery.Exec(itemId, userId)
	return err
}

// ItemFilter narrows a library listing. An empty/false-zero field means
// "don't filter on this".
type ItemFilter struct {
	Query     string
	Delivered *bool // nil = any, true = delivered, false = pending
}

// ListItems is intentionally not a prepared statement, since the WHERE
// clause shape depends on which filters are set.
func (r *DataRepository) ListItems(userId string, f ItemFilter) ([]models.LibraryItem, error) {
	clauses := []string{"user_uuid = ?"}
	args := []interface{}{userId}

	if f.Query != "" {
		clauses = append(clauses, "(title LIKE ? OR authors LIKE ?)")
		like := "%" + f.Query + "%"
		args = append(args, like, like)
	}
	if f.Delivered != nil {
		if *f.Delivered {
			clauses = append(clauses, "delivered_at IS NOT NULL")
		} else {
			clauses = append(clauses, "delivered_at IS NULL")
		}
	}

	query := fmt.Sprintf(
		"SELECT * FROM library_item WHERE %s ORDER BY added_date DESC",
		strings.Join(clauses, " AND "),
	)
	query = r.Db.Rebind(query)

	rows, err := r.Db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.LibraryItem, 0)
	for rows.Next() {
		item := models.LibraryItem{}
		if err := rows.StructScan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// -------------------------------------------------------------------------
// Tags

func (r *DataRepository) AttachTag(itemId string, tagName string) error {
	newId := uuid.New().String()
	if _, err := r.upsertTagQuery.Exec(newId, tagName); err != nil {
		return err
	}

	var tagId string
	if err := r.Db.Get(&tagId, r.Db.Rebind("SELECT uuid FROM tag WHERE name = ?"), tagName); err != nil {
		return err
	}

	_, err := r.attachTagQuery.Exec(itemId, tagId)
	return err
}

func (r *DataRepository) DetachTag(itemId string, tagId string) error {
	_, err := r.detachTagQuery.Exec(itemId, tagId)
	return err
}

func (r *DataRepository) GetTagsForItem(itemId string) ([]models.Tag, error) {
	rows, err := r.getTagsForItemQuery.Queryx(itemId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]models.Tag, 0)
	for rows.Next() {
		t := models.Tag{}
		if err := rows.StructScan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}
