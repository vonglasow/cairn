#!/usr/bin/env bash
set -euo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
STORE=${1:-/tmp/cairn-essai}

rm -rf "$STORE"
mkdir -p "$STORE"

export CAIRN_STORE="$STORE/cairn.db"
export CAIRN_NOTES="$STORE/notes"

cairn() { go run "$HERE/cmd/cairn" "$@" >/dev/null; }

cairn objective add "Livrer cairn 1.0" --state active --priority high --deadline 2026-12-15 --label produit
cairn objective add "Fiabiliser le store" --state active --priority normal --label technique
cairn objective add "Refonte du board" --state on-hold --priority low
cairn objective add "Migration Postgres" --state reached --priority normal
cairn objective add "Portage mobile" --state dropped --priority low

cairn task add "Écrire la spec du packaging" --state current --priority high --deadline 2026-10-04 --estimate 8h --spent 3h --label backend --label sprint-172 --objective 1 --desc "Trois ADR à relire avant de commencer."
cairn task add "Relire l'ADR 0014" --parent 1 --state todo --estimate 2h --spent 1h
cairn task add "Trancher le multi-chart" --parent 1 --state done --estimate 1h --spent 4h
cairn task add "Rédiger la section Flux" --parent 1 --state todo

cairn task add "Reprendre les migrations" --state current --priority critical --deadline 2026-09-28 --estimate 2h --spent 9h --label store --objective 2
cairn task add "Sauvegarde avant rollback" --parent 5 --state done
cairn task add "Tester le rollback à vide" --parent 5 --state todo

cairn task add "Réponse du client sur les quotas" --state blocked --because "en attente du client" --priority high --label produit --objective 1
cairn task add "Revue de la PR 31745" --state watching --blocker 1 --priority normal --label backend
cairn task add "Décision sur le nommage" --state paused --because "arbitrage équipe" --priority low

cairn task add "Nettoyer les lockfiles" --state todo --priority low --label technique --objective 2
cairn task add "Mettre à jour le README" --state todo --deadline 2026-10-06 --label doc
cairn task add "Chiffrer les sauvegardes" --state backlog --priority normal --objective 2
cairn task add "Dessiner les colonnes" --state backlog --priority low --objective 3
cairn task add "Benchmark du store" --state backlog --label technique

cairn task add "Basculer sur Postgres 16" --state done --priority high --label store --objective 4
cairn task add "Supprimer l'ancien schéma" --state done --label store --objective 4
cairn task add "Prototype iOS" --state done --priority low --objective 5

cairn task archive 17
cairn task archive 18

cairn task focus 1
cairn task focus 5

go run "$HERE/cmd/cairn" gui
