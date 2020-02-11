package ncdtransport

type PgSQLTransport struct{}

func NewPgSQLTransport() *PgSQLTransport {
	return new(PgSQLTransport)
}
