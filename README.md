# Docker Manager API

API REST en Go pour décrire et déployer des infrastructures multi-conteneurs via le Docker Engine.

## Prérequis

- Go 1.25+
- Docker

## Variables d'environnement

Copier `.env.example` en `.env` et renseigner les valeurs :

```bash
cp .env.example .env
```

| Variable | Défaut | Description |
|---|---|---|
| `API_KEY` | — | **Obligatoire.** Clé API pour authentifier les requêtes (`X-API-Key`) |
| `PORT` | `3000` | Port d'écoute de l'API |
| `DATABASE_URL` | `app.db` | Chemin vers la base SQLite |

## Développement local

Lancer l'API avec hot reload via Docker Compose :

```bash
docker compose -f docker-compose.dev.yml up --build
```

Ou directement avec Go (nécessite Air) :

```bash
go run . 
```

L'API est accessible sur `http://localhost:3000`.  
La documentation Swagger est disponible sur `http://localhost:3000/docs`.

## Production

```bash
docker compose up --build -d
```

L'image est construite depuis le `Dockerfile` (build multi-stage Go → Alpine).  
La base de données SQLite est persistée dans un volume Docker (`db_data`).

## Authentification

Toutes les requêtes sur `/api/*` nécessitent le header :

```
X-API-Key: <valeur de API_KEY>
```

Les requêtes sans clé valide retournent `401 Unauthorized`.

## Collection Bruno

Les requêtes de test sont dans le dossier `bruno/`. Configurer l'environnement `local` avec `baseUrl` et `API_KEY` avant utilisation.
