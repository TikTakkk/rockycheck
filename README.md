# rockycheck
Outil d’audit (Reverse IP → Domain / Domain → Subdomains). Usage légal seulement.

# BUY API REVERSE
**https://t.me/tomaMA212** 

# ROCKYCHECK

**ROCKYCHECK** — Outil d'audit local pour effectuer des recherches *Reverse IP → Domain* et *Domain → Subdomains*.  
Conçu pour la cartographie réseau, la recherche sécurité et les audits autorisés.

---

> ⚠️ **Usage légal uniquement**  
> N'effectuez des tests, scans ou énumérations **que** sur des systèmes pour lesquels vous avez une **autorisation explicite** (propriété, contrat, ou permission écrite du propriétaire). L'auteur décline toute responsabilité en cas d'utilisation abusive.

---

## 🔎 Description
ROCKYCHECK est un utilitaire léger en **Go** permettant :
- d'interroger des APIs de reverse-IP (récupérer les domaines liés à une IP),
- d'interroger des APIs de reverse-domain (récupérer les sous-domaines d'un domaine),
- d'écrire les résultats dans des fichiers structurés et d'enregistrer les réponses brutes en cas d'erreurs (dossier `debug_responses/`).

Le code contient des **placeholders** pour les endpoints API — AUCUNE clé active ne doit être poussée sur le dépôt public. L'utilisateur configure ses propres endpoints/clé API localement.

---

## ✅ Fonctionnalités principales
- Reverse IP → domaines (via API configurable)
- Reverse domaine → sous-domaines (via API configurable + fallback regex)
- Multi-threading (paramétrable)
- Affichage dynamique (IPs/domains traités, CPM, progress)
- Déduplication automatique des résultats
- Sauvegarde des réponses brutes (debug) si parsing JSON impossible
- Protection : le programme refuse de s'exécuter sans confirmation d'autorisation (flag `--auth` ou variable d'environnement `I_AM_AUTHORIZED=true`)

---

## 📁 Structure du projet

```
ROCKYCHECK/
├─ main.go # Code principal
├─ README.md # Documentation (vous êtes ici)
├─ LICENSE # Licence (MIT)
├─ .gitignore # Fichiers ignorés
├─ debug_responses/ # Réponses brutes sauvegardées (créé à l'exécution)
└─ outputs/
├─ iptodomains.txt
└─ domainstosubdomains.txt

---


---

## 🔧 Prérequis
- Go (1.18+) installé et sur le PATH
- Connexion Internet (si tu utilises des APIs externes)
- Autorisation explicite pour auditer les cibles

```
---

## ⚙️ Configuration des APIs
Dans `main.go` tu trouveras deux constantes (exemples) :

```go
const (
    reverseIPAPI     = "https://exemple.com/?api_key=SOME_KEY&ip={ip}&limit=5000"
    reverseDomainAPI = "https://exemple.com/?api_key=PUBLIC_LICENSE&domain={domain}"
)
```

---

▶️ Comment lancer

Important : Le programme vérifie si tu confirmes que tu es autorisé à l'utiliser. Tu dois soit lancer avec --auth soit exporter la variable d'environnement I_AM_AUTHORIZED=true.

---

Sous Windows (PowerShell)

go run main.go --auth
