package idgen

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/ksuid"
)

type ID[T Resource] struct {
	uid ksuid.KSUID
}

func New[T Resource]() ID[T] {
	return ID[T]{ksuid.New()}
}

func FromString[T Resource](id string) (ID[T], error) {
	prefix := normalizedPrefix[T]()

	if !strings.HasPrefix(id, prefix) {
		var res T
		return ID[T]{ksuid.Nil}, fmt.Errorf("%T must have prefix %q", res, prefix)
	}

	uid, err := ksuid.Parse(strings.TrimPrefix(id, prefix))
	if err != nil {
		return ID[T]{ksuid.Nil}, fmt.Errorf("parse uid value: %w", err)
	}

	return ID[T]{uid}, nil
}

func (id ID[T]) String() string {
	return normalizedPrefix[T]() + id.uid.String()
}

func normalizedPrefix[T Resource]() string {
	var res T

	prefix := strings.TrimSuffix(res.IDPrefix(), "_")
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}

	return strings.ToLower(prefix) + "_"
}

var _ pgtype.TextValuer = (*ID[resource])(nil)

func (id ID[T]) TextValue() (pgtype.Text, error) {
	return pgtype.Text{
		String: id.uid.String(),
		Valid:  true,
	}, nil
}

var _ pgtype.TextScanner = (*ID[resource])(nil)

func (id *ID[T]) ScanText(v pgtype.Text) error {
	return id.uid.Scan(v.String) //nolint:wrapcheck
}
