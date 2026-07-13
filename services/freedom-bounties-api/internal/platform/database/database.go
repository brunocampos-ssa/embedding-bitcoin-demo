package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS bounties(id TEXT PRIMARY KEY,title TEXT NOT NULL,description TEXT NOT NULL,format TEXT NOT NULL,language TEXT NOT NULL,reward_sats INTEGER NOT NULL CHECK(reward_sats>0),state TEXT NOT NULL,created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS assignments(id TEXT PRIMARY KEY,bounty_id TEXT NOT NULL UNIQUE,actor TEXT NOT NULL,created_at TEXT NOT NULL,FOREIGN KEY(bounty_id) REFERENCES bounties(id));
CREATE TABLE IF NOT EXISTS submissions(id TEXT PRIMARY KEY,bounty_id TEXT NOT NULL,actor TEXT NOT NULL,evidence_url TEXT NOT NULL,notes TEXT NOT NULL,state TEXT NOT NULL,created_at TEXT NOT NULL,approved_at TEXT,FOREIGN KEY(bounty_id) REFERENCES bounties(id));
CREATE TABLE IF NOT EXISTS payouts(id TEXT PRIMARY KEY,submission_id TEXT NOT NULL,asset TEXT NOT NULL,rail TEXT NOT NULL,amount_base_units INTEGER NOT NULL,fee_base_units INTEGER NOT NULL,destination_type TEXT NOT NULL,destination_masked TEXT NOT NULL,destination_raw TEXT NOT NULL,provider_preparation_id TEXT NOT NULL,provider_payment_id TEXT,state TEXT NOT NULL,idempotency_key TEXT UNIQUE,prepared_at TEXT NOT NULL,expires_at TEXT NOT NULL,updated_at TEXT NOT NULL,failure_code TEXT,FOREIGN KEY(submission_id) REFERENCES submissions(id));
CREATE UNIQUE INDEX IF NOT EXISTS one_successful_payout_per_submission ON payouts(submission_id) WHERE state='SUCCEEDED';
CREATE TABLE IF NOT EXISTS audit_records(id INTEGER PRIMARY KEY AUTOINCREMENT,request_id TEXT NOT NULL,actor TEXT NOT NULL,action TEXT NOT NULL,entity_type TEXT NOT NULL,entity_id TEXT NOT NULL,outcome TEXT NOT NULL,metadata TEXT NOT NULL DEFAULT '{}',created_at TEXT NOT NULL);
`

func Open(ctx context.Context, path string) (*sql.DB, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err = db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}
	return db, nil
}
func Seed(ctx context.Context, db *sql.DB) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO bounties(id,title,description,format,language,reward_sats,state,created_at) VALUES
('bounty-finance','Deliver an introductory workshop on personal finance and macroeconomics','Help women starting their careers understand inflation, purchasing power, interest rates, emergency reserves, investment risks, and everyday macroeconomics.','60-minute online workshop; evidence may be a meeting, recording, slide, or organizer link','Portuguese',100,'SUBMITTED',?),
('bounty-interview','Mentor someone preparing for a first software-engineering interview','Provide a practical mentoring session and preparation plan.','45-minute online session','Portuguese',75,'OPEN',?),
('bounty-security','Deliver a basic cybersecurity workshop for women-led small businesses','Teach practical account, device, and phishing safety.','60-minute online workshop','Portuguese',100,'OPEN',?)`, now, now, now)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT OR IGNORE INTO assignments(id,bounty_id,actor,created_at) VALUES('assignment-finance','bounty-finance','demo-recipient',?)`, now)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT OR IGNORE INTO submissions(id,bounty_id,actor,evidence_url,notes,state,created_at) VALUES('submission-finance','bounty-finance','demo-recipient','https://example.com/demo-workshop-evidence','Organizer confirmed the workshop and shared demonstration-only evidence.','SUBMITTED',?)`, now)
	return err
}
