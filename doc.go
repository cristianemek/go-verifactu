// Package verifactu records invoices for VERI*FACTU, the AEAT billing regime,
// and submits them to the tax agency.
//
// The Engine owns the chain: it assigns each record its sequence number, links
// it to the previous one by fingerprint, and keeps track of what is still
// pending. Two calls are enough:
//
//	entry, err := engine.Alta(ctx, tenant, factura)  // record and chain
//	envio, err := engine.Remitir(ctx, tenant)        // send what is pending
//
// Alta is idempotent per invoice, so retrying after a timeout is safe. Remitir
// respects the wait the AEAT asks for between submissions.
//
// A Tenant is one taxpayer plus one software installation, and each has its own
// chain. Where that chain lives is up to the [Store] adapter: store/memory for
// tests, store/ledger for files on disk, store/sqlite for a database. All three
// behave the same, which store/storetest checks.
//
// The [Transport] port sends a batch to the AEAT; the aeat package implements it
// over SOAP with mTLS. The record package holds the AEAT data model: the
// fingerprint, the QR URL, and the validation rules that would get an invoice
// rejected.
//
// This library is not a SIF: it is a piece to build one with. The responsible
// declaration of Article 13 of RD 1007/2023 is on whoever ships the software.
package verifactu
