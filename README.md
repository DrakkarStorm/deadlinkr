# deadlinkr

Cet outil permet de scanner un site web pour identifier les liens cassés (dead links), ce qui est crucial pour maintenir la qualité d'un site web. Les liens morts nuisent à l'expérience utilisateur et peuvent affecter négativement le référencement.

## Fonctionnalités principales

1. Crawling de site web : Parcourir récursivement toutes les pages d'un domaine
2. Vérification des liens : Tester chaque lien pour voir s'il retourne une erreur (404, 500, etc.)
3. Filtrage par type : Vérifier les liens internes, externes, ou les deux
4. Limitation de profondeur : Contrôler la profondeur du crawling
5. Rapport détaillé : Générer un rapport des problèmes trouvés

## Usage
### Structure de commandes possible avec Cobra
```
Copydeadlinkr scan [url] - Scanner un site web complet
deadlinkr check [url] - Vérifier une seule page
deadlinkr report - Afficher les résultats du dernier scan
deadlinkr export --format=csv/json/html - Exporter les résultats
```

### Options et flags

```
--depth=N - Limiter la profondeur de crawling
--concurrency=N - Nombre de requêtes simultanées
--timeout=N - Délai d'attente pour chaque requête
--ignore-external - Ignorer les liens externes
--only-external - Vérifier uniquement les liens externes
--user-agent="string" - Définir un user-agent personnalisé
--include-pattern="regex" - Inclure seulement les URLs correspondant au pattern
--exclude-pattern="regex" - Exclure les URLs correspondant au pattern
```

## Améliorations possibles pour des vidéos de suivi

1. Ajouter un mode "fix" pour corriger automatiquement les liens internes cassés
2. Intégrer une API pour suggérer des alternatives pour les liens cassés
3. Ajouter un serveur web pour visualiser les rapports de façon interactive
4. Implémenter un mode "watch" pour surveiller en continu un site
5. Ajouter un support pour l'authentification (sites protégés par mot de passe)