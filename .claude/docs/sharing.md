# Sharing tasks with other people

Read this before proposing a way for several people to share one board:
storage that no single machine owns, a synchronisation, a ledger. Two studies
were made — a decentralised store (IPFS and its kind), and a blockchain — and
both end in a recommendation. Neither is built.

## What sharing would have to keep

What cairn is, and what any sharing must not break:

- **One binary, no C toolchain on Windows.** The store is SQLite through
  `modernc.org/sqlite`, pure Go; cgo exists only for the window on macOS and
  Linux.
- **A handful of colleagues**, inside one company, who trust each other. The
  data is work: confidential, and under the GDPR — a task must be deletable,
  for real.
- **The trace is already a log.** Every act on a task is written down, in
  order ([trace.md](trace.md)). It is the natural thing to share: acts can be
  merged where a state cannot.

## Study 1 — a decentralised store

**IPFS, with IPNS for what changes.** Content is public by default: whoever
holds an address fetches it, and a private swarm limits who joins without
encrypting anything. Mutable data goes through IPNS records resolved over the
DHT, slow and unreliable for a list that changes every few minutes. Deleting
is not possible on a network — every node that fetched a block keeps it — so
the only honest erasure is encrypting every block and destroying the key. The
Go side (kubo, boxo) embeds, at the cost of a libp2p stack, a DHT and pinning
to run for five people. **Rejected**: all of the work, and encryption, keys and
availability still ours.

**A CRDT under SQLite.** Nothing fits today. Automerge's Go binding wraps the
Rust library through cgo; cr-sqlite and sqlite-sync are C extensions, which
`modernc.org/sqlite` cannot load, the second under the Elastic licence with a
backend of its own; libSQL's embedded replicas need its C driver; PowerSync and
Electric need Postgres and a sync service. No mature pure-Go CRDT library for
structured data was found. **Rejected**, for now: each one breaks the
one-binary rule.

**Syncing the store file** (Syncthing, a shared drive). Unsafe: a file copied
mid-write, or with its `-wal` out of step, is a corrupt store, and two people's
changes become two files, never one. **Rejected.**

**Recommended — the trace as the unit of sharing, over a small relay.** If
sharing becomes a need, it is built the way cairn already thinks:

1. **Make the trace replayable.** Today it records what was done, not always
   enough to rebuild the state — the order of a column cannot be read back
   ([trace.md](trace.md#the-act-is-traced-not-its-consequences)). Each act
   gets an id that never collides (the author's, plus a hybrid logical
   clock), and the store becomes what replaying the acts gives.
2. **Merge acts, not states.** The same act received twice is applied once;
   two edits of one field keep the later one by that clock; a deletion wins.
   Hand-written in Go, a few hundred lines, over a log that already exists.
3. **Carry them over a relay**: `cairn serve`, one more command of the same
   binary, holding the log behind TLS and a token, clients pushing and pulling
   what they have not seen. Deleting there deletes for everybody — which is
   what the GDPR asks, and what no peer network can promise. A shared Git
   repository of act files can stand in for the relay, at the cost of a
   deletion that has to rewrite history.

Until somebody needs to share a board, none of it is worth building: the
backups ([gui.md](gui.md)) already hand a whole store over, and the MCP server
lets an agent work in it.

## Study 2 — a blockchain

What one brings is a record nobody owns, that nobody can alter after the fact,
for writers who do not trust each other and have no third party to trust. A
board shared by colleagues of one employer has none of those problems: they
trust each other, and the employer is the third party. The decision trees
published for the question — NIST's (NISTIR 8202), SFIA's "should you use a
blockchain?" — all end at "no" there.

What it would cost is the opposite of what cairn needs. A chain is append-only:
the EDPB's guidelines on blockchains (02/2025) say plainly that erasure and
rectification may be impossible on one, and that personal data should stay off
it — at which point the chain holds nothing worth sharing. A permissioned
chain (Hyperledger Fabric) removes the fees and keeps the weight: orderers,
peers and a certificate authority to run, for a to-do list.

**Recommended: no blockchain.** What it would give — knowing who did what, and
that it was not changed since — is a signature on each act of the trace, if
it is ever asked for.

## Sources

- IPFS and IPNS: <https://discuss.ipfs.tech/t/what-is-ipns-how-does-it-work-and-how-to-deal-with-mutable-content/457>
- cr-sqlite: <https://github.com/vlcn-io/cr-sqlite>; sqlite-sync: <https://github.com/sqliteai/sqlite-sync>
- automerge-go: <https://pkg.go.dev/github.com/automerge/automerge-go>
- Turso's rewrite and its sync: <https://docs.turso.tech/sdk/rust/reference.md>
- Syncthing's own limits: <https://docs.syncthing.net/users/faq.html>
- EDPB guidelines 02/2025 on blockchains: <https://www.mondaq.com/fin-tech/1617044/blockchain-and-gdpr-edpb-guidelines-022025-adopted>
- NIST on when a blockchain is warranted: <https://www.nist.gov/news-events/news/2018/01/nist-report-blockchain-technology-aims-go-beyond-hype>
- SFIA, should you use a blockchain: <https://sfia-online.org/en/sfia-8/themes-for-sfia-8/blockchain/should-you-use-a-blockchain>
