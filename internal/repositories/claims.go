package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/labstack/gommon/log"
	"github.com/lib/pq"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const duplicateViolationErrorCode = "23505"

// ErrClaimDuplication claim duplication error
var ErrClaimDuplication = errors.New("claim duplication error")

type claims struct {
	conn db.Querier
}

func NewClaims(conn db.Querier) ports.ClaimsRepository {
	return &claims{
		conn: conn,
	}
}

func (c *claims) Save(ctx context.Context, conn db.Querier, claim *domain.Claim) (uuid.UUID, error) {
	var err error
	id := claim.ID

	if claim.MTPProof.Status == pgtype.Undefined {
		claim.MTPProof.Status = pgtype.Null
	}
	if claim.Data.Status == pgtype.Undefined {
		claim.Data.Status = pgtype.Null
	}
	if claim.SignatureProof.Status == pgtype.Undefined {
		claim.SignatureProof.Status = pgtype.Null
	}
	if claim.CredentialStatus.Status == pgtype.Undefined {
		claim.CredentialStatus.Status = pgtype.Null
	}

	if id == uuid.Nil {
		s := `INSERT INTO claims (identifier,
                    other_identifier,
                    expiration,
                    updatable,
                    version,
					rev_nonce,
                    signature_proof,
                    issuer,
                    mtp_proof,
                    data,
                    identity_state,
                    schema_hash,
                    schema_url,
                    schema_type,
          			credential_status,
					revoked,
                    core_claim,
                    index_hash)
		VALUES ($1,  $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id`

		err = conn.QueryRow(ctx, s,
			claim.Identifier,
			claim.OtherIdentifier,
			claim.Expiration,
			claim.Updatable,
			claim.Version,
			claim.RevNonce,
			claim.SignatureProof,
			claim.Issuer,
			claim.MTPProof,
			claim.Data,
			claim.IdentityState,
			claim.SchemaHash,
			claim.SchemaURL,
			claim.SchemaType,
			claim.CredentialStatus,
			claim.Revoked,
			claim.CoreClaim,
			claim.HIndex).Scan(&id)
	} else {
		s := `INSERT INTO claims (
					id,
                    identifier,
                    other_identifier,
                    expiration,
                    updatable,
                    version,
					rev_nonce,
                    signature_proof,
                    issuer,
                    mtp_proof,
                    data,
                    identity_state,
					schema_hash,
                    schema_url,
                    schema_type,
                    credential_status,
                    revoked,
                    core_claim,
                    index_hash
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		ON CONFLICT ON CONSTRAINT claims_pkey 
		DO UPDATE SET 
			( expiration, updatable, version, rev_nonce, signature_proof, mtp_proof, data, identity_state, 
			other_identifier, schema_hash, schema_url, schema_type, issuer, credential_status, revoked, core_claim)
			= (EXCLUDED.expiration, EXCLUDED.updatable, EXCLUDED.version, EXCLUDED.rev_nonce, EXCLUDED.signature_proof,
		EXCLUDED.mtp_proof, EXCLUDED.data, EXCLUDED.identity_state, EXCLUDED.other_identifier, EXCLUDED.schema_hash, 
		EXCLUDED.schema_url, EXCLUDED.schema_type, EXCLUDED.issuer, EXCLUDED.credential_status, EXCLUDED.revoked, EXCLUDED.core_claim)
			RETURNING id`
		err = conn.QueryRow(ctx, s,
			claim.ID,
			claim.Identifier,
			claim.OtherIdentifier,
			claim.Expiration,
			claim.Updatable,
			claim.Version,
			claim.RevNonce,
			claim.SignatureProof,
			claim.Issuer,
			claim.MTPProof,
			claim.Data,
			claim.IdentityState,
			claim.SchemaHash,
			claim.SchemaURL,
			claim.SchemaType,
			claim.CredentialStatus,
			claim.Revoked,
			claim.CoreClaim,
			claim.HIndex).Scan(&id)
	}

	if err == nil {
		return id, nil
	}

	pqErr, ok := err.(*pq.Error)
	if ok {
		if pqErr.Code == duplicateViolationErrorCode {
			return uuid.Nil, ErrClaimDuplication
		}
	}

	log.Errorf("error saving the claim: %v", err.Error())
	return uuid.Nil, fmt.Errorf("error saving the claim: %v", err)
}